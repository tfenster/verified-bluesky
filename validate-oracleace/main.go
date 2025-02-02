package main

import (
	"fmt"

	"strings"

	"github.com/antchfx/htmlquery"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
	"golang.org/x/net/html"
)

func init() {

	aceLevels := map[string][]string{
		"Associate": {},
		"Pro":       {},
		"Director":  {},
	}

	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "oracleace",
		ModuleName:           "Oracle ACEs",
		ModuleNameShortened:  "Oracle ACEs",
		ModuleLabel:          "oracleace",
		ExplanationText:      "This is your ID in the Oracle ACEs list. This is the last part of the URL after https://apexapps.oracle.com/apex/ace/profile/. For this to work, you need to have the link to your Bluesky profile in the Social links on your Oracle ACE profile.",
		FirstAndSecondLevel:  aceLevels,
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating Oracle ACE with ID: " + verificationId)
			url := "https://apexapps.oracle.com/apex/ace/profile/" + verificationId
			xpathQuery := fmt.Sprintf("//a[@href='https://bsky.app/profile/%s' and @title='Bluesky']", bskyHandle)

			return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
		},
		NamingFunc: func(m shared.ModuleSpecifics, verificationId string) (shared.Naming, error) {
			fmt.Println("Getting Oracle ACE Level with ID: " + verificationId)
			url := "https://apexapps.oracle.com/apex/ace/profile/" + verificationId

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				return shared.Naming{}, fmt.Errorf("Error fetching the Oracle ACE profile: " + err.Error())
			}
			defer resp.Body.Close()

			doc, err := html.Parse(resp.Body)
			if err != nil {
				fmt.Println("Error parsing HTML:", err)
				return shared.Naming{}, fmt.Errorf("Error parsing the Oracle ACE profile: " + err.Error())
			}
			firstAndSecondLevel := map[string][]string{}
			aceLevel, err := FindACELevel(doc, verificationId, url)
			if err != nil {
				return shared.Naming{}, err
			}
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
				NamingFunc:           m.NamingFunc,
			})
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func FindACELevel(doc *html.Node, value string, url string) (string, error) {
	xpathQuery := "//img[@id='ace-Level']"
	nodes, err := htmlquery.QueryAll(doc, xpathQuery)
	if err != nil {
		fmt.Println("Error performing XPath query: %v", err)
		return "", fmt.Errorf("Could not find ACE level on the ACE profile at " + url + ": " + err.Error())
	}
	if len(nodes) == 0 {
		fmt.Println("Could not find ACE level on the ACE profile at " + url)
		return "", fmt.Errorf("Could not find ACE level on the ACE profile at " + url)
	}
	return strings.Split(htmlquery.SelectAttr(nodes[0], "alt"), " ")[1], nil
}

func main() {}
