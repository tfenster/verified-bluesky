package main

import (
	"fmt"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

func init() {

	moduleSpecifics, _ := shared.GetModuleSpecifics("cttt")

	moduleSpecifics.VerificationFunc = func(verificationId string, bskyHandle string) (bool, error) {
		fmt.Println("Validating CTTT speaker with name: " + verificationId + " and Bluesky handle: " + bskyHandle)
		return shared.SessionizeVerification(verificationId, bskyHandle, "lo9br5ry")
	}
	moduleSpecifics.NamingFunc = func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
		return shared.SetupNamingStructure(m)
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
