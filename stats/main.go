package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/kv"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Println("Getting stats")
			store, err := kv.OpenStore("default")
			if err != nil {
				http.Error(w, "Error opening kv store "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer store.Close()

			fmt.Println("Getting keys")
			keys, err := store.GetKeys()
			if err != nil {
				http.Error(w, "Error getting keys from kv store "+err.Error(), http.StatusInternalServerError)
				return
			}

			typeCounters := make(map[string]int)

			fmt.Println("Counting")
			for _, key := range keys {
				keyPrefix := strings.Split(key, "-")[0]
				if (keyPrefix == "accessJwt") || (keyPrefix == "endpoint") {
					continue
				}
				if _, ok := typeCounters[keyPrefix]; !ok {
					typeCounters[keyPrefix] = 1
				} else {
					typeCounters[keyPrefix]++
				}
			}

			fmt.Println("Replying")
			jsonResult, err := json.Marshal(typeCounters)
			if err != nil {
				http.Error(w, "Error encoding result to JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, string(jsonResult))

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func main() {}
