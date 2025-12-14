package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shared"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

type memberInfoResponse struct {
	Members []string `json:"members"`
}

type phonebookResponse struct {
	People map[string]phonebookEntry `json:"people"`
}

type phonebookEntry struct {
	Name string   `json:"name"`
	URLs []string `json:"urls"`
}

func init() {
	moduleSpecifics, _ := shared.GetModuleSpecifics("afm")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating Apache Foundation Member with ID: " + verificationId)

		if err := ensureIsMember(verificationId); err != nil {
			return false, err
		}

		entry, err := getPhonebookEntry(verificationId)
		if err != nil {
			return false, err
		}

		if containsBlueskyURL(entry.URLs, bskyHandle) {
			fmt.Print("Bluesky link found in phonebook entry\n")
			return true, nil
		}

		return false, fmt.Errorf("Bluesky link https://bsky.app/profile/%s not found for Apache Foundation Member %s", bskyHandle, verificationId)
	}

	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func ensureIsMember(verificationId string) error {
	memberInfo, err := fetchMemberInfo()
	if err != nil {
		return err
	}

	for _, member := range memberInfo.Members {
		if member == verificationId {
			return nil
		}
	}

	return fmt.Errorf("Verification ID %s is not listed as an Apache Foundation Member", verificationId)
}

func fetchMemberInfo() (memberInfoResponse, error) {
	resp, err := shared.SendGet("https://whimsy.apache.org/public//member-info.json", "")
	if err != nil {
		fmt.Println("Error fetching member info: " + err.Error())
		return memberInfoResponse{}, fmt.Errorf("Error fetching Apache Foundation member info: %w", err)
	}
	defer resp.Body.Close()

	var memberInfo memberInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&memberInfo); err != nil {
		fmt.Println("Error decoding member info JSON: " + err.Error())
		return memberInfoResponse{}, fmt.Errorf("Error decoding Apache Foundation member info JSON: %w", err)
	}

	return memberInfo, nil
}

func getPhonebookEntry(verificationId string) (phonebookEntry, error) {
	resp, err := shared.SendGet("https://whimsy.apache.org/public//public_ldap_people.json", "")
	if err != nil {
		fmt.Println("Error fetching phonebook: " + err.Error())
		return phonebookEntry{}, fmt.Errorf("Error fetching Apache Foundation phonebook entry: %w", err)
	}
	defer resp.Body.Close()

	var phonebook phonebookResponse
	if err := json.NewDecoder(resp.Body).Decode(&phonebook); err != nil {
		fmt.Println("Error decoding phonebook JSON: " + err.Error())
		return phonebookEntry{}, fmt.Errorf("Error decoding Apache Foundation phonebook JSON: %w", err)
	}

	entry, ok := phonebook.People[verificationId]
	if !ok {
		return phonebookEntry{}, fmt.Errorf("Phonebook entry not found for Apache Foundation Member %s", verificationId)
	}

	return entry, nil
}

func containsBlueskyURL(urls []string, bskyHandle string) bool {
	expected := "https://bsky.app/profile/" + bskyHandle
	alt := "http://bsky.app/profile/" + bskyHandle
	altNoScheme := "bsky.app/profile/" + bskyHandle

	for _, u := range urls {
		candidate := strings.TrimSpace(u)
		if strings.EqualFold(candidate, expected) || strings.EqualFold(candidate, alt) || strings.EqualFold(candidate, altNoScheme) {
			return true
		}
	}

	return false
}

func main() {}
