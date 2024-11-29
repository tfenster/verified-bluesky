package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/kv"
	"github.com/fermyon/spin/sdk/go/v2/variables"
	"github.com/shared"
)

type KVEntry struct {
	Key   string
	Value string
}

func init() {

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPut:
			// FIXME: THIS ONLY WORKS FOR RDS AND MVPS!
			adminMode, err := variables.Get("admin_mode")
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if adminMode != "true" {
				http.Error(w, "admin mode not enabled", http.StatusUnauthorized)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			var kvEntry KVEntry
			err = json.Unmarshal(body, &kvEntry)
			if err != nil {
				http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			accessJwt, endpoint, err := shared.LoginToBsky()
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			fmt.Println("Setting label for " + kvEntry.Key + " and " + kvEntry.Value)

			store, err := kv.OpenStore("default")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer store.Close()

			valueFromStore, err := store.Get(kvEntry.Key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if string(valueFromStore) != kvEntry.Value {
				http.Error(w, "Value does not match", http.StatusInternalServerError)
				return
			}
			
			moduleKey := strings.Split(kvEntry.Key, "-")[0]
			err = shared.SetLabel("ms-" + moduleKey, string(kvEntry.Value), accessJwt, endpoint)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, "Label ms-" + moduleKey + " set for " + kvEntry.Value)

		case http.MethodGet:
			adminMode, err := variables.Get("admin_mode")
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if adminMode != "true" {
				http.Error(w, "admin mode not enabled", http.StatusUnauthorized)
				return
			}

			fmt.Println("Getting KV entries")
			_, _, err = shared.LoginToBskyWithReq(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			store, err := kv.OpenStore("default")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer store.Close()

			keys, err := store.GetKeys()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			kvEntries := make([]KVEntry, len(keys))
			for _, key := range keys {
				value, err := store.Get(key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				kvEntries = append(kvEntries, KVEntry{key, string(value)})
			}

			jsonResult, err := json.Marshal(kvEntries)
			if err != nil {
				http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, string(jsonResult))

		case http.MethodDelete:
			adminMode, err := variables.Get("admin_mode")
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if adminMode != "true" {
				http.Error(w, "admin mode not enabled", http.StatusUnauthorized)
				return
			}

			fmt.Println("Deleting all Starter Packs and Lists")
			accessJwt, endpoint, err := shared.LoginToBskyWithReq(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}

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

func main() {}
