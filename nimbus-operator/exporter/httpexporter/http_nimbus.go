// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

// Initialize HTTP Client: Set up the client for HTTP communication.
// Format Nimbus Policy Data: Convert Nimbus Policy data into JSON format.
// Send Data: Send the converted data to the adapter's URL using a POST request.
// Process Response: Handle the response from the adapter and log as necessary.

package httpexporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	v1 "github.com/5GSEC/nimbus/nimbus-operator/api/v1"
)

// HttpNimbusExporter struct defines the HTTP client and the URL for exporting Nimbus policies.
type HttpNimbusExporter struct {
	client *http.Client
	url    string
}

// NewHttpNimbusExporter creates a new HttpNimbusExporter with the provided URL.
func NewHttpNimbusExporter(url string) *HttpNimbusExporter {
	return &HttpNimbusExporter{
		client: &http.Client{},
		url:    url,
	}
}

// ExportNimbusPolicy exports a NimbusPolicy to a remote server via HTTP POST.
func (h *HttpNimbusExporter) ExportNimbusPolicy(ctx context.Context, policy *v1.NimbusPolicy) error {
	// Convert the NimbusPolicy into JSON format.
	data, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal NimbusPolicy: %v", err)
	}

	// Create a new HTTP POST request with the policy data.
	req, err := http.NewRequestWithContext(ctx, "POST", h.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request to the server.
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK (HTTP 200).
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK response received: %v", resp.Status)
	}

	return nil
}
