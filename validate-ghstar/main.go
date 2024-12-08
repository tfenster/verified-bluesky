package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
	"golang.org/x/net/html"
)

func init() {
	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "ghstar",
		ModuleName:           "Github Stars",
		ModuleNameShortened:  "GitHub Stars",
		ModuleLabel:          "ghstar",
		ExplanationText:      "This is your ID in the Github Stars list. If you open your profile, it is the last part of the URL after https://stars.github.com/profiles/ and without the / in the end. For this to work, you need to have the link to your Bluesky profile in the Additional links on your Github Stars profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating Github Star with ID: " + verificationId)
			url := "https://stars.github.com/profiles/" + verificationId

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				return false, fmt.Errorf("Error fetching the Github Star profile: "+err.Error())
			}
			defer resp.Body.Close()

			doc, err := html.Parse(resp.Body)
			if err != nil {
				fmt.Println("Error parsing HTML:", err)
				return false, fmt.Errorf("Error parsing the Github Star profile: "+err.Error())
			}

			found := FindBskyHandle(doc, bskyHandle)

			if !found {
				fmt.Print("Github Star with ID '" + verificationId + "' and handle '" + bskyHandle + "' not found\n")
				return false, fmt.Errorf("Link to social network with handle %s not found for Github Star %s", bskyHandle, verificationId)
			}

			return true, nil
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
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
