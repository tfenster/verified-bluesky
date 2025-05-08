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

func (m ModuleSpecifics) Handle(w http.ResponseWriter, r *http.Request) {
	// list of bsky handles that are blacklisted, which means request to verify them will be rejected
	bskyHandleBlacklist := []string{}

	verifyOnly, err := variables.Get("verify_only")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

		// store in kv store
		err = Store(naming, validationRequest.VerificationId, validationRequest.BskyHandle)
		if err != nil {
			http.Error(w, "Error storing user in k/v store: "+err.Error(), http.StatusInternalServerError)
			return
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
