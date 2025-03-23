package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {
	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "awshero",
		ModuleName:           "AWS Heroes",
		ModuleNameShortened:  "AWS Heroes",
		ModuleLabel:          "awshero",
		ExplanationText:      "This is your ID in the AWS Heroes list. If you open your profile, it is the last part of the URL after https://aws.amazon.com/developer/community/heroes/ (without everything after \"/?\"). For this to work, you need to have the link to your Bluesky profile in the social links on your AWS Hero profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating AWS Hero with ID: " + verificationId)
			url := "https://aws.amazon.com/developer/community/heroes/" + verificationId + "/?did=dh_card&trk=dh_card"
			xpathQuery := fmt.Sprintf("//a[contains(@href, 'https://bsky.app/profile/%s')]", bskyHandle)
			return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
