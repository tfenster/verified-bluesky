package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

type GithubStarsResponse struct {
	Data struct {
		PublicProfile *struct {
			Username string `json:"username"`
			Links    []struct {
				ID       string `json:"id"`
				Link     string `json:"link"`
				Platform string `json:"platform"`
			} `json:"links"`
		} `json:"publicProfile"`
	} `json:"data"`
	Errors []interface{} `json:"errors"`
}

func init() {
	moduleSpecifics, _ := shared.GetModuleSpecifics("ghstar")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating GitHub Star with ID: " + verificationId)

		// GraphQL query to get the user's links
		graphqlQuery := map[string]interface{}{
			"operationName": "GetStars",
			"variables": map[string]interface{}{
				"username": verificationId,
			},
			"query": `
query GetStars($username: String!) {
  publicProfile(username: $username) {
    username
    links {
      id
      link
      platform
      __typename
    }
  }
}`,
		}

		payload, err := json.Marshal(graphqlQuery)
		if err != nil {
			return false, fmt.Errorf("failed to marshal GraphQL query: %w", err)
		}

		// Make POST request to GitHub Stars API
		resp, err := http.Post(
			"https://api-stars.github.com/",
			"application/json",
			bytes.NewBuffer(payload),
		)
		if err != nil {
			return false, fmt.Errorf("failed to make request to GitHub Stars API: %w", err)
		}
		defer resp.Body.Close()

		// Decode the response directly into the struct
		var result GithubStarsResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return false, fmt.Errorf("failed to decode response: %w", err)
		}

		// Check if there's an error in the response
		if len(result.Errors) > 0 {
			return false, fmt.Errorf("GraphQL error: %v", result.Errors)
		}

		// Check if the public profile was found
		if result.Data.PublicProfile == nil {
			return false, fmt.Errorf("GitHub Star profile not found")
		}

		// Check if the bskyHandle appears in any of the links
		expectedURL := "https://bsky.app/profile/" + bskyHandle
		for _, link := range result.Data.PublicProfile.Links {
			if link.Link == expectedURL {
				return true, nil
			}
		}

		return false, fmt.Errorf("bsky handle not found in GitHub Star profile")
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
