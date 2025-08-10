package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {
	moduleSpecifics, _ := shared.GetModuleSpecifics("ibmchamp")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating IBM Champion with ID: " + verificationId)
		url := "https://community.ibm.com/community/user/champions/expert/" + verificationId
		xpathQuery := fmt.Sprintf("//input[contains(@title, 'https://bsky.app/profile/%s')]", bskyHandle)
		return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
