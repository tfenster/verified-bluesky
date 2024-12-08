package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "rd",
		ModuleName:           "Microsoft Regional Directors (RDs)",
		ModuleNameShortened:  "RDs",
		ExplanationText:      "This is your RD ID, a GUID. If you open your profile on <a href=\"https://rd.microsoft.com\" target=\"_blank\">rd.microsoft.com</a>, it is the last part of the URL, after the last /. For this to work, you need to have the link to your Bluesky profile in the list of social networks on your RD profile (use \"Other\" as type).",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			// get RD profile
			fmt.Println("Validating RD with ID: " + verificationId)
			url := fmt.Sprintf("https://mavenapi-prod.azurewebsites.net/api/rd/UserProfiles/public/%s", url.QueryEscape(verificationId))

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				return false, fmt.Errorf("Error fetching the RD profile, probably caused by an invalid RD ID: "+err.Error(), http.StatusInternalServerError)
			}
			defer resp.Body.Close()

			var response Response
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				fmt.Println("Error decoding RD JSON: " + err.Error())
				return false, fmt.Errorf("Error decoding RD JSON, probably caused by an invalid RD ID: "+err.Error(), http.StatusInternalServerError)
			}

			// check if bsky handle is in RD profile
			if containsSocialNetworkWithHandle(response.UserProfile.UserProfileSocialNetwork, bskyHandle) {
				fmt.Print("Social network with handle '" + bskyHandle + "' found\n")
				return true, nil
			} else {
				fmt.Print("Social network with handle '" + bskyHandle + "' not found\n")
				return false, fmt.Errorf("Link to social network with handle %s not found for RD %s", bskyHandle, verificationId)
			}
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
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
