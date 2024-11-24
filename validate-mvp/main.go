package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shared"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/variables"
)

type UserProfile struct {
	TechnologyFocusArea      []string        `json:"technologyFocusArea"`
	AwardCategory            []string        `json:"awardCategory"`
	ID                       int             `json:"id"`
	UserProfileSocialNetwork []SocialNetwork `json:"userProfileSocialNetwork"`
}

type SocialNetwork struct {
	ID                int    `json:"id"`
	UserProfileId     int    `json:"userProfileId"`
	SocialNetworkId   int    `json:"socialNetworkId"`
	Handle            string `json:"handle"`
	SocialNetworkName string `json:"socialNetworkName"`
}

type Response struct {
	UserProfile UserProfile `json:"userProfile"`
}

type MvpValidationRequest struct {
	BskyHandle string `json:"bskyHandle"`
	MvpId      string `json:"mvpId"`
}

func init() {

	mvpAwardsAndTechnologyFocusAreas := map[string][]string{
		"AI Platform": {
			"Azure AI Services",
			"Azure AI Studio",
			"Azure Machine Learning Studio",
			"Responsible AI with Azure",
		},
		"Business Applications": {
			"AI ERP",
			"Business Central",
			"Copilot Studio",
			"Customer Experience",
			"Customer Service",
			"Power Apps",
			"Power Automate",
			"Power Pages",
		},
		"Cloud and Datacenter Management": {
			"Datacenter Management (Group Policy, System Center)",
			"Enterprise and Platform Security",
			"High Availability",
			"Hyper-V",
			"Linux on Hyper-V",
			"On-premises and Hybrid AKS, Container Management",
			"Container Management",
			"On-Premises Networking",
			"On-Premises Storage",
			"Windows Server",
		},
		"Data Platform": {
			"Analysis Services",
			"Azure Arc (Arc SQL Server, Arc SQL MI)",
			"Azure Cosmos DB",
			"Azure Data Lake",
			"Azure Database for MySQL",
			"Azure Database for PostgreSQL",
			"Azure SQL (Database, Pools, Serverless, Hyperscale, Managed Instance, Virtual Machines)",
			"Azure Synapse Analytics",
			"Data Engineering & Data Science in Fabric",
			"Data Integration",
			"Database Development & DevOps",
			"Microsoft Fabric",
			"Microsoft Purview - Data Governance",
			"Paginated Operational Reports (RDL)",
			"Power BI",
			"Real-Time Intelligence",
			"SQL Server (on Windows, Linux, Containers)",
			"SQL Server ML Services",
			"Tools & Connectivity",
		},
		"Developer Technologies": {
			".NET",
			"C++",
			"Developer Security",
			"Developer Tools",
			"DevOps",
			"Java",
			"Python",
			"Web Development",
		},
		"Internet of Things": {
			"Azure Edge Devices",
			"Azure IoT Services & Development",
		},
		"M365": {
			"Access",
			"Excel",
			"Exchange",
			"Loop",
			"M365 Copilot",
			"M365 Copilot Extensibility",
			"M365 Development",
			"Mesh",
			"Microsoft 365",
			"Microsoft advanced content management and experiences",
			"Microsoft Graph",
			"Microsoft Stream",
			"Microsoft Teams",
			"Microsoft Viva",
			"OneDrive",
			"OneNote",
			"Outlook",
			"Planner",
			"PowerPoint",
			"SharePoint",
			"Visio",
			"Word",
		},
		"Microsoft Azure": {
			"Azure Application PaaS",
			"Azure Compute Infrastructure",
			"Azure Cost, Resource & Configuration Management",
			"Azure HPC & AI Infrastructure",
			"Azure Hybrid & Migration",
			"Azure Infrastructure as Code",
			"Azure Innovation Hub",
			"Azure Integration PaaS",
			"Azure Kubernetes and Open Source",
			"Azure Networking",
			"Azure Storage",
			"Azure Well-Architected, Resiliency & Observability",
			"PowerShell",
		},
		"Security": {
			"Cloud Security (Microsoft Defender for Cloud)",
			"Azure network security products",
			"GitHub Advanced Security",
			"Identity & Access",
			"Microsoft Intune",
			"Microsoft Purview",
			"Microsoft Security Copilot",
			"SIEM & XDR (Microsoft Sentinel & Microsoft Defender XDR suite)",
		},
		"Windows Development": {
			"Windows Design",
			"Windows Development",
		},
		"Windows and Devices": {
			"Azure Virtual Desktop",
			"Surface",
			"Windows",
			"Windows 365",
		},
	}

	mvpAwardTranslationMap := map[string]string{
		"AI Platform":                     "AI",
		"Business Applications":           "BizApps",
		"Cloud and Datacenter Management": "CDM",
		"Data Platform":                   "Data Plat.",
		"Developer Technologies":          "Dev Tech",
		"Internet of Things":              "IoT",
		"Microsoft Azure":                 "Azure",
		"Windows Development":             "Windows Dev",
	}

	mvpTechFocusTranslationMap := map[string]string{
		"Azure Machine Learning Studio":                       "Azure ML Studio",
		"Datacenter Management (Group Policy, System Center)": "Datacenter Management",
		"On-premises and Hybrid AKS, Container Management":    "On-prem. & Hybrid AKS, Containers",
		"Enterprise and Platform Security":                    "Enterpr. & Platf. Security",
		"Azure Arc (Arc SQL Server, Arc SQL MI)":              "Azure Arc",
		"Azure Database for MySQL":                            "Azure DB for MySQL",
		"Azure Database for PostgreSQL":                       "Azure DB PostgreSQL",
		"Azure SQL (Database, Pools, Serverless, Hyperscale, Managed Instance, Virtual Machines)": "Azure SQL",
		"Azure Synapse Analytics":                                        "Azure Synapse",
		"Data Engineering & Data Science in Fabric":                      "Data Eng. in Fabric",
		"Database Development & DevOps":                                  "DB Dev & DevOps",
		"Microsoft Purview - Data Governance":                            "Microsoft Purview",
		"Paginated Operational Reports (RDL)":                            "Pag. Op. Reports",
		"Real-Time Intelligence":                                         "RT Intelligence",
		"SQL Server (on Windows, Linux, Containers)":                     "SQL Server",
		"SQL Server ML Services":                                         "SQL Server ML",
		"Azure IoT Services & Development":                               "Azure IoT Services & Dev",
		"Microsoft advanced content management and experiences":          "Advanced content mmgmt",
		"Azure Application PaaS":                                         "Application PaaS",
		"Azure Compute Infrastructure":                                   "Compute Infrastructure",
		"Azure Cost, Resource & Configuration Management":                "Cost, Resource & Conf Mg.",
		"Azure HPC & AI Infrastructure":                                  "HPC & AI Infrastructure",
		"Azure Hybrid & Migration":                                       "Hybrid & Migration",
		"Azure Infrastructure as Code":                                   "Infrastructure as Code",
		"Azure Innovation Hub":                                           "Innovation Hub",
		"Azure Integration PaaS":                                         "Integration PaaS",
		"Azure Kubernetes and Open Source":                               "K8s and Open Source",
		"Azure Networking":                                               "Networking",
		"Azure Storage":                                                  "Storage",
		"Azure Well-Architected, Resiliency & Observability":             "Well-Architected etc.",
		"Cloud Security (Microsoft Defender for Cloud)":                  "Cloud Security",
		"Azure network security products":                                "Azure net security",
		"GitHub Advanced Security":                                       "GitHub Adv. Security",
		"Microsoft Security Copilot":                                     "Microsoft Sec. Copilot",
		"SIEM & XDR (Microsoft Sentinel & Microsoft Defender XDR suite)": "SIEM & XDR",
		"Azure Virtual Desktop":                                          "Azure VD",
	}

	moduleKey := "mvp"
	moduleName := "Microsoft Most Valuable Professionals (MVPs)"
	moduleNameShortened := "MVPs"

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			lastSlash := strings.LastIndex(r.URL.Path, "/")
			if lastSlash > 0 {
				title := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
				if title != "" {
					fmt.Println("Getting Starter Pack and List for " + title)
					accessJwt, endpoint, err := shared.LoginToBsky()
					if err != nil {
						http.Error(w, err.Error(), http.StatusUnauthorized)
					}

					bskyHandle, err := variables.Get("bsky_handle")
					if err != nil {
						http.Error(w, err.Error(), http.StatusUnauthorized)
					}
					
					matchingList := shared.ListOrStarterPackWithUrl{}
					allLists, err := shared.GetLists(accessJwt, endpoint)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					for _, l := range allLists {
						if l.Name == title {
							matchingList = shared.ConvertToStruct(l.URI, title, "list", bskyHandle)
							break
						}
					}

					matchingStarterPacks := []shared.ListOrStarterPackWithUrl{}
					allStarterPacks, err := shared.GetStarterPacks(accessJwt, endpoint)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					for _, sp := range allStarterPacks {
						if sp.Record.Name == title {
							matchingStarterPacks = append(matchingStarterPacks, shared.ConvertToStruct(sp.URI, title, "sp", bskyHandle))
						}
					}

					jsonResult, err := json.Marshal(shared.ListAndStarterPacks{List: matchingList, StarterPacks: matchingStarterPacks})
					if err != nil {
						http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)

					fmt.Fprintln(w, string(jsonResult))
					return
				}
			}
			fmt.Println("Getting all Starter Packs and Lists")
			naming, err := shared.SetupFlatNamingStructure(moduleKey, moduleName, moduleNameShortened, mvpAwardsAndTechnologyFocusAreas, mvpAwardTranslationMap, mvpTechFocusTranslationMap)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			jsonResult, err := json.Marshal(naming)
			if err != nil {
				http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, string(jsonResult))

		case http.MethodPost:
			// get request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			var mvpValidationRequest MvpValidationRequest
			err = json.Unmarshal(body, &mvpValidationRequest)
			if err != nil {
				http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}
			mvpValidationRequest.BskyHandle = strings.ToLower(mvpValidationRequest.BskyHandle)

			// get MVP profile
			fmt.Println("Validating MVP with ID: " + mvpValidationRequest.MvpId)
			url := fmt.Sprintf("https://mavenapi-prod.azurewebsites.net/api/mvp/UserProfiles/public/%s", url.QueryEscape(mvpValidationRequest.MvpId))

			resp, err := shared.SendGet(url, "")
			if err != nil {
				fmt.Println("Error fetching the URL: " + err.Error())
				http.Error(w, "Error fetching the MVP profile: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			var response Response
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				fmt.Println("Error decoding MVP JSON: " + err.Error())
				http.Error(w, "Error decoding MVP JSON, probably caused by an invalid MVP ID: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// check if bsky handle is in MVP profile
			if containsSocialNetworkWithHandle(response.UserProfile.UserProfileSocialNetwork, mvpValidationRequest.BskyHandle) {
				fmt.Print("Social network with handle '" + mvpValidationRequest.BskyHandle + "' found\n")
			} else {
				fmt.Print("Social network with handle '" + mvpValidationRequest.BskyHandle + "' not found\n")
				http.Error(w, fmt.Sprintf("Link to social network with handle %s not found for MVP %s", mvpValidationRequest.BskyHandle, mvpValidationRequest.MvpId), http.StatusNotFound)
				return
			}

			// get bsky profile
			fmt.Println("Getting Bluesky profile for handle " + mvpValidationRequest.BskyHandle)
			accessJwt, endpoint, err := shared.LoginToBsky()

			profile, err := shared.GetProfile(mvpValidationRequest.BskyHandle, accessJwt, endpoint)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if profile == (shared.ProfileResponse{}) {
				http.Error(w, "Error getting profile", http.StatusInternalServerError)
				return
			}

			// create naming structure
			firstAndSecondLevel := map[string][]string{}
			for i, awardCategory := range response.UserProfile.AwardCategory {
				firstAndSecondLevel[awardCategory] = []string{response.UserProfile.TechnologyFocusArea[i]}
			}
			naming, err := shared.SetupNamingStructure(moduleKey, moduleName, moduleNameShortened, firstAndSecondLevel, mvpAwardTranslationMap, mvpTechFocusTranslationMap)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// store MVP and add to bsky starter pack
			fmt.Println("Storing MVP and adding to Bluesky starter pack")
			result, err := shared.StoreAndAddToBskyStarterPack(naming, mvpValidationRequest.MvpId, mvpValidationRequest.BskyHandle, profile.DID, accessJwt, endpoint)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			jsonResult, err := json.Marshal(result)
			if err != nil {
				http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, string(jsonResult))

		case http.MethodPut:
			accessJwt, endpoint, err := shared.LoginToBskyWithReq(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}

			fmt.Println("Setting up all Starter Packs and Lists")
			naming, err := shared.SetupNamingStructure(moduleKey, moduleName, moduleNameShortened, mvpAwardsAndTechnologyFocusAreas, mvpAwardTranslationMap, mvpTechFocusTranslationMap)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			result, err := shared.CreateAllStarterPacksAndLists(naming, accessJwt, endpoint)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, result)

		case http.MethodDelete:
			accessJwt, endpoint, err := shared.LoginToBskyWithReq(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}

			fmt.Println("Deleting all Starter Packs and Lists")

			starterPacks, err := shared.GetStarterPacks(accessJwt, endpoint)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, starterPack := range starterPacks {
				listRkey := starterPack.Record.List[strings.LastIndex(starterPack.Record.List, "/")+1:]
				_, err = shared.DeleteList(listRkey, accessJwt, endpoint)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				starterPackRkey := starterPack.URI[strings.LastIndex(starterPack.URI, "/")+1:]
				_, err = shared.DeleteStarterPack(starterPackRkey, accessJwt, endpoint)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			lists, err := shared.GetLists(accessJwt, endpoint)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, list := range lists {
				listRkey := list.URI[strings.LastIndex(list.URI, "/")+1:]
				_, err = shared.DeleteList(listRkey, accessJwt, endpoint)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func containsSocialNetworkWithHandle(socialNetworks []SocialNetwork, handle string) bool {
	for _, sn := range socialNetworks {
		if sn.Handle == handle || sn.Handle == "bsky.app/profile/"+handle {
			return true
		}
	}
	return false
}

func main() {}
