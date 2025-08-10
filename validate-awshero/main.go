package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {
	moduleSpecifics, _ := shared.GetModuleSpecifics("awshero")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating AWS Hero with ID: " + verificationId)
		url := "https://aws.amazon.com/developer/community/heroes/" + verificationId + "/?did=dh_card&trk=dh_card"
		xpathQuery := fmt.Sprintf("//a[contains(@href, 'https://bsky.app/profile/%s')]", bskyHandle)
		return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
