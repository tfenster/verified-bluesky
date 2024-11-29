package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

type Person struct {
    ID          int     `json:"id"`
    SessionID   int     `json:"sessionId"`
    Image       string  `json:"image"`
    Name        string  `json:"name"`
    Tagline     string  `json:"tagline"`
    Company     string  `json:"company"`
    IsFeatured  bool    `json:"isFeatured"`
    CountryName *string `json:"countryName"`
    Biography   string  `json:"biography"`
}

type RunEventsResponse struct {
    Data       []Person `json:"data"`
    Successful bool     `json:"successful"`
    Alerts     *string  `json:"alerts"`
}

func init() {

	moduleKey := "dynamicsminds"
	moduleName := "DynamicsMinds speakers"
	moduleNameShortened := "DynamicsMinds"

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
						fmt.Fprintln(w, "This is your name, exactly as it appears on the DynamicsMinds speakers page. For this to work, you need to have the link to your Bluesky profile (https://bsky.app/profile/...) somewhere in your biography.</small></div>")
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

			var dmValidationRequest shared.ValidationRequest
			err = json.Unmarshal(body, &dmValidationRequest)
			if err != nil {
				http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}
			dmValidationRequest.BskyHandle = strings.ToLower(dmValidationRequest.BskyHandle)

			// get DM profile
			fmt.Println("Validating DM speaker with name: " + dmValidationRequest.VerificationId)
			url := "https://api.runevents.net/api/sessions-and-speakers/external-speakers?eventSlug=dynamicsminds-2024"

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				http.Error(w, "Error fetching the DM speaker list: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			var response RunEventsResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				fmt.Println("Error decoding DM Speaker JSON: " + err.Error())
				http.Error(w, "Error decoding DM Speaker JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// check if bsky handle is in DM profile
			found := false
			for _, speaker := range response.Data {
				if speaker.Name == dmValidationRequest.VerificationId {
					fmt.Print("Speaker with name '" + dmValidationRequest.VerificationId + "' found\n")
					if strings.Contains(speaker.Biography, dmValidationRequest.BskyHandle) {
						fmt.Print("Biography contains handle '" + dmValidationRequest.BskyHandle + "'\n")
						found = true
						break
					}
				}
			}
			if !found {
				fmt.Print("Speaker with name '" + dmValidationRequest.VerificationId + "' and handle '" + dmValidationRequest.BskyHandle + "' not found\n")
				http.Error(w, fmt.Sprintf("Link to social network with handle %s not found for DM speaker %s", dmValidationRequest.BskyHandle, dmValidationRequest.VerificationId), http.StatusNotFound)
				return
			}

			// get bsky profile
			fmt.Println("Getting Bluesky profile for handle " + dmValidationRequest.BskyHandle)
			accessJwt, endpoint, err := shared.LoginToBsky()

			profile, err := shared.GetProfile(dmValidationRequest.BskyHandle, accessJwt, endpoint)
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

			// store DM speaker and add to bsky starter pack
			fmt.Println("Storing DM speaker and adding to Bluesky starter pack")
			result, err := shared.StoreAndAddToBskyStarterPack(naming, dmValidationRequest.VerificationId, dmValidationRequest.BskyHandle, profile.DID, "dynamicsminds", accessJwt, endpoint)

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

func main() {}
