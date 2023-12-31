// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
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
		body, err := io.ReadAll(r.Body)
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

	// Create a custom HTTP server with timeouts
	server := &http.Server{
		Addr:         ":13000",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Start the server
	log.Println("Server starting on port 13000...")
	log.Fatal(server.ListenAndServe())
}
