package main

import (
	"encoding/json"
	"fmt"
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
	moduleSpecifics, _ := shared.GetModuleSpecifics("mvp")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		// get MVP profile
		fmt.Println("Validating MVP with ID: " + verificationId)
		profile, err := getMvpProfile(verificationId)
		if err != nil {
			return false, err
		}

		// check if bsky handle is in MVP profile
		if containsSocialNetworkWithHandle(profile.UserProfile.UserProfileSocialNetwork, bskyHandle) {
			fmt.Print("Social network with handle '" + bskyHandle + "' found\n")
			return true, nil
		} else {
			fmt.Print("Social network with handle '" + bskyHandle + "' not found\n")
			return false, fmt.Errorf(fmt.Sprintf("Link to social network with handle %s not found for MVP %s", bskyHandle, verificationId))
		}
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, verificationId string) (shared.Naming, error) {
		profile, err := getMvpProfile(verificationId)
		if err != nil {
			return shared.Naming{}, err
		}
		firstAndSecondLevel := map[string][]string{}
		for i, awardCategory := range profile.UserProfile.AwardCategory {
			firstAndSecondLevel[awardCategory] = []string{profile.UserProfile.TechnologyFocusArea[i]}
		}
		return shared.SetupNamingStructure(shared.ModuleSpecifics{
			ModuleKey:            m.ModuleKey,
			ModuleName:           m.ModuleName,
			ModuleNameShortened:  m.ModuleNameShortened,
			ModuleLabel:          m.ModuleLabel,
			ExplanationText:      m.ExplanationText,
			FirstAndSecondLevel:  firstAndSecondLevel,
			Level1TranslationMap: m.Level1TranslationMap,
			Level2TranslationMap: m.Level2TranslationMap,
			VerificationFunc:     m.VerificationFunc,
			NamingFunc:           m.NamingFunc,
		})
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func getMvpProfile(verificationId string) (Response, error) {
	url := fmt.Sprintf("https://mavenapi-prod.azurewebsites.net/api/mvp/UserProfiles/public/%s", url.QueryEscape(verificationId))

	resp, err := shared.SendGet(url, "")
	if err != nil {
		fmt.Println("Error fetching the URL: " + err.Error())
		return Response{}, fmt.Errorf("Error fetching the MVP profile, probably caused by an invalid MVP ID: " + err.Error())
	}
	defer resp.Body.Close()

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding MVP JSON: " + err.Error())
		return Response{}, fmt.Errorf("Error decoding MVP JSON, probably caused by an invalid MVP ID: " + err.Error())
	}

	return response, nil
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
