package main

import (
	"fmt"

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
			xpathQuery := fmt.Sprintf("//div[contains(@class, 'person__padding')]//button[@data-modal-slug='%s']/following::a[@href='https://bsky.app/profile/%s']]", verificationId, bskyHandle)
			return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
