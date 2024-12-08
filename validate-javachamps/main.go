package main

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/shared"
)

type Social struct {
    Twitter    string `yaml:"twitter,omitempty"`
    Mastodon   string `yaml:"mastodon,omitempty"`
    Bluesky    string `yaml:"bluesky,omitempty"`
    Youtube    string `yaml:"youtube,omitempty"`
    Linkedin   string `yaml:"linkedin,omitempty"`
    Github     string `yaml:"github,omitempty"`
    Website    string `yaml:"website,omitempty"`
    Sessionize string `yaml:"sessionize,omitempty"`
    Xing       string `yaml:"xing,omitempty"`
}

type Member struct {
    Name    string   `yaml:"name"`
    Social  Social   `yaml:"social"`
    Avatar  string   `yaml:"avatar"`
    Status  []string `yaml:"status,omitempty"`
}

type JCResponse struct {
    Members []Member `yaml:"members"`
}

func init() {

	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "javachamps",
		ModuleName:           "Java Champions",
		ModuleNameShortened:  "Java Champions",
		ModuleLabel:          "javachamps",
		ExplanationText:      "This is your name, exactly as it appears on the Java Champions page. For this to work, you need to have the link to your Bluesky profile (https://bsky.app/profile/...) somewhere in your social links.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			fmt.Println("Validating Java Champion with name: " + verificationId)
			url := "https://javachampions.org/resources/java-champions.yml"

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				return false, fmt.Errorf("Error fetching the Java Champion list: "+err.Error())
			}
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return false, fmt.Errorf("Error reading response body: "+err.Error())
			}
			defer resp.Body.Close()

			fmt.Println(string(respBody))

			var response JCResponse
			err = yaml.Unmarshal(respBody, &response)
			if err != nil {
				fmt.Println("Error decoding Java Champion YAML: " + err.Error())
				return false, fmt.Errorf("Error decoding Java Champion YAML: "+err.Error())
			}

			// check if bsky handle is in JC profile
			found := false
			for _, member := range response.Members {
				if member.Name == verificationId {
					fmt.Print("Java Champion with name '" + verificationId + "' found\n")
					if member.Social.Bluesky == "https://bsky.app/profile/" + bskyHandle {
						fmt.Print("Java Champion with name '" + verificationId + "' and handle '" + bskyHandle + "' found\n")
						found = true
						return true, nil
					}
				}
			}
			if !found {
				fmt.Print("Java Champion with name '" + verificationId + "' and handle '" + bskyHandle + "' not found\n")
				return false, fmt.Errorf("Link to social network with handle %s not found for Java Champion %s", bskyHandle, verificationId)
			} else {
				return true, nil
			}
		},
		NamingFunc: func(m shared.ModuleSpecifics, _ string) (shared.Naming, error) {
			return shared.SetupNamingStructure(m)
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func main() {}
