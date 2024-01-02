package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var (
	// Memory store for saved Nimbus Policies
	nimbusPolicies []interface{}
	lock           sync.Mutex
)

func main() {
	// Handler for exporting Nimbus Policies
	http.HandleFunc("/api/v1/nimbus/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Unmarshal the JSON data from the request
		var data interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}

		// Store the received Nimbus Policy
		lock.Lock()
		nimbusPolicies = append(nimbusPolicies, data)
		lock.Unlock()

		// Log the received policy
		fmt.Printf("Received Nimbus Policy: %+v\n", data)
		w.WriteHeader(http.StatusOK)
	})

	// Handler for retrieving stored Nimbus Policies
	http.HandleFunc("/api/v1/nimbus/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET method is accepted", http.StatusMethodNotAllowed)
			return
		}

		lock.Lock()
		defer lock.Unlock()
		// Encode and respond with the stored policies
		if err := json.NewEncoder(w).Encode(nimbusPolicies); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	})

	// Start the server on port 13000
	log.Println("Server starting on port 13000...")
	log.Fatal(http.ListenAndServe(":13000", nil))
}
