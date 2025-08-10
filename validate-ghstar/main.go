package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {
	moduleSpecifics, _ := shared.GetModuleSpecifics("ghstar")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating Github Star with ID: " + verificationId)
		url := "https://stars.github.com/profiles/" + verificationId
		xpathQuery := fmt.Sprintf("//option[@title='open link in new tab' and @value='https://bsky.app/profile/%s']", bskyHandle)

		return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
