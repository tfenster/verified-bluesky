package main

import (
	"encoding/json"
	"fmt"
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

	moduleSpecifics, _ := shared.GetModuleSpecifics("dynamicsminds")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating DM speaker with name: " + verificationId)
		url := "https://api.runevents.net/api/sessions-and-speakers/external-speakers?eventSlug=dynamicsminds-2024"

		resp, err := shared.SendGet(url, "")
		if err != nil {
			fmt.Println("Error fetching the URL: " + err.Error())
			return false, fmt.Errorf("Error fetching the DM speaker list: " + err.Error())
		}
		defer resp.Body.Close()

		var response RunEventsResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			fmt.Println("Error decoding DM Speaker JSON: " + err.Error())
			return false, fmt.Errorf("Error decoding DM Speaker JSON: " + err.Error())
		}

		// check if bsky handle is in DM profile
		found := false
		for _, speaker := range response.Data {
			if speaker.Name == verificationId {
				fmt.Print("Speaker with name '" + verificationId + "' found\n")
				if strings.Contains(speaker.Biography, bskyHandle) {
					fmt.Print("Biography contains handle '" + bskyHandle + "'\n")
					found = true
					return true, nil
				}
			}
		}
		if !found {
			fmt.Print("Speaker with name '" + verificationId + "' and handle '" + bskyHandle + "' not found\n")
			return false, fmt.Errorf("Link to social network with handle %s not found for DM speaker %s", bskyHandle, verificationId)
		} else {
			return true, nil
		}
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
