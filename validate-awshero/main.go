package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

type AWSHeroResponse struct {
	Profile struct {
		BasicInfo struct {
			Alias string `json:"alias"`
		} `json:"basicInfo"`
		Socials struct {
			Personal string `json:"personal"`
		} `json:"socials"`
	} `json:"profile"`
}

func init() {
	moduleSpecifics, _ := shared.GetModuleSpecifics("awshero")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating AWS Hero with ID: " + verificationId)

		// Prepare the request payload
		requestBody := map[string]string{
			"alias": verificationId,
		}

		payload, err := json.Marshal(requestBody)
		if err != nil {
			return false, fmt.Errorf("failed to marshal request: %w", err)
		}

		// Create the request
		req, err := http.NewRequest("POST", "https://api.builder.aws.com/ums/getProfileByAlias", bytes.NewBuffer(payload))
		if err != nil {
			return false, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Add("user-agent", "verifiedbsky.net")
		req.Header.Add("content-type", "application/json")
		req.Header.Add("builder-session-token", "dummy")

		// Make the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, fmt.Errorf("failed to make request to AWS Builder API: %w", err)
		}
		defer resp.Body.Close()

		// Decode the response
		var result AWSHeroResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return false, fmt.Errorf("failed to decode response: %w", err)
		}

		// Check if the alias matches
		if result.Profile.BasicInfo.Alias != verificationId {
			return false, fmt.Errorf("AWS Hero profile not found")
		}

		// Check if the bskyHandle appears in the personal social link
		expectedURL := "https://bsky.app/profile/" + bskyHandle
		if result.Profile.Socials.Personal == expectedURL {
			return true, nil
		}

		return false, fmt.Errorf("bsky handle not found in AWS Hero profile")
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
