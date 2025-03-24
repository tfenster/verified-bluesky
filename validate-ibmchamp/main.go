package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {
	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "ibmchamp",
		ModuleName:           "IBM Champions",
		ModuleNameShortened:  "IBM Champions",
		ModuleLabel:          "ibmchamp",
		ExplanationText:      "This is your ID in the IBM Champions list. If you open your profile, it is the last part of the URL after https://community.ibm.com/community/user/champions/expert/. For this to work, you need to have the link to your Bluesky profile in the social links on your IBM Champion profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating IBM Champion with ID: " + verificationId)
			url := "https://community.ibm.com/community/user/champions/expert/" + verificationId
			xpathQuery := fmt.Sprintf("//input[contains(@title, 'https://bsky.app/profile/%s')]", bskyHandle)
			return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
