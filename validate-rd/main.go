package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shared"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

type UserProfile struct {
	TechnologyFocusArea      []string        `json:"technologyFocusArea"`
	AwardCategory            []string        `json:"awardCategory"`
	ID                       int             `json:"id"`
	UserProfileSocialNetwork []SocialNetwork `json:"userProfileSocialNetwork"`
}

type SocialNetwork struct {
	ID                int    `json:"id"`
	UserProfileId     int    `json:"userProfileId"`
	SocialNetworkId   int    `json:"socialNetworkId"`
	Handle            string `json:"handle"`
	SocialNetworkName string `json:"socialNetworkName"`
}

type Response struct {
	UserProfile UserProfile `json:"userProfile"`
}

func init() {

	moduleKey := "rd"
	moduleName := "Microsoft Regional Directors (RDs)"
	moduleNameShortened := "RDs"

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			lastSlash := strings.LastIndex(r.URL.Path, "/")
			if lastSlash > 0 {
				title := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
				if title != "" {
					if (title == "verificationText") {
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusOK)
						fmt.Fprintln(w, "This is your RD ID, a GUID. If you open your profile on <a href=\"https://rd.microsoft.com\" target=\"_blank\">rd.microsoft.com</a>, it is the last part of the URL, after the last /. For this to work, you need to have the link to your Bluesky profile in the list of social networks on your RD profile (use \"Other\" as type).</small></div>")
						return
					}
					fmt.Println("Getting Starter Pack and List for " + title)
					accessJwt, endpoint, err := shared.LoginToBsky()
					if err != nil {
						http.Error(w, err.Error(), http.StatusUnauthorized)
					}
					err = shared.RespondWithStarterPacksAndListsForTitle(title, w, accessJwt, endpoint)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}					
					return
				}
			}
			fmt.Println("Getting all Starter Packs and Lists")
			accessJwt, endpoint, err := shared.LoginToBsky()
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
			err = shared.RespondWithAllStarterPacksAndListsForModule(moduleKey, moduleName, moduleNameShortened, make(map[string][]string), make(map[string]string), make(map[string]string), w, accessJwt, endpoint)
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

			var rdValidationRequest shared.ValidationRequest
			err = json.Unmarshal(body, &rdValidationRequest)
			if err != nil {
				http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}
			rdValidationRequest.BskyHandle = strings.ToLower(rdValidationRequest.BskyHandle)

			// get RD profile
			fmt.Println("Validating RD with ID: " + rdValidationRequest.VerificationId)
			url := fmt.Sprintf("https://mavenapi-prod.azurewebsites.net/api/rd/UserProfiles/public/%s", url.QueryEscape(rdValidationRequest.VerificationId))

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				http.Error(w, "Error fetching the RD profile, probably caused by an invalid RD ID: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			var response Response
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				fmt.Println("Error decoding RD JSON: " + err.Error())
				http.Error(w, "Error decoding RD JSON, probably caused by an invalid RD ID: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// check if bsky handle is in RD profile
			if containsSocialNetworkWithHandle(response.UserProfile.UserProfileSocialNetwork, rdValidationRequest.BskyHandle) {
				fmt.Print("Social network with handle '" + rdValidationRequest.BskyHandle + "' found\n")
			} else {
				fmt.Print("Social network with handle '" + rdValidationRequest.BskyHandle + "' not found\n")
				http.Error(w, fmt.Sprintf("Link to social network with handle %s not found for RD %s", rdValidationRequest.BskyHandle, rdValidationRequest.VerificationId), http.StatusNotFound)
				return
			}

			// get bsky profile
			fmt.Println("Getting Bluesky profile for handle " + rdValidationRequest.BskyHandle)
			accessJwt, endpoint, err := shared.LoginToBsky()

			profile, err := shared.GetProfile(rdValidationRequest.BskyHandle, accessJwt, endpoint)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if profile == (shared.ProfileResponse{}) {
				http.Error(w, "Error getting profile", http.StatusInternalServerError)
				return
			}

			naming, err := shared.SetupNamingStructure(moduleKey, moduleName, moduleNameShortened, make(map[string][]string), make(map[string]string), make(map[string]string))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// store RD and add to bsky starter pack
			fmt.Println("Storing RD and adding to Bluesky starter pack")
			result, err := shared.StoreAndAddToBskyStarterPack(naming, rdValidationRequest.VerificationId, rdValidationRequest.BskyHandle, profile.DID, accessJwt, endpoint)

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
			accessJwt, endpoint, err := shared.LoginToBskyWithReq(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}

			fmt.Println("Setting up all Starter Packs and Lists")
			err = shared.SetupAllStarterPacksAndLists(moduleKey, moduleName, moduleNameShortened, make(map[string][]string), make(map[string]string), make(map[string]string), w, accessJwt, endpoint)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func containsSocialNetworkWithHandle(socialNetworks []SocialNetwork, handle string) bool {
	for _, sn := range socialNetworks {
		if sn.Handle == handle || sn.Handle == "bsky.app/profile/"+handle {
			return true
		}
	}
	return false
}

func main() {}
