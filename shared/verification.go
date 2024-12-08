package shared

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ModuleSpecifics struct {
	ModuleKey          string
	ModuleName         string
	ModuleNameShortened string
	ModuleLabel		string
	ExplanationText	string
	VerificationFunc  func(verificationId string, bskyHandle string) (bool, error)
	FirstAndSecondLevel map[string][]string
	Level1TranslationMap map[string]string
	Level2TranslationMap map[string]string
}

func (m ModuleSpecifics) Handle(w http.ResponseWriter, r *http.Request) {
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

		naming, err := SetupNamingStructure(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// store add to bsky starter pack
		fmt.Println("Storing verified user and adding to Bluesky starter pack")
		result, err := StoreAndAddToBskyStarterPack(naming, validationRequest.VerificationId, validationRequest.BskyHandle, profile.DID, m.ModuleLabel, accessJwt, endpoint)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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