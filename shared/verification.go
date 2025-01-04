package shared

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fermyon/spin/sdk/go/v2/kv"
	"github.com/fermyon/spin/sdk/go/v2/variables"
)

type ModuleSpecifics struct {
	ModuleKey          string
	ModuleName         string
	ModuleNameShortened string
	ModuleLabel		string
	ExplanationText	string
	VerificationFunc  func(verificationId string, bskyHandle string) (bool, error)
	NamingFunc        func(m ModuleSpecifics, verificationId string) (Naming, error)
	FirstAndSecondLevel map[string][]string
	Level1TranslationMap map[string]string
	Level2TranslationMap map[string]string
}

func (m ModuleSpecifics) Handle(w http.ResponseWriter, r *http.Request) {
	// list of bsky handles that are blacklisted, which means request to verify them will be rejected
	bskyHandleBlacklist := []string{
		"khmarbaise.bsky.social",
	}

	verifyOnly, err := variables.Get("verify_only")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch r.Method {

	case http.MethodGet:
		lastSlash := strings.LastIndex(r.URL.Path, "/")
		if lastSlash > 0 {
			title := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
			if title != "" {
				if (title == "verificationText") {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, m.ExplanationText)
					return
				}
				fmt.Println("Getting Starter Pack and List for " + title)
				accessJwt, endpoint, err := LoginToBsky()
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
				}
				err = RespondWithStarterPacksAndListsForTitle(title, w, accessJwt, endpoint)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}
		
		if (verifyOnly == "true") {
			http.Error(w, "Method not allowed in verify_only mode", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("Getting all Starter Packs and Lists")
		accessJwt, endpoint, err := LoginToBsky()
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
		err = RespondWithAllStarterPacksAndListsForModule(m, w, accessJwt, endpoint)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		// get request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var validationRequest ValidationRequest
		err = json.Unmarshal(body, &validationRequest)
		if err != nil {
			http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
			return
		}
		validationRequest.BskyHandle = strings.ToLower(validationRequest.BskyHandle)

		// check if bsky handle is blacklisted
		for _, blacklistedHandle := range bskyHandleBlacklist {
			if validationRequest.BskyHandle == blacklistedHandle {
				http.Error(w, "The Bluesky handle " + validationRequest.BskyHandle + " has requested to not be verified through this service", http.StatusBadRequest)
				return
			}
		}

		// verify externally
		fmt.Println("Validating with external service")
		verified, err := m.VerificationFunc(validationRequest.VerificationId, validationRequest.BskyHandle)
		if !verified {
			http.Error(w, "Verification failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		// get bsky profile
		fmt.Println("Getting Bluesky profile for handle " + validationRequest.BskyHandle)
		accessJwt, endpoint, err := LoginToBsky()

		profile, err := GetProfile(validationRequest.BskyHandle, accessJwt, endpoint)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if profile == (ProfileResponse{}) {
			http.Error(w, "Error getting profile", http.StatusInternalServerError)
			return
		}

		naming, err := m.NamingFunc(m, validationRequest.VerificationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// store in kv store
		err = Store(naming, validationRequest.VerificationId, validationRequest.BskyHandle)
		if err != nil {
			http.Error(w, "Error storing user in k/v store: " + err.Error(), http.StatusInternalServerError)
			return
		}

		result := []ListOrStarterPackWithUrl{}

		if (verifyOnly == "true") {
			result = append(result, ListOrStarterPackWithUrl{
				Title: naming.Title,
				URL: "",
			})
			for firstLevel, secondLevels := range naming.FirstAndSecondLevel {
				result = append(result, ListOrStarterPackWithUrl{
					Title: firstLevel.Title,
					URL: "",
				})
				for _, secondLevel := range secondLevels {
					result = append(result, ListOrStarterPackWithUrl{
						Title: secondLevel.Title,
						URL: "",
					})
				}
			}
		} else {
			// add to bsky starter pack
			fmt.Println("Adding verified user to Bluesky starter pack")
			result, err = AddToBskyStarterPacksAndList(naming, validationRequest.VerificationId, validationRequest.BskyHandle, profile.DID, m.ModuleLabel, accessJwt, endpoint)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
		

	case http.MethodPut:
		if (verifyOnly == "true") {
			http.Error(w, "Method not allowed in verify_only mode", http.StatusMethodNotAllowed)
			return
		}
		accessJwt, endpoint, err := LoginToBskyWithReq(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		fmt.Println("Setting up all Starter Packs and Lists")
		err = SetupAllStarterPacksAndLists(m, w, accessJwt, endpoint)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func Store(naming Naming, moduleKey string, bskyHandle string) error {
	fmt.Println("Storing verified user in kv store")
	store, err := kv.OpenStore("default")
	if err != nil {
		return err
	}
	defer store.Close()

	key := naming.Key + "-" + moduleKey

	return store.Set(key, []byte(bskyHandle))
}