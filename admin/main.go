package main

import (
	"fmt"
	"net/http"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/kv"
	"github.com/fermyon/spin/sdk/go/v2/variables"
	"github.com/shared"
)

func init() {

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPut:
			// ONLY WORKS FOR MS MVPS AND RDS!
			adminMode, err := variables.Get("admin_mode")
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if adminMode != "true" {
				http.Error(w, "admin mode not enabled", http.StatusUnauthorized)
				return
			}

			fmt.Println("Setting labels")
			accessJwt, endpoint, err := shared.LoginToBskyWithReq(r)
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
			for _, key := range keys {
				value, err := store.Get(key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				moduleKey := strings.Split(key, "-")[0]
				shared.SetLabel("ms-" + moduleKey, string(value), accessJwt, endpoint)
			}

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
