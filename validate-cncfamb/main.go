package main

import (
	"fmt"

	"github.com/antchfx/htmlquery"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {
	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "cncfamb",
		ModuleName:           "CNCF Ambassadors",
		ModuleNameShortened:  "CNCF Ambassadors",
		ModuleLabel:          "cncfamb",
		ExplanationText:      "This is your ID in the CNCF Ambassadors list. If you open your profile, it is the last part of the URL after https://www.cncf.io/people/ambassadors/?p=. For this to work, you need to have the link to your Bluesky profile in the social links on your CNCF Ambassador profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating CNCF Ambassador with ID: " + verificationId)
			url := "https://www.cncf.io/people/ambassadors/?p=" + verificationId

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				return false, fmt.Errorf("Error fetching the CNCF Ambassador profile: "+err.Error())
			}
			defer resp.Body.Close()

			doc, err := htmlquery.Parse(resp.Body)
			if err != nil {
				fmt.Println("Error parsing HTML:", err)
				return false, fmt.Errorf("Error parsing the CNCF Ambassador profile: "+err.Error())
			}

			xpathQuery := fmt.Sprintf("//div[contains(@class, 'person__padding')]//button[@data-modal-slug='%s']/following::a[@href='https://bsky.app/profile/%s']]", verificationId, bskyHandle)
			fmt.Println("XPath query: " + xpathQuery)
			nodes, err := htmlquery.QueryAll(doc, xpathQuery)
			if err != nil {
				fmt.Println("Error performing XPath query: %v", err)
				return false, fmt.Errorf("Could not find Bluesky URL https://bsky.app/profile/" + bskyHandle + " on the CNCF Ambassador profile of " + verificationId + ": "+err.Error())
			}
			
			if (len(nodes) == 0) {
				fmt.Println("Could not find Bluesky URL https://bsky.app/profile/" + bskyHandle + " on the CNCF Ambassador profile of " + verificationId)
				return false, fmt.Errorf("Could not find Bluesky URL https://bsky.app/profile/" + bskyHandle + " on the CNCF Ambassador profile of " + verificationId)
			}
			return true, nil
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}