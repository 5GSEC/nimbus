// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter/exporter"
)

var (
	// Memory store for saved Nimbus Policies
	nimbusPolicies []interface{}
	lock           sync.Mutex
)

func main() {
	logger, _ := zap.NewProduction()
	log := logger.Sugar()

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
		var nimbusPolicy v1.NimbusPolicy
		err = json.Unmarshal(body, &nimbusPolicy)
		if err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}

		// Store the received Nimbus Policy
		lock.Lock()
		nimbusPolicies = append(nimbusPolicies, nimbusPolicy)
		lock.Unlock()
		log.Infof("Exporting '%s' NimbusPolicy to security engines", nimbusPolicy.Name)
		exporter.ExportNpToAdapters(log, nimbusPolicy)
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
	log.Info("Starting server on port 13000...")
	log.Fatal(server.ListenAndServe())
}
