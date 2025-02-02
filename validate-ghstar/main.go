package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
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
			xpathQuery := fmt.Sprintf("//option[@title='open link in new tab' and @value='https://bsky.app/profile/%s']", bskyHandle)

			return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
