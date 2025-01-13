package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {

	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "cttt",
		ModuleName:           "Cloud Technology Townhall Tallinn speakers",
		ModuleNameShortened:  "CTTT speakers",
		ModuleLabel: 		  "cttt",
		ExplanationText:      "This is your name, exactly as it appears on the CTTT speakers page and on Sessionize. For this to work, you need to have the link to your Bluesky profile (https://bsky.app/profile/...) as link of type \"Other\" on your Sessionize profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating CTTT speaker with name: " + verificationId + " and Bluesky handle: " + bskyHandle) 
			return shared.SessionizeVerification(verificationId, bskyHandle, "lo9br5ry")
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
