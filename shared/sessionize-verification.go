package shared

import (
	"encoding/json"
	"fmt"
)

type Link struct {
    Title    string `json:"title"`
    URL      string `json:"url"`
    LinkType string `json:"linkType"`
}

type Session struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type Profile struct {
    ID             string    `json:"id"`
    FirstName      string    `json:"firstName"`
    LastName       string    `json:"lastName"`
    FullName       string    `json:"fullName"`
    Bio            string    `json:"bio"`
    TagLine        string    `json:"tagLine"`
    ProfilePicture string    `json:"profilePicture"`
    Sessions       []Session `json:"sessions"`
    IsTopSpeaker   bool      `json:"isTopSpeaker"`
    Links          []Link    `json:"links"`
    QuestionAnswers []interface{} `json:"questionAnswers"`
    Categories     []interface{} `json:"categories"`
}

func SessionizeVerification(verificationId string, bskyHandle string, sessionizeId string) (bool, error) {
	fmt.Println("Verifying Sessionize speaker " + verificationId + " with Bluesky handle " + bskyHandle + " at event " + sessionizeId)
	profile, err := getSessionizeSpeakerProfileAtEvent(verificationId, sessionizeId)
	if err != nil {
		return false, err
	}
	return checkIfSpeakerHasBlueskyLink(profile, bskyHandle)
}

func getSessionizeSpeakerProfileAtEvent(verificationId string, sessionizeId string) (Profile, error) {
	url := fmt.Sprintf("https://sessionize.com/api/v2/%s/view/Speakers", sessionizeId)

	resp, err := SendGet(url, "")
	if err != nil {
		fmt.Println("Error fetching the URL: " + err.Error())
		return Profile{}, fmt.Errorf("Error fetching the Sessionize speaker list: "+err.Error())
	}
	defer resp.Body.Close()

	var profiles []Profile
	err = json.NewDecoder(resp.Body).Decode(&profiles)
	if err != nil {
		fmt.Println("Error decoding Sessionize speaker list JSON: " + err.Error())
		return Profile{}, fmt.Errorf("Error decoding Sessionize speaker list JSON: "+err.Error())
	}

	for _, profile := range profiles {
		if profile.FullName == verificationId {
			return profile, nil
		}
	}

	return Profile{}, fmt.Errorf("Speaker " + verificationId + " not found in Sessionize speaker list of the event")
}

func checkIfSpeakerHasBlueskyLink(profile Profile, bskyHandle string) (bool, error) {
	for _, link := range profile.Links {
		if link.LinkType == "Other" && link.URL == "https://bsky.app/profile/"+bskyHandle {
			return true, nil
		}
	}
	return false, fmt.Errorf("Speaker does not have the Bluesky link https://bsky.app/profile/" + bskyHandle + " on their Sessionize profile")	
}