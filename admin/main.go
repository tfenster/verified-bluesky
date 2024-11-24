package main

import (
	"fmt"
	"net/http"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/variables"
	"github.com/shared"
)

func init() {

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

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
