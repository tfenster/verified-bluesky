package main

import (
	"fmt"

	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
	"golang.org/x/net/html"
)

func init() {

	aceLevels := map[string][]string{
		"Associate": {},
		"Pro": {},
		"Director": {},
	}

	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "oracleace",
		ModuleName:           "Oracle ACEs",
		ModuleNameShortened:  "Oracle ACEs",
		ModuleLabel:          "oracleace",
		ExplanationText:      "This is your ID in the Oracle ACEs list. If you open your profile, it is the numerical last part of the URL after https://apexapps.oracle.com/apex/ace/profile/<name><ID>. Make sure to only put in the numbers. For this to work, you need to have the link to your Bluesky profile in the Social links on your Oracle ACE profile.",
		FirstAndSecondLevel:  aceLevels,
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating Oracle ACE with ID: " + verificationId)
			url := "https://apexapps.oracle.com/pls/apex/r/ace_program/oracle-aces/ace?ace_id=" + verificationId + "&clear=2"

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				return false, fmt.Errorf("Error fetching the Oracle ACE profile: "+err.Error())
			}
			defer resp.Body.Close()

			doc, err := html.Parse(resp.Body)
			if err != nil {
				fmt.Println("Error parsing HTML:", err)
				return false, fmt.Errorf("Error parsing the Oracle ACE profile: "+err.Error())
			}

			found := FindBskyHandle(doc, bskyHandle)

			if !found {
				fmt.Print("Oracle ACE with ID '" + verificationId + "' and handle '" + bskyHandle + "' not found\n")
				return false, fmt.Errorf("Link to social network with handle %s not found for Oracle ACE %s", bskyHandle, verificationId)
			}

			return true, nil
		},
		NamingFunc: func(m shared.ModuleSpecifics, verificationId string) (shared.Naming, error) {
			fmt.Println("Getting Oracle ACE Level with ID: " + verificationId)
			url := "https://apexapps.oracle.com/pls/apex/r/ace_program/oracle-aces/ace?ace_id=" + verificationId + "&clear=2"

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				return shared.Naming{}, fmt.Errorf("Error fetching the Oracle ACE profile: "+err.Error())
			}
			defer resp.Body.Close()

			doc, err := html.Parse(resp.Body)
			if err != nil {
				fmt.Println("Error parsing HTML:", err)
				return shared.Naming{}, fmt.Errorf("Error parsing the Oracle ACE profile: "+err.Error())
			}
			firstAndSecondLevel := map[string][]string{}
			aceLevel := FindACELevel(doc, verificationId)
			if aceLevel == "" {
				fmt.Println("Could not identify ACE Level for Oracle ACE with ID " + verificationId)
				return shared.Naming{}, fmt.Errorf("Could not identifiy ACE Level for Oracle ACE with ID %s", verificationId)
			}
			firstAndSecondLevel[aceLevel] = []string{}
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
				NamingFunc: 		  m.NamingFunc,
			})
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func FindBskyHandle(n *html.Node, value string) bool {
	title := "Bluesky"
	fullProfile := "https://bsky.app/profile/" + value
	if n.Type == html.ElementNode && n.Data == "a" {
        var linkHref, linkTitle string
        for _, attr := range n.Attr {
			fmt.Println("working on attribute: " + attr.Key + " with value: " + attr.Val)
            if attr.Key == "href" {
				fmt.Println("found value: " + attr.Val)
                linkHref = attr.Val
            }
            if attr.Key == "title" {
				fmt.Println("found title: " + attr.Val)
                linkTitle = attr.Val
            }
        }
		fmt.Println("comparing href: " + linkHref + " to " + fullProfile + " and title: " + linkTitle + " to " + title)
		if linkHref == fullProfile && linkTitle == title {
			fmt.Printf("Found <a> tag with href='%s' and title='%s'\n", fullProfile, title)
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

func FindACELevel(n *html.Node, value string) string {
	id := "ace-Level"
	if n.Type == html.ElementNode && n.Data == "img" {
        var imgId, imgAlt string
        for _, attr := range n.Attr {
			fmt.Println("working on attribute: " + attr.Key + " with value: " + attr.Val)
            if attr.Key == "id" {
				fmt.Println("found id: " + attr.Val)
                imgId = attr.Val
            }
            if attr.Key == "alt" {
				fmt.Println("found alt: " + attr.Val)
                imgAlt = attr.Val
            }
        }
		fmt.Println("comparing id: " + imgId + " to " + id)
		if imgId == id {
			return strings.Split(imgAlt, " ")[1]
		}
    }
    for c := n.FirstChild; c != nil; c = c.NextSibling {
        retVal := FindACELevel(c, value)
		if retVal != "" {
			return retVal;
		}
    }
	return "";
}

func main() {}
