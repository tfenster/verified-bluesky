package shared

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fermyon/spin/sdk/go/v2/kv"
	"github.com/fermyon/spin/sdk/go/v2/variables"
)

type ModuleSpecifics struct {
	ModuleKey            string
	ModuleName           string
	ModuleNameShortened  string
	ModuleLabel          string
	ExplanationText      string
	VerificationFunc     func(verificationId string, bskyHandle string) (bool, error)
	NamingFunc           func(m ModuleSpecifics, verificationId string) (Naming, error)
	FirstAndSecondLevel  map[string][]string
	Level1TranslationMap map[string]string
	Level2TranslationMap map[string]string
}

// GetModuleSpecifics returns the ModuleSpecifics for a given moduleKey
func GetModuleSpecifics(moduleKey string) (ModuleSpecifics, error) {
	switch moduleKey {
	case "mvp":
		return getMvpModuleSpecifics(), nil
	case "awshero":
		return getAwsHeroModuleSpecifics(), nil
	case "rd":
		return getRdModuleSpecifics(), nil
	case "ghstar":
		return getGhStarModuleSpecifics(), nil
	case "javachamps":
		return getJavaChampsModuleSpecifics(), nil
	case "ibmchamp":
		return getIbmChampModuleSpecifics(), nil
	case "oracleace":
		return getOracleAceModuleSpecifics(), nil
	case "cncfamb":
		return getCncfAmbModuleSpecifics(), nil
	case "afm":
		return getAfmModuleSpecifics(), nil
	default:
		return ModuleSpecifics{}, fmt.Errorf("unknown module key: %s", moduleKey)
	}
}

// Module-specific configurations
func getMvpModuleSpecifics() ModuleSpecifics {
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
			"Cloud Security (Microsoft Defender for Cloud, Azure network security products, GitHub Advanced Security)",
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
		"Azure Machine Learning Studio":                                                           "Azure ML Studio",
		"Datacenter Management (Group Policy, System Center)":                                     "Datacenter Management",
		"On-premises and Hybrid AKS, Container Management":                                        "On-prem. & Hybrid AKS, Containers",
		"Enterprise and Platform Security":                                                        "Enterpr. & Platf. Security",
		"Azure Arc (Arc SQL Server, Arc SQL MI)":                                                  "Azure Arc",
		"Azure Database for MySQL":                                                                "Azure DB for MySQL",
		"Azure Database for PostgreSQL":                                                           "Azure DB PostgreSQL",
		"Azure SQL (Database, Pools, Serverless, Hyperscale, Managed Instance, Virtual Machines)": "Azure SQL",
		"Azure Synapse Analytics":                                                                 "Azure Synapse",
		"Data Engineering & Data Science in Fabric":                                               "Data Eng. in Fabric",
		"Database Development & DevOps":                                                           "DB Dev & DevOps",
		"Microsoft Purview - Data Governance":                                                     "Microsoft Purview",
		"Paginated Operational Reports (RDL)":                                                     "Pag. Op. Reports",
		"Real-Time Intelligence":                                                                  "RT Intelligence",
		"SQL Server (on Windows, Linux, Containers)":                                              "SQL Server",
		"SQL Server ML Services":                                                                  "SQL Server ML",
		"Azure IoT Services & Development":                                                        "Azure IoT Services & Dev",
		"Microsoft advanced content management and experiences":                                   "Advanced content mmgmt",
		"Azure Application PaaS":                                                                  "Application PaaS",
		"Azure Compute Infrastructure":                                                            "Compute Infrastructure",
		"Azure Cost, Resource & Configuration Management":                                         "Cost, Resource & Conf Mg.",
		"Azure HPC & AI Infrastructure":                                                           "HPC & AI Infrastructure",
		"Azure Hybrid & Migration":                                                                "Hybrid & Migration",
		"Azure Infrastructure as Code":                                                            "Infrastructure as Code",
		"Azure Innovation Hub":                                                                    "Innovation Hub",
		"Azure Integration PaaS":                                                                  "Integration PaaS",
		"Azure Kubernetes and Open Source":                                                        "K8s and Open Source",
		"Azure Networking":                                                                        "Networking",
		"Azure Storage":                                                                           "Storage",
		"Azure Well-Architected, Resiliency & Observability":                                      "Well-Architected etc.",
		"Cloud Security (Microsoft Defender for Cloud, Azure network security products, GitHub Advanced Security)": "Cloud Security",
		"Azure network security products":                                "Azure net security",
		"GitHub Advanced Security":                                       "GitHub Adv. Security",
		"Microsoft Security Copilot":                                     "Microsoft Sec. Copilot",
		"SIEM & XDR (Microsoft Sentinel & Microsoft Defender XDR suite)": "SIEM & XDR",
		"Azure Virtual Desktop":                                          "Azure VD",
	}

	return ModuleSpecifics{
		ModuleKey:            "mvp",
		ModuleName:           "Microsoft Most Valuable Professionals (MVPs)",
		ModuleNameShortened:  "MVPs",
		ModuleLabel:          "ms-mvp",
		ExplanationText:      "This is your MVP ID, a GUID. If you open your profile on <a href=\"https://mvp.microsoft.com\" target=\"_blank\">mvp.microsoft.com</a>, it is the last part of the URL, after the last /. For this to work, you need to have the link to your Bluesky profile in the list of social networks on your MVP profile (use \"Other\" as type).",
		FirstAndSecondLevel:  mvpAwardsAndTechnologyFocusAreas,
		Level1TranslationMap: mvpAwardTranslationMap,
		Level2TranslationMap: mvpTechFocusTranslationMap,
	}
}

func getAwsHeroModuleSpecifics() ModuleSpecifics {
	return ModuleSpecifics{
		ModuleKey:            "awshero",
		ModuleName:           "AWS Heroes",
		ModuleNameShortened:  "AWS Heroes",
		ModuleLabel:          "awshero",
		ExplanationText:      "This is your ID in the AWS Heroes list. If you open your profile, it is the last part of the URL after https://builder.aws.com/community/heroes/. For this to work, you need to have the link to your Bluesky profile in the social links on your AWS Hero profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func getRdModuleSpecifics() ModuleSpecifics {
	return ModuleSpecifics{
		ModuleKey:            "rd",
		ModuleName:           "Microsoft Regional Directors (RDs)",
		ModuleNameShortened:  "RDs",
		ModuleLabel:          "ms-rd",
		ExplanationText:      "This is your RD ID, a GUID. If you open your profile on <a href=\"https://rd.microsoft.com\" target=\"_blank\">rd.microsoft.com</a>, it is the last part of the URL, after the last /. For this to work, you need to have the link to your Bluesky profile in the list of social networks on your RD profile (use \"Other\" as type).",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func getGhStarModuleSpecifics() ModuleSpecifics {
	return ModuleSpecifics{
		ModuleKey:            "ghstar",
		ModuleName:           "Github Stars",
		ModuleNameShortened:  "GitHub Stars",
		ModuleLabel:          "ghstar",
		ExplanationText:      "This is your ID in the Github Stars list. If you open your profile, it is the last part of the URL after https://stars.github.com/profiles/ and without the / in the end. For this to work, you need to have the link to your Bluesky profile in the Additional links on your Github Stars profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func getJavaChampsModuleSpecifics() ModuleSpecifics {
	return ModuleSpecifics{
		ModuleKey:            "javachamps",
		ModuleName:           "Java Champions",
		ModuleNameShortened:  "Java Champions",
		ModuleLabel:          "javachamps",
		ExplanationText:      "This is your name, exactly as it appears on the Java Champions page. For this to work, you need to have the link to your Bluesky profile (https://bsky.app/profile/...) somewhere in your social links.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func getIbmChampModuleSpecifics() ModuleSpecifics {
	return ModuleSpecifics{
		ModuleKey:            "ibmchamp",
		ModuleName:           "IBM Champions",
		ModuleNameShortened:  "IBM Champions",
		ModuleLabel:          "ibmchamp",
		ExplanationText:      "This is your ID in the IBM Champions list. If you open your profile, it is the last part of the URL after https://community.ibm.com/community/user/champions/expert/. For this to work, you need to have the link to your Bluesky profile in the social links on your IBM Champion profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func getOracleAceModuleSpecifics() ModuleSpecifics {

	aceLevels := map[string][]string{
		"Associate": {},
		"Pro":       {},
		"Director":  {},
	}

	return ModuleSpecifics{
		ModuleKey:            "oracleace",
		ModuleName:           "Oracle ACEs",
		ModuleNameShortened:  "Oracle ACEs",
		ModuleLabel:          "oracleace",
		ExplanationText:      "This is your ID in the Oracle ACEs list. This is the last part of the URL after https://apexadb.oracle.com/ords/ace/profile/. For this to work, you need to have the link to your Bluesky profile in the Social links on your Oracle ACE profile.",
		FirstAndSecondLevel:  aceLevels,
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func getCncfAmbModuleSpecifics() ModuleSpecifics {
	return ModuleSpecifics{
		ModuleKey:            "cncfamb",
		ModuleName:           "CNCF Ambassadors",
		ModuleNameShortened:  "CNCF Ambassadors",
		ModuleLabel:          "cncfamb",
		ExplanationText:      "This is your ID in the CNCF Ambassadors list. If you open your profile, it is the last part of the URL after https://www.cncf.io/people/ambassadors/?p=. For this to work, you need to have the link to your Bluesky profile in the social links on your CNCF Ambassador profile.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func getAfmModuleSpecifics() ModuleSpecifics {
	return ModuleSpecifics{
		ModuleKey:            "afm",
		ModuleName:           "Apache Foundation Members",
		ModuleNameShortened:  "Apache Foundation Members",
		ModuleLabel:          "afm",
		ExplanationText:      "This is your ID in the Apache Foundation Members list. You can find it at https://www.apache.org/foundation/members.html. For this to work, you need to have the link to your Bluesky profile in the social links in the Apache Foundation Members phonebook at https://people.apache.org/phonebook.html.",
		FirstAndSecondLevel:  make(map[string][]string),
		Level1TranslationMap: make(map[string]string),
		Level2TranslationMap: make(map[string]string),
	}
}

func (m ModuleSpecifics) Handle(w http.ResponseWriter, r *http.Request) {
	// list of bsky handles that are blacklisted, which means request to verify them will be rejected
	bskyHandleBlacklist := []string{}

	// Check for verify_only query parameter first, then fall back to variable
	verifyOnly := r.URL.Query().Get("verify_only")
	if verifyOnly == "" {
		var err error
		verifyOnly, err = variables.Get("verify_only")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	switch r.Method {

	case http.MethodGet:
		lastSlash := strings.LastIndex(r.URL.Path, "/")
		if lastSlash > 0 {
			title := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
			if title != "" {
				if title == "verificationText" {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, m.ExplanationText)
					return
				}
				fmt.Println("Getting Starter Pack and List for " + title)
				accessJwt, endpoint, err := LoginToBsky()
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
				}
				err = RespondWithStarterPacksAndListsForTitle(title, w, accessJwt, endpoint)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}

		if verifyOnly == "true" {
			http.Error(w, "Method not allowed in verify_only mode", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("Getting all Starter Packs and Lists")
		accessJwt, endpoint, err := LoginToBsky()
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		err = RespondWithAllStarterPacksAndListsForModule(m, w, accessJwt, endpoint)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		// get request body
		validationRequest, err := GetBody(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// check if bsky handle is blacklisted
		for _, blacklistedHandle := range bskyHandleBlacklist {
			if validationRequest.BskyHandle == blacklistedHandle {
				http.Error(w, "The Bluesky handle "+validationRequest.BskyHandle+" has requested to not be verified through this service", http.StatusBadRequest)
				return
			}
		}

		// verify externally
		fmt.Println("Validating with external service")
		verified, err := m.VerificationFunc(validationRequest.VerificationId, validationRequest.BskyHandle)
		if !verified {
			http.Error(w, "Verification failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// get bsky profile
		accessJwt, endpoint, err := LoginToBsky()

		profile, err := GetProfile(validationRequest.BskyHandle, accessJwt, endpoint)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if profile == (ProfileResponse{}) {
			http.Error(w, "Error getting profile", http.StatusInternalServerError)
			return
		}

		naming, err := m.NamingFunc(m, validationRequest.VerificationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if verifyOnly != "true" {
			// store in kv store
			err = Store(naming, validationRequest.VerificationId, validationRequest.BskyHandle)
			if err != nil {
				http.Error(w, "Error storing user in k/v store: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		result := []ListOrStarterPackWithUrl{}

		if verifyOnly == "true" {
			result = append(result, ListOrStarterPackWithUrl{
				Title: naming.Title,
				URL:   "",
			})
			for firstLevel, secondLevels := range naming.FirstAndSecondLevel {
				result = append(result, ListOrStarterPackWithUrl{
					Title: firstLevel.Title,
					URL:   "",
				})
				for _, secondLevel := range secondLevels {
					result = append(result, ListOrStarterPackWithUrl{
						Title: secondLevel.Title,
						URL:   "",
					})
				}
			}
		} else {
			// add to bsky starter pack
			fmt.Println("Adding verified user to Bluesky starter pack")
			result, err = AddToBskyStarterPacksAndList(naming, validationRequest.VerificationId, validationRequest.BskyHandle, profile.DID, m.ModuleLabel, accessJwt, endpoint)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
		if verifyOnly == "true" {
			http.Error(w, "Method not allowed in verify_only mode", http.StatusMethodNotAllowed)
			return
		}
		accessJwt, endpoint, err := LoginToBskyWithReq(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		fmt.Println("Setting up all Starter Packs and Lists")
		err = SetupAllStarterPacksAndLists(m, w, accessJwt, endpoint)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		if verifyOnly == "true" {
			http.Error(w, "Method not allowed in verify_only mode", http.StatusMethodNotAllowed)
			return
		}
		accessJwt, endpoint, err := LoginToBskyWithReq(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		validationRequest, err := GetBody(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		naming, err := SetupNamingStructure(m)
		if err != nil {
			http.Error(w, "Error setting up naming structure: "+err.Error(), http.StatusInternalServerError)
			return
		}

		isInStore, err := CheckStore(naming, validationRequest.VerificationId, validationRequest.BskyHandle)
		if err != nil {
			http.Error(w, "Error checking if User is in store: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !isInStore {
			http.Error(w, "User not found in store", http.StatusBadRequest)
			return
		}

		fmt.Println("Delete user from all Starter Packs and Lists")

		allLists, err := GetLists(accessJwt, endpoint)
		if err != nil {
			http.Error(w, "Error getting lists: "+err.Error(), http.StatusInternalServerError)
			return
		}
		allStarterPacks, err := GetStarterPacks(accessJwt, endpoint)
		if err != nil {
			http.Error(w, "Error getting starter packs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		name := naming.Title
		err = DeleteUserFromStarterPacksAndListWithName(name, validationRequest.BskyHandle, allLists, allStarterPacks, accessJwt, endpoint)
		if err != nil {
			http.Error(w, "Error deleting user "+validationRequest.BskyHandle+" from starter packs and lists "+name+" (root level): "+err.Error(), http.StatusInternalServerError)
			return
		}

		for firstLevel := range naming.FirstAndSecondLevel {
			name = firstLevel.Title
			err = DeleteUserFromStarterPacksAndListWithName(name, validationRequest.BskyHandle, allLists, allStarterPacks, accessJwt, endpoint)
			if err != nil {
				http.Error(w, "Error deleting user "+validationRequest.BskyHandle+" from starter packs and lists "+name+" (first level): "+err.Error(), http.StatusInternalServerError)
				return
			}

			for _, secondLevel := range naming.FirstAndSecondLevel[firstLevel] {
				name = secondLevel.Title
				err = DeleteUserFromStarterPacksAndListWithName(name, validationRequest.BskyHandle, allLists, allStarterPacks, accessJwt, endpoint)
				if err != nil {
					http.Error(w, "Error deleting user "+validationRequest.BskyHandle+" from starter packs and lists "+name+" (second level): "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		err = DeleteFromStore(naming, validationRequest.VerificationId)
		if err != nil {
			http.Error(w, "Error deleting user from k/v store: "+err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func DeleteUserFromStarterPacksAndListWithName(listName string, userToDelete string, allLists []List, allStarterPacks []StarterPack, accessJwt string, endpoint string) error {
	for _, list := range allLists {
		if list.Name == listName {
			_, err := CheckOrDeleteUserOnList(list.URI, userToDelete, true, accessJwt, endpoint)
			if err != nil {
				return fmt.Errorf("Error deleting user from list: %v", err)
			}
		}
	}
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return err
	}
	for _, starterPack := range allStarterPacks {
		if starterPack.Record.Name == listName {
			_, err := CheckOrDeleteUserOnList(starterPack.Record.List, userToDelete, true, accessJwt, endpoint)
			if err != nil {
				return fmt.Errorf("Error deleting user from starter pack list: %v", err)
			}
			now := time.Now()
			timestamp := now.Format("2006-01-02T15:04:05.000Z")
			err = PutRecordForStarterPack(bskyDid, starterPack.URI, starterPack.Record.Description, starterPack.Record.Name, starterPack.Record.CreatedAt, starterPack.Record.List, timestamp, accessJwt, endpoint)
			if err != nil {
				return fmt.Errorf("Error applying change to starter pack: %v", err)
			}
		}
	}
	return nil
}

func GetBody(r *http.Request, w http.ResponseWriter) (ValidationRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ValidationRequest{}, fmt.Errorf("Error reading body: %v", err)
	}
	defer r.Body.Close()

	var validationRequest ValidationRequest
	err = json.Unmarshal(body, &validationRequest)
	if err != nil {
		return ValidationRequest{}, fmt.Errorf("Error decoding body JSON: %v", err)
	}

	// Trim leading and trailing whitespace from input fields
	validationRequest.BskyHandle = strings.TrimSpace(validationRequest.BskyHandle)
	validationRequest.VerificationId = strings.TrimSpace(validationRequest.VerificationId)

	// Convert handle to lowercase after trimming
	validationRequest.BskyHandle = strings.ToLower(validationRequest.BskyHandle)
	return validationRequest, nil
}

func Store(naming Naming, moduleKey string, bskyHandle string) error {
	fmt.Println("Storing verified user in kv store")
	store, err := kv.OpenStore("default")
	if err != nil {
		return err
	}
	defer store.Close()

	key := naming.Key + "-" + moduleKey

	return store.Set(key, []byte(bskyHandle))
}

func CheckStore(naming Naming, moduleKey string, bskyHandle string) (bool, error) {
	store, err := kv.OpenStore("default")
	if err != nil {
		return false, err
	}
	defer store.Close()

	key := naming.Key + "-" + moduleKey

	exists, err := store.Exists(key)
	if (err != nil) || (!exists) {
		return false, err
	}

	value, err := store.Get(key)
	if err != nil {
		return false, err
	}
	return string(value) == bskyHandle, nil
}

func DeleteFromStore(naming Naming, moduleKey string) error {
	store, err := kv.OpenStore("default")
	if err != nil {
		return err
	}
	defer store.Close()

	key := naming.Key + "-" + moduleKey

	exists, err := store.Exists(key)
	if err != nil {
		return err
	}

	if exists {
		err = store.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}
