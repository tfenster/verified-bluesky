package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {
	moduleSpecifics, _ := shared.GetModuleSpecifics("cncfamb")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating CNCF Ambassador with ID: " + verificationId)
		url := "https://www.cncf.io/people/ambassadors/?p=" + verificationId
		xpathQuery := fmt.Sprintf("//div[contains(@class, 'person__padding')]//button[@data-modal-slug='%s']/following::a[@href='https://bsky.app/profile/%s']", verificationId, bskyHandle)
		return shared.HtmlXpathVerification(url, xpathQuery, bskyHandle)
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
