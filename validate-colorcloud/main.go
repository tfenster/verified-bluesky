package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {

	moduleSpecifics, _ := shared.GetModuleSpecifics("colorcloud")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating ColorCloud speaker with name: " + verificationId + " and Bluesky handle: " + bskyHandle)
		return shared.SessionizeVerification(verificationId, bskyHandle, "7261xanh")
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
