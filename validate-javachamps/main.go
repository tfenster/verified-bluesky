package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

type Social struct {
    Twitter    string `yaml:"twitter,omitempty"`
    Mastodon   string `yaml:"mastodon,omitempty"`
    Bluesky    string `yaml:"bluesky,omitempty"`
    Youtube    string `yaml:"youtube,omitempty"`
    Linkedin   string `yaml:"linkedin,omitempty"`
    Github     string `yaml:"github,omitempty"`
    Website    string `yaml:"website,omitempty"`
    Sessionize string `yaml:"sessionize,omitempty"`
    Xing       string `yaml:"xing,omitempty"`
}

type Member struct {
    Name    string   `yaml:"name"`
    Social  Social   `yaml:"social"`
    Avatar  string   `yaml:"avatar"`
    Status  []string `yaml:"status,omitempty"`
}

type JCResponse struct {
    Members []Member `yaml:"members"`
}

func init() {

	moduleKey := "javachamps"
	moduleName := "Java Champions"
	moduleNameShortened := "Java Champions"

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
						fmt.Fprintln(w, "This is your name, exactly as it appears on the Java Champions page. For this to work, you need to have the link to your Bluesky profile (https://bsky.app/profile/...) somewhere in your social links.")
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

			var jcValidationRequest shared.ValidationRequest
			err = json.Unmarshal(body, &jcValidationRequest)
			if err != nil {
				http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}
			jcValidationRequest.BskyHandle = strings.ToLower(jcValidationRequest.BskyHandle)

			// get Java Champion profile
			fmt.Println("Validating Java Champion with name: " + jcValidationRequest.VerificationId)
			url := "https://javachampions.org/resources/java-champions.yml"

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				http.Error(w, "Error fetching the Java Champion list: "+err.Error(), http.StatusInternalServerError)
				return
			}
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Error reading response body: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			fmt.Println(string(respBody))

			var response JCResponse
			err = yaml.Unmarshal(respBody, &response)
			if err != nil {
				fmt.Println("Error decoding Java Champion YAML: " + err.Error())
				http.Error(w, "Error decoding Java Champion YAML: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// check if bsky handle is in DM profile
			found := false
			for _, member := range response.Members {
				if member.Name == jcValidationRequest.VerificationId {
					fmt.Print("Java Champion with name '" + jcValidationRequest.VerificationId + "' found\n")
					if member.Social.Bluesky == "https://bsky.app/profile/" + jcValidationRequest.BskyHandle {
						fmt.Print("Java Champion with name '" + jcValidationRequest.VerificationId + "' and handle '" + jcValidationRequest.BskyHandle + "' found\n")
						found = true
						break
					}
				}
			}
			if !found {
				fmt.Print("Java Champion with name '" + jcValidationRequest.VerificationId + "' and handle '" + jcValidationRequest.BskyHandle + "' not found\n")
				http.Error(w, fmt.Sprintf("Link to social network with handle %s not found for Java Champion %s", jcValidationRequest.BskyHandle, jcValidationRequest.VerificationId), http.StatusNotFound)
				return
			}

			// get bsky profile
			fmt.Println("Getting Bluesky profile for handle " + jcValidationRequest.BskyHandle)
			accessJwt, endpoint, err := shared.LoginToBsky()

			profile, err := shared.GetProfile(jcValidationRequest.BskyHandle, accessJwt, endpoint)
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

			// store Java Champion and add to bsky starter pack
			fmt.Println("Storing Java Champion and adding to Bluesky starter pack")
			result, err := shared.StoreAndAddToBskyStarterPack(naming, jcValidationRequest.VerificationId, jcValidationRequest.BskyHandle, profile.DID, moduleKey, accessJwt, endpoint)

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
