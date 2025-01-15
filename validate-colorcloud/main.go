package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {

	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "colorcloud",
		ModuleName:           "ColorCloud speakers",
		ModuleNameShortened:  "ColorCloud speakers",
		ModuleLabel:          "colorcloud",
		ExplanationText:      "This is your name, exactly as it appears on the ColorCloud speakers page and on Sessionize. For this to work, you need to have the link to your Bluesky profile (https://bsky.app/profile/...) as link of type \"Other\" on your Sessionize profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating ColorCloud speaker with name: " + verificationId + " and Bluesky handle: " + bskyHandle)
			return shared.SessionizeVerification(verificationId, bskyHandle, "7261xanh")
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
