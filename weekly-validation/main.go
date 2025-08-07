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
	"github.com/shared"
)

type FailureCountRequest struct {
	BskyHandle   string `json:"bskyHandle"`
	FailureCount int    `json:"failureCount"`
}

type ValidationResult struct {
	BskyHandle   string `json:"bskyHandle"`
	IsValid      bool   `json:"isValid"`
	FailureCount int    `json:"failureCount"`
	Action       string `json:"action"` // "none", "removed"
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

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var request FailureCountRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	store, err := kv.OpenStore("default")
	if err != nil {
		http.Error(w, "Error opening store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer store.Close()

	failureKey := "failure-" + request.BskyHandle

	// Update failure count
	err = store.Set(failureKey, []byte(strconv.Itoa(request.FailureCount)))
	if err != nil {
		http.Error(w, "Error setting failure count: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var result ValidationResult
	result.BskyHandle = request.BskyHandle
	result.FailureCount = request.FailureCount
	result.Action = "none"

	// If failure count is 4, remove the user
	if request.FailureCount >= 4 {
		fmt.Printf("Removing user %s due to 4 consecutive failures\n", request.BskyHandle)

		// Find all keys for this user and remove them
		keys, err := store.GetKeys()
		if err != nil {
			http.Error(w, "Error getting keys: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var keysToRemove []string
		for _, key := range keys {
			if strings.Contains(key, "-") && !strings.HasPrefix(key, "failure-") &&
				key != "endpoint" && key != "accessJwt" && key != "" {
				value, err := store.Get(key)
				if err != nil {
					continue
				}
				if string(value) == request.BskyHandle {
					keysToRemove = append(keysToRemove, key)
				}
			}
		}

		// Remove user from all validation stores
		for _, key := range keysToRemove {
			fmt.Printf("Removing key %s for user %s\n", key, request.BskyHandle)
			err = store.Delete(key)
			if err != nil {
				fmt.Printf("Error deleting key %s: %v\n", key, err)
				continue
			}

			// Remove from Bluesky lists and starter packs, and remove label
			err = removeFromBlueskyAndLabel(key, request.BskyHandle)
			if err != nil {
				fmt.Printf("Error removing from Bluesky for key %s: %v\n", key, err)
			}
		}

		// Remove failure count key
		err = store.Delete(failureKey)
		if err != nil {
			fmt.Printf("Error deleting failure key: %v\n", err)
		}

		result.Action = "removed"
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

	bskyHandle := strings.ToLower(path)

	store, err := kv.OpenStore("default")
	if err != nil {
		http.Error(w, "Error opening store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer store.Close()

	// Get current failure count
	failureKey := "failure-" + bskyHandle
	failureCount := 0
	if exists, _ := store.Exists(failureKey); exists {
		if failureData, err := store.Get(failureKey); err == nil {
			if count, err := strconv.Atoi(string(failureData)); err == nil {
				failureCount = count
			}
		}
	}

	result := ValidationResult{
		BskyHandle:   bskyHandle,
		IsValid:      false,
		FailureCount: failureCount,
		Action:       "none",
	}

	// Find the user's verification entries
	keys, err := store.GetKeys()
	if err != nil {
		http.Error(w, "Error getting keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, key := range keys {
		if strings.Contains(key, "-") && !strings.HasPrefix(key, "failure-") &&
			key != "endpoint" && key != "accessJwt" && key != "" {
			value, err := store.Get(key)
			if err != nil {
				continue
			}
			if string(value) == bskyHandle {
				// Found a verification entry, check if it's still valid
				parts := strings.Split(key, "-")
				if len(parts) >= 2 {
					moduleKey := parts[0]
					verificationId := strings.Join(parts[1:], "-")

					isValid := checkValidation(moduleKey, verificationId, bskyHandle)
					if isValid {
						result.IsValid = true
						break
					}
				}
			}
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
	url := fmt.Sprintf("https://verifiedbsky.net/validate-%s/", moduleKey)

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
