package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
	"golang.org/x/net/html"
)

func init() {

	moduleKey := "ghstar"
	moduleName := "Github Stars"
	moduleNameShortened := "GitHub Stars"

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
						fmt.Fprintln(w, "This is your ID in the Github Stars list. If you open your profile, it is the last part of the URL after https://stars.github.com/profiles/ and without the / in the end. For this to work, you need to have the link to your Bluesky profile in the Additional links on your Github Stars profile.")
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

			var ghsValidationRequest shared.ValidationRequest
			err = json.Unmarshal(body, &ghsValidationRequest)
			if err != nil {
				http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}
			ghsValidationRequest.BskyHandle = strings.ToLower(ghsValidationRequest.BskyHandle)

			// get Guthub Star profile
			fmt.Println("Validating Github Star with ID: " + ghsValidationRequest.VerificationId)
			url := "https://stars.github.com/profiles/" + ghsValidationRequest.VerificationId

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				http.Error(w, "Error fetching the Github Star profile: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			doc, err := html.Parse(resp.Body)
			if err != nil {
				fmt.Println("Error parsing HTML:", err)
				http.Error(w, "Error parsing the Github Star profile: "+err.Error(), http.StatusInternalServerError)
				return
			}

			found := FindBskyHandle(doc, ghsValidationRequest.BskyHandle)

			if !found {
				fmt.Print("Speaker with ID '" + ghsValidationRequest.VerificationId + "' and handle '" + ghsValidationRequest.BskyHandle + "' not found\n")
				http.Error(w, fmt.Sprintf("Link to social network with handle %s not found for Github Star %s", ghsValidationRequest.BskyHandle, ghsValidationRequest.VerificationId), http.StatusNotFound)
				return
			}

			// get bsky profile
			fmt.Println("Getting Bluesky profile for handle " + ghsValidationRequest.BskyHandle)
			accessJwt, endpoint, err := shared.LoginToBsky()

			profile, err := shared.GetProfile(ghsValidationRequest.BskyHandle, accessJwt, endpoint)
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

			// store Github Star and add to bsky starter pack
			fmt.Println("Storing Github Star and adding to Bluesky starter pack")
			result, err := shared.StoreAndAddToBskyStarterPack(naming, ghsValidationRequest.VerificationId, ghsValidationRequest.BskyHandle, profile.DID, "ghstar", accessJwt, endpoint)

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

func FindBskyHandle(n *html.Node, value string) bool {
	title := "open link in new tab"
	fullProfile := "https://bsky.app/profile/" + value
	if n.Type == html.ElementNode && n.Data == "option" {
        var optionValue, optionTitle string
        for _, attr := range n.Attr {
			fmt.Println("working on attribute: " + attr.Key + " with value: " + attr.Val)
            if attr.Key == "value" {
				fmt.Println("found value: " + attr.Val)
                optionValue = attr.Val
            }
            if attr.Key == "title" {
				fmt.Println("found title: " + attr.Val)
                optionTitle = attr.Val
            }
        }
		fmt.Println("comparing value: " + optionValue + " to " + fullProfile + " and title: " + optionTitle + " to " + title)
		if optionValue == fullProfile && optionTitle == title {
			fmt.Printf("Found <option> tag with value='%s' and title='%s'\n", fullProfile, title)
			return true;
		}
    }
    for c := n.FirstChild; c != nil; c = c.NextSibling {
        retVal := FindBskyHandle(c, value)
		if retVal {
			return true;
		}
    }
	return false;
}

func main() {}
