package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

			kvEntries := make([]KVEntry, 0)
			for _, key := range keys {
				if (key == "endpoint") || (key == "accessJwt") || (key == "") {
					continue
				}
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
		
		case http.MethodPost:
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

			var backendUrl string
			err = json.Unmarshal(body, &backendUrl)
			if err != nil {
				http.Error(w, "Error decoding body JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			resp, err := shared.SendGet(backendUrl, "")
			if err != nil {
				http.Error(w, "Error getting data from backend URL " + backendUrl + ": " + err.Error(), http.StatusInternalServerError)
				return
			}

			var kvEntries []KVEntry
			err = json.NewDecoder(resp.Body).Decode(&kvEntries)
			if err != nil {
				http.Error(w, "Error decoding backend JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			store, err := kv.OpenStore("default")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer store.Close()

			for _, entry := range kvEntries {
				fmt.Println("Setting KV entry: <"+entry.Key+"> = <"+entry.Value+">")
				err = store.Set(entry.Key, []byte(entry.Value))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func main() {}
