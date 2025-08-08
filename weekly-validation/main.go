package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/kv"
	"github.com/fermyon/spin/sdk/go/v2/variables"
	"github.com/shared"
)

const (
	// MaxFailureCount is the number of consecutive failures before removing a user from a module.
	// This value determines how many times a validation can fail before the user is automatically
	// removed from that specific module.
	MaxFailureCount = 4
)

type FailureCountRequest struct {
	BskyHandle   string `json:"bskyHandle"`
	ModuleKey    string `json:"moduleKey"`
	FailureCount int    `json:"failureCount"`
}

type ValidationResult struct {
	BskyHandle    string                  `json:"bskyHandle"`
	ModuleResults map[string]ModuleResult `json:"moduleResults"`
	Action        string                  `json:"action"` // "none", "partial_removal", "full_removal"
}

type ModuleResult struct {
	ModuleKey    string `json:"moduleKey"`
	IsValid      bool   `json:"isValid"`
	FailureCount int    `json:"failureCount"`
	Removed      bool   `json:"removed"`
}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Handle failure count updates from GitHub workflow
			handleFailureCountUpdate(w, r)
		case http.MethodGet:
			// Handle validation check for a specific account
			handleValidationCheck(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleFailureCountUpdate(w http.ResponseWriter, r *http.Request) {
	// Authenticate request
	_, _, err := shared.LoginToBskyWithReq(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Ensure request body is closed
	defer r.Body.Close()

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var request FailureCountRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	store, err := kv.OpenStore("failures")
	if err != nil {
		http.Error(w, "Error opening store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer store.Close()

	failureKey := fmt.Sprintf("failure-%s-%s", request.ModuleKey, request.BskyHandle)

	// Update failure count for this specific module
	err = store.Set(failureKey, []byte(strconv.Itoa(request.FailureCount)))
	if err != nil {
		http.Error(w, "Error setting failure count: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result := ValidationResult{
		BskyHandle:    request.BskyHandle,
		ModuleResults: make(map[string]ModuleResult),
		Action:        "none",
	}

	// Add the updated module result
	result.ModuleResults[request.ModuleKey] = ModuleResult{
		ModuleKey:    request.ModuleKey,
		IsValid:      request.FailureCount == 0,
		FailureCount: request.FailureCount,
		Removed:      false,
	}

	// If failure count reaches the maximum threshold, remove the user from this specific module
	if request.FailureCount >= MaxFailureCount {
		fmt.Printf("Removing user %s from module %s due to %d consecutive failures\n", request.BskyHandle, request.ModuleKey, MaxFailureCount)

		// Find and remove the specific key for this module and user
		keys, err := store.GetKeys()
		if err != nil {
			http.Error(w, "Error getting keys: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var keyToRemove string
		for _, key := range keys {
			if strings.HasPrefix(key, request.ModuleKey+"-") && !strings.HasPrefix(key, "failure-") &&
				key != "endpoint" && key != "accessJwt" && key != "" {
				value, err := store.Get(key)
				if err != nil {
					continue
				}
				if string(value) == request.BskyHandle {
					keyToRemove = key
					break
				}
			}
		}

		if keyToRemove != "" {
			fmt.Printf("Removing key %s for user %s from module %s\n", keyToRemove, request.BskyHandle, request.ModuleKey)
			err = store.Delete(keyToRemove)
			if err != nil {
				fmt.Printf("Error deleting key %s: %v\n", keyToRemove, err)
			} else {
				// Remove from Bluesky lists and starter packs, and remove label for this module
				err = removeFromBlueskyAndLabel(keyToRemove, request.BskyHandle)
				if err != nil {
					fmt.Printf("Error removing from Bluesky for key %s: %v\n", keyToRemove, err)
				}

				result.ModuleResults[request.ModuleKey] = ModuleResult{
					ModuleKey:    request.ModuleKey,
					IsValid:      false,
					FailureCount: request.FailureCount,
					Removed:      true,
				}
				result.Action = "partial_removal"
			}
		}

		// Remove failure count key for this module
		err = store.Delete(failureKey)
		if err != nil {
			fmt.Printf("Error deleting failure key: %v\n", err)
		}
	}

	// Return response
	jsonResult, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(jsonResult))
}

func handleValidationCheck(w http.ResponseWriter, r *http.Request) {
	// Authenticate request
	_, _, err := shared.LoginToBskyWithReq(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract bsky handle from URL path
	path := strings.TrimPrefix(r.URL.Path, "/weekly-validation/")
	if path == "" {
		http.Error(w, "Bluesky handle required", http.StatusBadRequest)
		return
	}

	segments := strings.SplitN(path, "/", 2)
	bskyHandle := strings.ToLower(segments[0])

	store, err := kv.OpenStore("default")
	if err != nil {
		http.Error(w, "Error opening store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer store.Close()

	failureStore, err := kv.OpenStore("failures")
	if err != nil {
		http.Error(w, "Error opening store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer failureStore.Close()

	result := ValidationResult{
		BskyHandle:    bskyHandle,
		ModuleResults: make(map[string]ModuleResult),
		Action:        "none",
	}

	// Find all verification entries for this user
	keys, err := store.GetKeys()
	if err != nil {
		http.Error(w, "Error getting keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Track which modules this user has verification entries for
	userModules := make(map[string]string) // moduleKey -> verificationId

	for _, key := range keys {
		if strings.Contains(key, "-") && !strings.HasPrefix(key, "failure-") &&
			key != "endpoint" && key != "accessJwt" && key != "" {
			value, err := store.Get(key)
			if err != nil {
				continue
			}
			if string(value) == bskyHandle {
				parts := strings.Split(key, "-")
				if len(parts) >= 2 {
					moduleKey := parts[0]
					verificationId := strings.Join(parts[1:], "-")
					userModules[moduleKey] = verificationId
				}
			}
		}
	}

	// Check validation status and failure counts for each module
	for moduleKey, verificationId := range userModules {
		// Get current failure count for this module
		failureKey := fmt.Sprintf("failure-%s-%s", moduleKey, bskyHandle)
		failureCount := 0
		if exists, _ := failureStore.Exists(failureKey); exists {
			if failureData, err := failureStore.Get(failureKey); err == nil {
				if count, err := strconv.Atoi(string(failureData)); err == nil {
					failureCount = count
				}
			}
		}

		// Check if validation is still valid
		isValid := checkValidation(moduleKey, verificationId, bskyHandle)

		result.ModuleResults[moduleKey] = ModuleResult{
			ModuleKey:    moduleKey,
			IsValid:      isValid,
			FailureCount: failureCount,
			Removed:      false,
		}
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(jsonResult))
}

func checkValidation(moduleKey, verificationId, bskyHandle string) bool {
	// This would call the appropriate validation endpoint
	// For now, we'll use a simple HTTP client to call the validation endpoint
	// Base URL can be configured via Spin variable to allow localhost testing
	baseURL, err := variables.Get("validation_base_url")
	if err != nil || baseURL == "" {
		baseURL = "https://verifiedbsky.net"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	url := fmt.Sprintf("%s/validate-%s/?verify_only=true", baseURL, moduleKey)

	requestBody := map[string]string{
		"verificationId": verificationId,
		"bskyHandle":     bskyHandle,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Error marshaling request for %s: %v\n", moduleKey, err)
		return false
	}

	resp, err := shared.SendPost(url, string(jsonData), "application/json")
	if err != nil {
		fmt.Printf("Error validating %s with %s: %v\n", bskyHandle, moduleKey, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func removeFromBlueskyAndLabel(key, bskyHandle string) error {
	accessJwt, endpoint, err := shared.LoginToBsky()
	if err != nil {
		return fmt.Errorf("error logging in to Bluesky: %v", err)
	}

	// Get all lists and starter packs
	allLists, err := shared.GetLists(accessJwt, endpoint)
	if err != nil {
		return fmt.Errorf("error getting lists: %v", err)
	}

	allStarterPacks, err := shared.GetStarterPacks(accessJwt, endpoint)
	if err != nil {
		return fmt.Errorf("error getting starter packs: %v", err)
	}

	// Extract module key from the store key
	parts := strings.Split(key, "-")
	if len(parts) < 1 {
		return fmt.Errorf("invalid key format: %s", key)
	}
	moduleKey := parts[0]

	// Remove from all lists and starter packs
	for _, list := range allLists {
		_, err = shared.CheckOrDeleteUserOnList(list.URI, bskyHandle, true, accessJwt, endpoint)
		if err != nil {
			fmt.Printf("Error removing user from list %s: %v\n", list.Name, err)
		}
	}

	for _, starterPack := range allStarterPacks {
		_, err = shared.CheckOrDeleteUserOnList(starterPack.Record.List, bskyHandle, true, accessJwt, endpoint)
		if err != nil {
			fmt.Printf("Error removing user from starter pack %s: %v\n", starterPack.Record.Name, err)
		}
	}

	// Remove the label
	err = shared.RemoveLabel("ms-"+moduleKey, bskyHandle, accessJwt, endpoint)
	if err != nil {
		fmt.Printf("Error removing label ms-%s from %s: %v\n", moduleKey, bskyHandle, err)
	}

	return nil
}

func main() {}
