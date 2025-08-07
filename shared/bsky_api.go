package shared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fermyon/spin/sdk/go/v2/kv"
	"github.com/fermyon/spin/sdk/go/v2/variables"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

type AuthResponse struct {
	Did       string `json:"did"`
	DidDoc    DidDoc `json:"didDoc"`
	Handle    string `json:"handle"`
	Email     string `json:"email"`
	AccessJwt string `json:"accessJwt"`
}

type StarterPackResponse struct {
	StarterPacks []StarterPack `json:"starterPacks"`
	Cursor       string        `json:"cursor"`
}

type StarterPack struct {
	URI    string `json:"uri"`
	CID    string `json:"cid"`
	Record Record `json:"record"`
}

type Record struct {
	Type        string `json:"$type"`
	Description string `json:"description"`
	List        string `json:"list"`
	Name        string `json:"name"`
	CreatedAt   string `json:"createdAt"`
}

type ListsResponse struct {
	Lists  []List `json:"lists"`
	Cursor string `json:"cursor"`
}

type ListResponse struct {
	List   List   `json:"list"`
	Items  []Item `json:"items"`
	Cursor string `json:"cursor"`
}

type Item struct {
	URI     string  `json:"uri"`
	Subject Subject `json:"subject"`
}

type Subject struct {
	DID         string    `json:"did"`
	Handle      string    `json:"handle"`
	DisplayName string    `json:"displayName"`
	Avatar      string    `json:"avatar"`
	CreatedAt   time.Time `json:"createdAt"`
	Description string    `json:"description"`
	IndexedAt   time.Time `json:"indexedAt"`
}

type List struct {
	URI           string    `json:"uri"`
	CID           string    `json:"cid"`
	Name          string    `json:"name"`
	Purpose       string    `json:"purpose"`
	ListItemCount int       `json:"listItemCount"`
	IndexedAt     time.Time `json:"indexedAt"`
	Labels        []string  `json:"labels"`
	Description   string    `json:"description"`
}

type CreateRecordResponse struct {
	URI              string `json:"uri"`
	CID              string `json:"cid"`
	ValidationStatus string `json:"validationStatus"`
}

type Service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

type DidDoc struct {
	Context     []string  `json:"@context"`
	ID          string    `json:"id"`
	AlsoKnownAs []string  `json:"alsoKnownAs"`
	Service     []Service `json:"service"`
}

type ProfileResponse struct {
	DID         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
}

type ListOrStarterPackWithUrl struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type ListAndStarterPacks struct {
	List         ListOrStarterPackWithUrl   `json:"list"`
	StarterPacks []ListOrStarterPackWithUrl `json:"starterPacks"`
}

type ModerationRepoResponse struct {
	Did    string  `json:"did"`
	Handle string  `json:"handle"`
	Labels []Label `json:"labels"`
}

type Label struct {
	Ver int    `json:"ver"`
	Src string `json:"src"`
	Uri string `json:"uri"`
	Val string `json:"val"`
}

func AddToBskyStarterPacksAndList(naming Naming, moduleKey string, bskyHandle string, bskyDid string, label string, accessJwt string, endpoint string) ([]ListOrStarterPackWithUrl, error) {
	starterPacks, err := GetStarterPacks(accessJwt, endpoint)
	if err != nil {
		return []ListOrStarterPackWithUrl{}, err
	}

	lists, err := GetLists(accessJwt, endpoint)
	if err != nil {
		return []ListOrStarterPackWithUrl{}, err
	}

	addedToElements := make([]ListOrStarterPackWithUrl, 0)
	bskyHandleOwner, err := variables.Get("bsky_handle")
	if err != nil {
		return []ListOrStarterPackWithUrl{}, err
	}

	starterPack, err := AddUserToStarterPack(bskyHandle, bskyDid, naming.Title, naming.Description, starterPacks, accessJwt, endpoint)
	if err != nil {
		return []ListOrStarterPackWithUrl{}, err
	}
	addedToElements = append(addedToElements, ConvertToStruct(starterPack, naming.Title, "sp", bskyHandleOwner))

	list, err := AddUserToList(bskyDid, naming.Title, lists, accessJwt, endpoint)
	if err != nil {
		return []ListOrStarterPackWithUrl{}, err
	}
	addedToElements = append(addedToElements, ConvertToStruct(list, naming.Title, "list", bskyHandleOwner))

	for first, secondArray := range naming.FirstAndSecondLevel {
		starterPack, err = AddUserToStarterPack(bskyHandle, bskyDid, first.Title, first.Description, starterPacks, accessJwt, endpoint)
		if err != nil {
			return []ListOrStarterPackWithUrl{}, err
		}
		addedToElements = append(addedToElements, ConvertToStruct(starterPack, first.Title, "sp", bskyHandleOwner))

		list, err = AddUserToList(bskyDid, first.Title, lists, accessJwt, endpoint)
		if err != nil {
			return []ListOrStarterPackWithUrl{}, err
		}
		addedToElements = append(addedToElements, ConvertToStruct(list, first.Title, "list", bskyHandleOwner))

		for _, second := range secondArray {
			starterPack, err = AddUserToStarterPack(bskyHandle, bskyDid, second.Title, second.Description, starterPacks, accessJwt, endpoint)
			if err != nil {
				return []ListOrStarterPackWithUrl{}, err
			}
			addedToElements = append(addedToElements, ConvertToStruct(starterPack, second.Title, "sp", bskyHandleOwner))

			list, err = AddUserToList(bskyDid, second.Title, lists, accessJwt, endpoint)
			if err != nil {
				return []ListOrStarterPackWithUrl{}, err
			}
			addedToElements = append(addedToElements, ConvertToStruct(list, second.Title, "list", bskyHandleOwner))
		}
	}

	_, err = Follow(bskyDid, accessJwt, endpoint)
	if err != nil {
		// only print the error as this is not technically blocking the application usecase
		fmt.Println("Error following user: " + err.Error())
	}

	err = SetLabel(label, bskyHandle, accessJwt, endpoint)
	if err != nil {
		fmt.Println("Error setting label " + label + " on user: " + err.Error())
		return []ListOrStarterPackWithUrl{}, fmt.Errorf("Error setting label " + label + " on user: " + err.Error())
	}

	return addedToElements, nil
}

func ConvertToStruct(uri string, title string, listOrStarterPack string, bskyHandle string) ListOrStarterPackWithUrl {
	ref := uri[strings.LastIndex(uri, "/")+1:]
	if listOrStarterPack == "sp" {
		return ListOrStarterPackWithUrl{URL: "https://bsky.app/starter-pack/" + bskyHandle + "/" + ref, Title: "Starter pack " + title}
	} else {
		return ListOrStarterPackWithUrl{URL: "https://bsky.app/profile/" + bskyHandle + "/lists/" + ref, Title: "List " + title}
	}
}

func LoginToBsky() (string, string, error) {
	store, err := kv.OpenStore("default")
	if err != nil {
		return "", "", err
	}
	defer store.Close()

	accessJwtFromStore, err := store.Get("accessJwt")
	if err != nil && err.Error() != "no such key" {
		return "", "", err
	}
	if accessJwtFromStore != nil && string(accessJwtFromStore) != "" {
		// fmt.Println("Check if accessJwt is still valid")
		url := "https://bsky.social/xrpc/com.atproto.server.getSession"
		sessionResponse, err := SendGetLogConfigurable(url, string(accessJwtFromStore), true)
		if err == nil && sessionResponse.StatusCode == 200 {
			// fmt.Println("AccessJwt is still valid")
			endpointFromStore, err := store.Get("endpoint")
			if err != nil {
				return "", "", err
			}
			return string(accessJwtFromStore), string(endpointFromStore), nil
		}
	}

	fmt.Println("No accessJwt in store or not valid anymore, logging in again")
	bskyPwd, err := variables.Get("bsky_password")
	if err != nil {
		return "", "", err
	}
	return LoginToBskyWithPwd(bskyPwd)
}

func LoginToBskyWithReq(r *http.Request) (string, string, error) {
	pwd := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	return LoginToBskyWithPwd(pwd)
}

func LoginToBskyWithPwd(bskyPwd string) (string, string, error) {
	fmt.Println("Trying to log in to Bluesky")
	url := "https://bsky.social/xrpc/com.atproto.server.createSession"
	bskyHandle, err := variables.Get("bsky_handle")
	if err != nil {
		return "", "", err
	}

	payload := "{\"identifier\": \"" + bskyHandle + "\",\"password\": \"" + bskyPwd + "\"}"

	resp, err := SendPost(url, payload, "")
	defer resp.Body.Close()

	var response AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", "", err
	}

	fmt.Println("Logged in successfully")
	store, err := kv.OpenStore("default")
	if err != nil {
		return "", "", err
	}
	defer store.Close()

	fmt.Println("Storing accessJwt and endpoint")
	err = store.Set("accessJwt", []byte(response.AccessJwt))
	if err != nil {
		fmt.Println("Error storing accessJwt: " + err.Error())
	}
	err = store.Set("endpoint", []byte(response.DidDoc.Service[0].ServiceEndpoint))
	if err != nil {
		fmt.Println("Error storing endpoint: " + err.Error())
	}
	return response.AccessJwt, response.DidDoc.Service[0].ServiceEndpoint, nil
}

func GetProfile(bskyHandle string, accessJwt string, endpoint string) (ProfileResponse, error) {
	fmt.Println("Getting profile for Bluesky handle " + bskyHandle)
	url := endpoint + "/xrpc/app.bsky.actor.getProfile?actor=" + url.QueryEscape(bskyHandle)

	resp, err := SendGet(url, accessJwt)
	if err != nil {
		return ProfileResponse{}, err
	}

	var response ProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ProfileResponse{}, err
	}
	fmt.Println("Got profile successfully with DID " + response.DID)
	return response, nil
}

func GetStarterPacks(accessJwt string, endpoint string) ([]StarterPack, error) {
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return []StarterPack{}, err
	}
	fmt.Println("Getting starter packs for DID " + bskyDid)

	starterPacks := make([]StarterPack, 0)
	hasMore := 0
	counterArg := ""
	for hasMore < 1 {
		url := endpoint + "/xrpc/app.bsky.graph.getActorStarterPacks?limit=100&actor=" + url.QueryEscape(bskyDid) + counterArg

		resp, err := SendGet(url, accessJwt)
		if err != nil {
			return []StarterPack{}, err
		}

		var response StarterPackResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return []StarterPack{}, err
		}
		starterPacks = append(starterPacks, response.StarterPacks...)

		if response.Cursor != "" {
			counterArg = "&cursor=" + response.Cursor
		} else {
			hasMore = 1
		}
	}

	fmt.Printf("Got %d starter packs successfully\n", len(starterPacks))
	return starterPacks, nil
}

func AddUserToStarterPack(bskyHandle string, bskyDid string, starterPackTitle string, starterPackDescription string, starterPacks []StarterPack, accessJwt string, endpoint string) (string, error) {
	fmt.Println("Adding users to the right starter pack (title: " + starterPackTitle + ", description: " + starterPackDescription + ")")
	var starterPackUri string
	var starterPackListUri string
	var createdAt string
	var err error
	done := false
	matchingStarterPacks := []StarterPack{}
	for _, sp := range starterPacks {
		if sp.Record.Name == starterPackTitle {
			matchingStarterPacks = append(matchingStarterPacks, sp)
		}
	}
	fmt.Println("Found " + fmt.Sprintf("%d", len(matchingStarterPacks)) + " matching starter packs")

	if len(matchingStarterPacks) == 0 {
		fmt.Println("No matching starter pack found with title: " + starterPackTitle)
		return "", fmt.Errorf("No matching starter pack found with title: " + starterPackTitle)
	}

	for _, sp := range matchingStarterPacks {
		list, err := GetList(sp.Record.List, accessJwt, endpoint)
		if err != nil {
			return "", err
		}

		fmt.Println("Found existing starter pack with title " + starterPackTitle + " and an item count of " + fmt.Sprintf("%d", list.ListItemCount))
		userOnList, err := CheckOrDeleteUserOnList(sp.Record.List, bskyDid, false, accessJwt, endpoint)
		if err != nil {
			return "", fmt.Errorf("Error checking if user is on list: " + err.Error())
		}
		if userOnList {
			fmt.Println("User is already on existing starter pack")
			return sp.URI, nil
		}
		if list.ListItemCount < 149 {
			fmt.Println("Found existing starter pack with title " + starterPackTitle + " and space left")
			starterPackUri = sp.URI
			starterPackListUri = sp.Record.List
			starterPackDescription = sp.Record.Description
			createdAt = sp.Record.CreatedAt
			done = true
			break
		} else {
			fmt.Println("Found existing starter pack with title " + starterPackTitle + " but it's full")
		}
	}

	// we found matching starter packs but none of them had space left
	if !done {
		fmt.Println("Starter pack list is full, creating a new one")
		timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
		newListResponse, newStarterPackResponse, err := CreateStarterPack(starterPackTitle, starterPackDescription, timestamp, accessJwt, endpoint)
		if err != nil {
			return "", err
		}
		starterPackListUri = newListResponse.URI
		starterPackUri = newStarterPackResponse.URI
		createdAt = timestamp
		fmt.Println("Created new list and starter pack")
	}

	err = AddUserToStarterPackList(bskyDid, starterPackListUri, starterPackUri, starterPackTitle, starterPackDescription, createdAt, accessJwt, endpoint)
	if err != nil {
		return "", err
	}

	fmt.Println("Added users to the right starter pack")
	return starterPackUri, nil
}

func AddUserToList(bskyDid string, listTitle string, lists []List, accessJwt string, endpoint string) (string, error) {
	fmt.Println("Adding users to the right list (title: " + listTitle + ")")
	var listUri string
	var err error
	for _, list := range lists {
		if list.Name == listTitle {
			listUri = list.URI
			fmt.Println("Found existing list with title " + listTitle)
			break
		}
	}

	if listUri == "" {
		fmt.Println("No matching list found with title: " + listTitle)
		return "", fmt.Errorf("No matching list found with title: " + listTitle)
	}

	err = AddUserToStandaloneList(bskyDid, listUri, accessJwt, endpoint)
	if err != nil {
		return "", err
	}

	fmt.Println("Added users to the right list")
	return listUri, nil
}

func CheckOrDeleteUserOnList(listUri string, userToCheckHandleOrDid string, deleteOnMatch bool, accessJwt string, endpoint string) (bool, error) {
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return false, fmt.Errorf("Error getting bsky_did: " + err.Error())
	}
	fmt.Println("Check if user " + userToCheckHandleOrDid + " is on list " + listUri + ". Delete on match? " + fmt.Sprintf("%t", deleteOnMatch))
	hasMore := 0
	counterArg := ""
	for hasMore < 1 {
		url := endpoint + "/xrpc/app.bsky.graph.getList?list=" + listUri + "&limit=100" + counterArg
		resp, err := SendGet(url, accessJwt)
		if err != nil {
			return false, err
		}

		var response ListResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return false, err
		}

		for _, item := range response.Items {
			if item.Subject.Handle == userToCheckHandleOrDid || item.Subject.DID == userToCheckHandleOrDid {
				fmt.Println("User " + userToCheckHandleOrDid + " is on list " + listUri)
				if deleteOnMatch {
					fmt.Println("Deleting user from list")
					err = RemoveUserFromList(bskyDid, item.URI, accessJwt, endpoint)
					if err != nil {
						return true, err
					}
				}
				return true, nil
			}
		}

		if response.Cursor != "" {
			counterArg = "&cursor=" + response.Cursor
		} else {
			hasMore = 1
		}
	}

	fmt.Println("User " + userToCheckHandleOrDid + " is not on list " + listUri)
	return false, nil
}

func RemoveUserFromList(bskyDid string, userUriToRemove string, accessJwt string, endpoint string) error {
	fmt.Println("Removing user with uri " + userUriToRemove)
	url := endpoint + "/xrpc/com.atproto.repo.deleteRecord"
	rkey := userUriToRemove[strings.LastIndex(userUriToRemove, "/")+1:]
	payload := "{ \"collection\":\"app.bsky.graph.listitem\", \"repo\":\"" + bskyDid + "\", \"rkey\":\"" + rkey + "\" }"

	resp, err := SendPost(url, payload, accessJwt)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error removing user from list, status code: " + fmt.Sprintf("%d", resp.StatusCode))
	}
	return nil
}

func CreateAllStarterPacksAndLists(naming Naming, accessJwt string, endpoint string) (string, error) {
	starterPacks, err := GetStarterPacks(accessJwt, endpoint)
	if err != nil {
		return "", err
	}

	lists, err := GetLists(accessJwt, endpoint)
	if err != nil {
		return "", err
	}

	if CheckIfStarterPackExists(naming.Title, starterPacks) {
		fmt.Println("Starter pack with title " + naming.Title + " already exists")
	} else {
		fmt.Println("Creating starter pack with title " + naming.Title + " and description " + naming.Description)
		_, _, err := CreateStarterPack(naming.Title, naming.Description, time.Now().Format("2006-01-02T15:04:05.000Z"), accessJwt, endpoint)
		if err != nil {
			return "", err
		}
	}

	if CheckIfListExists(naming.Title, lists) {
		fmt.Println("List with title " + naming.Title + " already exists")
	} else {
		fmt.Println("Creating List with title " + naming.Title + " and description " + naming.Description)
		_, err := CreateList(naming.Title, naming.Description, accessJwt, endpoint)
		if err != nil {
			return "", err
		}
	}

	for first, secondArray := range naming.FirstAndSecondLevel {
		if CheckIfStarterPackExists(first.Title, starterPacks) {
			fmt.Println("Starter pack with title " + first.Title + " already exists")
		} else {
			fmt.Println("Creating starter pack with title " + first.Title + " and description " + first.Description)
			_, _, err := CreateStarterPack(first.Title, first.Description, time.Now().Format("2006-01-02T15:04:05.000Z"), accessJwt, endpoint)
			if err != nil {
				return "", err
			}
		}
		if CheckIfListExists(first.Title, lists) {
			fmt.Println("List with title " + first.Title + " already exists")
		} else {
			fmt.Println("Creating List with title " + first.Title + " and description " + first.Description)
			_, err := CreateList(first.Title, first.Description, accessJwt, endpoint)
			if err != nil {
				return "", err
			}
		}
		for _, second := range secondArray {
			if CheckIfStarterPackExists(second.Title, starterPacks) {
				fmt.Println("Starter pack with title " + second.Title + " already exists")
			} else {
				fmt.Println("Creating starter pack with title " + second.Title + " and description " + second.Description)
				_, _, err = CreateStarterPack(second.Title, second.Description, time.Now().Format("2006-01-02T15:04:05.000Z"), accessJwt, endpoint)
				if err != nil {
					return "", err
				}
			}
			if CheckIfListExists(second.Title, lists) {
				fmt.Println("List with title " + second.Title + " already exists")
			} else {
				fmt.Println("Creating List with title " + second.Title + " and description " + second.Description)
				_, err := CreateList(second.Title, second.Description, accessJwt, endpoint)
				if err != nil {
					return "", err
				}
			}
		}
	}

	return "All starter packs and lists created successfully", nil
}

func CheckIfStarterPackExists(starterPackTitle string, starterPacks []StarterPack) bool {
	for _, sp := range starterPacks {
		if sp.Record.Name == starterPackTitle {
			return true
		}
	}
	return false
}

func CheckIfListExists(listTitle string, lists []List) bool {
	for _, list := range lists {
		if list.Name == listTitle {
			return true
		}
	}
	return false
}

func GetList(listUri string, accessJwt string, endpoint string) (List, error) {
	fmt.Println("Getting list for URI " + listUri)
	url := endpoint + "/xrpc/app.bsky.graph.getList?list=" + listUri
	resp, err := SendGet(url, accessJwt)
	if err != nil {
		return List{}, err
	}

	var response ListResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return List{}, err
	}

	fmt.Println("Got list successfully")
	return response.List, nil
}

func GetLists(accessJwt string, endpoint string) ([]List, error) {
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return []List{}, err
	}
	fmt.Println("Getting lists for DID " + bskyDid)

	lists := make([]List, 0)
	hasMore := 0
	counterArg := ""
	for hasMore < 1 {
		url := endpoint + "/xrpc/app.bsky.graph.getLists?limit=100&actor=" + url.QueryEscape(bskyDid) + counterArg

		resp, err := SendGet(url, accessJwt)
		if err != nil {
			return []List{}, err
		}

		var response ListsResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return []List{}, err
		}
		lists = append(lists, response.Lists...)

		if response.Cursor != "" {
			counterArg = "&cursor=" + response.Cursor
		} else {
			hasMore = 1
		}
	}

	fmt.Printf("Got %d lists successfully\n", len(lists))
	return lists, nil
}

func CreateList(listTitle string, listDescription string, accessJwt string, endpoint string) (CreateRecordResponse, error) {
	fmt.Println("Creating list with title " + listTitle + " and description " + listDescription)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return CreateRecordResponse{}, err
	}

	url := endpoint + "/xrpc/com.atproto.repo.createRecord"

	payload := "{\"collection\": \"app.bsky.graph.list\",\"repo\": \"" + bskyDid + "\",\"record\": {\"name\": \"" + listTitle + "\",\"description\": \"" + listDescription + "\",\"createdAt\": \"" + time.Now().Format("2006-01-02T15:04:05.000Z") + "\",\"purpose\": \"app.bsky.graph.defs#curatelist\",\"$type\": \"app.bsky.graph.list\"}}"

	fmt.Println("Creating list")
	resp, err := SendPost(url, payload, accessJwt)

	var listResponse CreateRecordResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return CreateRecordResponse{}, err
	}
	if (listResponse == CreateRecordResponse{}) {
		return CreateRecordResponse{}, fmt.Errorf("Error creating list, couldn't parse JSON")
	}

	return listResponse, nil
}

func AddUserToStandaloneList(userToAddDid string, listUri string, accessJwt string, endpoint string) error {
	fmt.Println("Adding user " + userToAddDid + " to list with URI " + listUri)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return err
	}

	url := endpoint + "/xrpc/com.atproto.repo.createRecord"

	payload := "{\"collection\":\"app.bsky.graph.listitem\",\"repo\":\"" + bskyDid + "\",\"record\":{\"subject\":\"" + userToAddDid + "\",\"list\":\"" + listUri + "\",\"createdAt\":\"" + time.Now().Format("2006-01-02T15:04:05.000Z") + "\",\"$type\":\"app.bsky.graph.listitem\"}}"

	_, err = SendPost(url, payload, accessJwt)
	if err != nil {
		return err
	}

	fmt.Println("Added user to list successfully")
	return nil
}

func CreateStarterPack(starterPackTitle string, starterPackDescription string, createdAt string, accessJwt string, endpoint string) (CreateRecordResponse, CreateRecordResponse, error) {
	fmt.Println("Creating starter pack with title " + starterPackTitle + " and description " + starterPackDescription)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return CreateRecordResponse{}, CreateRecordResponse{}, err
	}

	url := endpoint + "/xrpc/com.atproto.repo.createRecord"

	payload := "{\"collection\": \"app.bsky.graph.list\",\"repo\": \"" + bskyDid + "\",\"record\": {\"name\": \"" + starterPackTitle + "\",\"description\": \"" + starterPackDescription + "\",\"createdAt\": \"" + createdAt + "\",\"purpose\": \"app.bsky.graph.defs#referencelist\",\"$type\": \"app.bsky.graph.list\"}}"

	fmt.Println("Creating list for starter pack")
	resp, err := SendPost(url, payload, accessJwt)

	var listResponse CreateRecordResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return CreateRecordResponse{}, CreateRecordResponse{}, err
	}
	if (listResponse == CreateRecordResponse{}) {
		return CreateRecordResponse{}, CreateRecordResponse{}, fmt.Errorf("Error creating list for starter pack, couldn't parse JSON")
	}

	payload = "{\"collection\": \"app.bsky.graph.starterpack\",\"repo\": \"" + bskyDid + "\",\"record\": {\"name\": \"" + starterPackTitle + "\",\"description\": \"" + starterPackDescription + "\",\"list\": \"" + listResponse.URI + "\",\"feeds\": [],\"createdAt\": \"" + createdAt + "\",\"$type\": \"app.bsky.graph.starterpack\"}}"

	fmt.Println("Making list a starter pack")
	resp, err = SendPost(url, payload, accessJwt)
	if err != nil {
		return CreateRecordResponse{}, CreateRecordResponse{}, err
	}

	var starterPackResponse CreateRecordResponse
	err = json.NewDecoder(resp.Body).Decode(&starterPackResponse)
	if err != nil {
		return CreateRecordResponse{}, CreateRecordResponse{}, err
	}
	if (starterPackResponse == CreateRecordResponse{}) {
		return CreateRecordResponse{}, CreateRecordResponse{}, fmt.Errorf("Error creating list for starter pack, couldn't parse JSON")
	}

	fmt.Println("Created starter pack successfully at " + starterPackResponse.URI + " pointing to list at " + listResponse.URI)
	return listResponse, starterPackResponse, nil
}

func AddUserToStarterPackList(userToAddDid string, listUri string, starterPackUri string, starterPackTitle string, starterPackDescription string, createdAt string, accessJwt string, endpoint string) error {
	fmt.Println("Adding user " + userToAddDid + " to list with URI " + listUri + " and starter pack with URI " + starterPackUri)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return err
	}

	url := endpoint + "/xrpc/com.atproto.repo.applyWrites"
	now := time.Now()
	timestamp := now.Format("2006-01-02T15:04:05.000Z")

	payload := "{\"repo\": \"" + bskyDid + "\",\"writes\": [{\"$type\": \"com.atproto.repo.applyWrites#create\",\"collection\": \"app.bsky.graph.listitem\",\"value\": {\"$type\": \"app.bsky.graph.listitem\",\"subject\": \"" + userToAddDid + "\",\"list\": \"" + listUri + "\",\"createdAt\": \"" + timestamp + "\"}}]}"

	_, err = SendPost(url, payload, accessJwt)
	if err != nil {
		return err
	}

	err = PutRecordForStarterPack(bskyDid, starterPackUri, starterPackDescription, starterPackTitle, createdAt, listUri, timestamp, accessJwt, endpoint)

	if err != nil {
		return err
	}

	fmt.Println("Added user to list successfully")
	return nil
}

func PutRecordForStarterPack(bskyDid string, starterPackUri string, starterPackDescription string, starterPackTitle string, createdAt string, listUri string, timestamp string, accessJwt string, endpoint string) error {
	url := endpoint + "/xrpc/com.atproto.repo.putRecord"
	rkey := starterPackUri[strings.LastIndex(starterPackUri, "/")+1:]

	payload := "{\"repo\": \"" + bskyDid + "\",\"collection\": \"app.bsky.graph.starterpack\",\"rkey\": \"" + rkey + "\",\"record\": {\"name\": \"" + starterPackTitle + "\",\"description\": \"" + starterPackDescription + "\",\"list\": \"" + listUri + "\",\"feeds\": [],\"createdAt\": \"" + createdAt + "\",\"updatedAt\": \"" + timestamp + "\"}}"

	_, err := SendPost(url, payload, accessJwt)
	if err != nil {
		return err
	}

	return nil
}

func DeleteStarterPack(rkey string, accessJwt string, endpoint string) (string, error) {
	fmt.Println("Deleting starter pack with rkey " + rkey)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return "", err
	}

	url := endpoint + "/xrpc/com.atproto.repo.deleteRecord"

	payload := "{\"repo\": \"" + bskyDid + "\",\"collection\": \"app.bsky.graph.starterpack\",\"rkey\": \"" + rkey + "\"}"

	_, err = SendPost(url, payload, accessJwt)
	if err != nil {
		return "", err
	}

	fmt.Println("Deleted starter pack successfully")
	return "Deleted starter pack successfully", nil
}

func DeleteList(rkey string, accessJwt string, endpoint string) (string, error) {
	fmt.Println("Deleting list with rkey " + rkey)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return "", err
	}

	url := endpoint + "/xrpc/com.atproto.repo.applyWrites"

	payload := "{\"repo\": \"" + bskyDid + "\",\"writes\": [{\"$type\": \"com.atproto.repo.applyWrites#delete\",\"collection\": \"app.bsky.graph.list\",\"rkey\": \"" + rkey + "\"}]}"

	_, err = SendPost(url, payload, accessJwt)
	if err != nil {
		return "", err
	}

	fmt.Println("Deleted list successfully")
	return "Deleted list successfully", nil
}

func Follow(toFollowDid string, accessJwt string, endpoint string) (string, error) {
	fmt.Println("Following user with DID " + toFollowDid)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return "", err
	}

	url := endpoint + "/xrpc/com.atproto.repo.createRecord"

	payload := "{\"collection\": \"app.bsky.graph.follow\",\"repo\": \"" + bskyDid + "\",\"record\": {\"subject\": \"" + toFollowDid + "\",\"createdAt\": \"" + time.Now().Format("2006-01-02T15:04:05.000Z") + "\",\"$type\": \"app.bsky.graph.follow\"}}"

	_, err = SendPost(url, payload, accessJwt)
	if err != nil {
		return "", err
	}

	fmt.Println("Followed user successfully")
	return "Followed user successfully", nil
}

func RespondWithStarterPacksAndListsForTitle(title string, w http.ResponseWriter, accessJwt string, endpoint string) error {
	bskyHandle, err := variables.Get("bsky_handle")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	matchingList := ListOrStarterPackWithUrl{}
	allLists, err := GetLists(accessJwt, endpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	for _, l := range allLists {
		if l.Name == title {
			matchingList = ConvertToStruct(l.URI, title, "list", bskyHandle)
			break
		}
	}

	matchingStarterPacks := []ListOrStarterPackWithUrl{}
	allStarterPacks, err := GetStarterPacks(accessJwt, endpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	for _, sp := range allStarterPacks {
		if sp.Record.Name == title {
			matchingStarterPacks = append(matchingStarterPacks, ConvertToStruct(sp.URI, title, "sp", bskyHandle))
		}
	}

	jsonResult, err := json.Marshal(ListAndStarterPacks{List: matchingList, StarterPacks: matchingStarterPacks})
	if err != nil {
		http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, string(jsonResult))
	return nil
}

func RespondWithAllStarterPacksAndListsForModule(moduleSpecifics ModuleSpecifics, w http.ResponseWriter, accessJwt string, endpoint string) error {
	naming, err := SetupFlatNamingStructure(moduleSpecifics)
	if err != nil {
		return err
	}

	jsonResult, err := json.Marshal(naming)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, string(jsonResult))
	return nil
}

func SetupAllStarterPacksAndLists(moduleSpecifics ModuleSpecifics, w http.ResponseWriter, accessJwt string, endpoint string) error {
	naming, err := SetupNamingStructure(moduleSpecifics)
	if err != nil {
		return err
	}

	result, err := CreateAllStarterPacksAndLists(naming, accessJwt, endpoint)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, result)
	return nil
}

func SetLabel(label string, targetHandle string, accessJwt string, endpoint string) error {
	fmt.Println("Adding label " + label + " to handle " + targetHandle)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return err
	}

	bskyLabelerDid, err := variables.Get("bsky_labeler_did")
	if err != nil {
		return err
	}

	targetProfile, err := GetProfile(targetHandle, accessJwt, endpoint)
	if err != nil {
		return err
	}

	additionalHeaders := map[string]string{"atproto-accept-labelers": bskyLabelerDid + ";redact", "atproto-proxy": bskyDid + "#atproto_labeler"}

	url := endpoint + "/xrpc/tools.ozone.moderation.getRepo?did=" + url.QueryEscape(targetProfile.DID)

	resp, err := SendGetWithHeader(url, accessJwt, additionalHeaders)
	if err != nil {
		return err
	}

	var response ModerationRepoResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	found := false
	for _, existingLabel := range response.Labels {
		if existingLabel.Val == label {
			found = true
			break
		}
	}

	if found {
		fmt.Println("Label already exists")
		return nil
	} else {
		url = endpoint + "/xrpc/tools.ozone.moderation.emitEvent"

		payload := "{\"subject\": {\"$type\": \"com.atproto.admin.defs#repoRef\",\"did\": \"" + targetProfile.DID + "\"},\"createdBy\": \"" + bskyDid + "\",\"subjectBlobCids\": [],\"event\": {\"$type\": \"tools.ozone.moderation.defs#modEventLabel\",\"createLabelVals\": [\"" + label + "\"],\"negateLabelVals\": []}}"

		_, err = SendPostWithHeaders(url, payload, accessJwt, additionalHeaders)
		if err != nil {
			return err
		}

		payload = "{\"subject\": {\"$type\": \"com.atproto.admin.defs#repoRef\",\"did\": \"" + targetProfile.DID + "\"},\"createdBy\": \"" + bskyDid + "\",\"subjectBlobCids\": [],\"event\": {\"$type\": \"tools.ozone.moderation.defs#modEventAcknowledge\"}}"

		_, err = SendPostWithHeaders(url, payload, accessJwt, additionalHeaders)
		if err != nil {
			return err
		}

		fmt.Println("Label added successfully")
		return nil
	}
}

func RemoveLabel(label string, targetHandle string, accessJwt string, endpoint string) error {
	fmt.Println("Removing label " + label + " from handle " + targetHandle)
	bskyDid, err := variables.Get("bsky_did")
	if err != nil {
		return err
	}

	bskyLabelerDid, err := variables.Get("bsky_labeler_did")
	if err != nil {
		return err
	}

	targetProfile, err := GetProfile(targetHandle, accessJwt, endpoint)
	if err != nil {
		return err
	}

	additionalHeaders := map[string]string{"atproto-accept-labelers": bskyLabelerDid + ";redact", "atproto-proxy": bskyDid + "#atproto_labeler"}

	requestURL := endpoint + "/xrpc/tools.ozone.moderation.getRepo?did=" + url.QueryEscape(targetProfile.DID)

	resp, err := SendGetWithHeader(requestURL, accessJwt, additionalHeaders)
	if err != nil {
		return err
	}

	var response ModerationRepoResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	found := false
	for _, existingLabel := range response.Labels {
		if existingLabel.Val == label {
			found = true
			break
		}
	}

	if !found {
		fmt.Println("Label does not exist")
		return nil
	} else {
		requestURL = endpoint + "/xrpc/tools.ozone.moderation.emitEvent"

		payload := "{\"subject\": {\"$type\": \"com.atproto.admin.defs#repoRef\",\"did\": \"" + targetProfile.DID + "\"},\"createdBy\": \"" + bskyDid + "\",\"subjectBlobCids\": [],\"event\": {\"$type\": \"tools.ozone.moderation.defs#modEventLabel\",\"createLabelVals\": [],\"negateLabelVals\": [\"" + label + "\"]}}"

		_, err = SendPostWithHeaders(requestURL, payload, accessJwt, additionalHeaders)
		if err != nil {
			return err
		}

		payload = "{\"subject\": {\"$type\": \"com.atproto.admin.defs#repoRef\",\"did\": \"" + targetProfile.DID + "\"},\"createdBy\": \"" + bskyDid + "\",\"subjectBlobCids\": [],\"event\": {\"$type\": \"tools.ozone.moderation.defs#modEventAcknowledge\"}}"

		_, err = SendPostWithHeaders(requestURL, payload, accessJwt, additionalHeaders)
		if err != nil {
			return err
		}

		fmt.Println("Label removed successfully")
		return nil
	}
}

func SendPost(url string, payload string, accessJwt string) (*http.Response, error) {
	return SendPostWithHeaders(url, payload, accessJwt, map[string]string{})
}

func SendPostWithHeaders(url string, payload string, accessJwt string, additionalHeaders map[string]string) (*http.Response, error) {
	fmt.Println("Sending POST request to " + url)
	// check if url constains login
	if strings.Contains(url, "createSession") {
		fmt.Println("Not logging the payload for login")
	} else {
		fmt.Println("Payload: " + payload)
	}
	request, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		fmt.Println("Error creating POST request: " + err.Error())
		return nil, err
	}

	if accessJwt != "" {
		request.Header.Add("Authorization", "Bearer "+accessJwt)
	}
	for key, value := range additionalHeaders {
		request.Header.Add(key, value)
	}
	request.Header.Add("Content-Type", "application/json")

	resp, err := spinhttp.Send(request)
	if err != nil {
		fmt.Println("Error sending POST request: " + err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Sprintf("The POST request returned status code %d", resp.StatusCode))
		return nil, fmt.Errorf("The POST request returned %d", resp.StatusCode)
	}
	return resp, nil
}

func SendGet(url string, accessJwt string) (*http.Response, error) {
	return SendGetWithHeaderLogConfigurable(url, accessJwt, map[string]string{}, false)
}

func SendGetLogConfigurable(url string, accessJwt string, noLog bool) (*http.Response, error) {
	return SendGetWithHeaderLogConfigurable(url, accessJwt, map[string]string{}, noLog)
}

func SendGetWithHeader(url string, accessJwt string, additionalHeaders map[string]string) (*http.Response, error) {
	return SendGetWithHeaderLogConfigurable(url, accessJwt, additionalHeaders, false)
}

func SendGetWithHeaderLogConfigurable(url string, accessJwt string, additionalHeaders map[string]string, noLog bool) (*http.Response, error) {
	if !noLog {
		fmt.Println("Sending GET request to " + url)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating GET request: " + err.Error())
		return nil, err
	}

	if accessJwt != "" {
		request.Header.Add("Authorization", "Bearer "+accessJwt)
	}
	for key, value := range additionalHeaders {
		request.Header.Add(key, value)
	}

	resp, err := spinhttp.Send(request)

	if err != nil {
		fmt.Println("Error sending GET request: " + err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Sprintf("The GET request returned status code %d", resp.StatusCode))
		return nil, fmt.Errorf("The GET request returned %d", resp.StatusCode)
	}

	return resp, nil
}
