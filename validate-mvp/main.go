package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/shared"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
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
		"Cloud Security (Microsoft Defender for Cloud, Azure network security products, GitHub Advanced Security)":                  "Cloud Security",
		"Azure network security products":                                "Azure net security",
		"GitHub Advanced Security":                                       "GitHub Adv. Security",
		"Microsoft Security Copilot":                                     "Microsoft Sec. Copilot",
		"SIEM & XDR (Microsoft Sentinel & Microsoft Defender XDR suite)": "SIEM & XDR",
		"Azure Virtual Desktop":                                          "Azure VD",
	}

	moduleSpecifics := shared.ModuleSpecifics{
		ModuleKey:            "mvp",
		ModuleName:           "Microsoft Most Valuable Professionals (MVPs)",
		ModuleNameShortened:  "MVPs",
		ModuleLabel:          "ms-mvp",
		ExplanationText:      "This is your MVP ID, a GUID. If you open your profile on <a href=\"https://mvp.microsoft.com\" target=\"_blank\">mvp.microsoft.com</a>, it is the last part of the URL, after the last /. For this to work, you need to have the link to your Bluesky profile in the list of social networks on your MVP profile (use \"Other\" as type).",
		FirstAndSecondLevel:  mvpAwardsAndTechnologyFocusAreas,
		Level1TranslationMap: mvpAwardTranslationMap,
		Level2TranslationMap: mvpTechFocusTranslationMap,
		VerificationFunc: func(verificationId string, bskyHandle string) (bool, error) {
			// get MVP profile
			fmt.Println("Validating MVP with ID: " + verificationId)
			profile, err := getMvpProfile(verificationId)
			if err != nil {
				return false, err
			}

			// check if bsky handle is in MVP profile
			if containsSocialNetworkWithHandle(profile.UserProfile.UserProfileSocialNetwork, bskyHandle) {
				fmt.Print("Social network with handle '" + bskyHandle + "' found\n")
				return true, nil
			} else {
				fmt.Print("Social network with handle '" + bskyHandle + "' not found\n")
				return false, fmt.Errorf(fmt.Sprintf("Link to social network with handle %s not found for MVP %s", bskyHandle, verificationId))
			}
		},
		NamingFunc: func(m shared.ModuleSpecifics, verificationId string) (shared.Naming, error) {
			profile, err := getMvpProfile(verificationId)
			if err != nil {
				return shared.Naming{}, err
			}
			firstAndSecondLevel := map[string][]string{}
			for i, awardCategory := range profile.UserProfile.AwardCategory {
				firstAndSecondLevel[awardCategory] = []string{profile.UserProfile.TechnologyFocusArea[i]}
			}
			return shared.SetupNamingStructure(shared.ModuleSpecifics{
				ModuleKey:            m.ModuleKey,
				ModuleName:           m.ModuleName,
				ModuleNameShortened:  m.ModuleNameShortened,
				ModuleLabel:          m.ModuleLabel,
				ExplanationText:      m.ExplanationText,
				FirstAndSecondLevel:  firstAndSecondLevel,
				Level1TranslationMap: m.Level1TranslationMap,
				Level2TranslationMap: m.Level2TranslationMap,
				VerificationFunc:     m.VerificationFunc,
				NamingFunc: 		  m.NamingFunc,
			})
		},
	}

	spinhttp.Handle(moduleSpecifics.Handle)
}

func getMvpProfile(verificationId string) (Response, error) {
	url := fmt.Sprintf("https://mavenapi-prod.azurewebsites.net/api/mvp/UserProfiles/public/%s", url.QueryEscape(verificationId))

	resp, err := shared.SendGet(url, "")
	if err != nil {
		fmt.Println("Error fetching the URL: " + err.Error())
		return Response{}, fmt.Errorf("Error fetching the MVP profile, probably caused by an invalid MVP ID: "+err.Error())
	}
	defer resp.Body.Close()

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding MVP JSON: " + err.Error())
		return Response{}, fmt.Errorf("Error decoding MVP JSON, probably caused by an invalid MVP ID: "+err.Error())
	}

	return response, nil
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
