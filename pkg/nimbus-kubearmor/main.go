// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/5GSEC/nimbus/api/v1"
	transformer "github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/core/transformer"
	watcher "github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/receiver/nimbuspolicywatcher"
	"github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/receiver/verifier"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
)

// Initialize the global scheme variable
var scheme = runtime.NewScheme()

func init() {
	// Register the NimbusPolicy type in the schema
	utilruntime.Must(v1.AddToScheme(scheme))
	// Register the KubeArmorPolicy type in the schema
	utilruntime.Must(kubearmorv1.AddToScheme(scheme))
}

func main() {
	log.Println("Starting Kubernetes client configuration")
	// Set up the Kubernetes client configuration
	var cfg *rest.Config
	var err error

	// Check if running inside the cluster
	if cfg, err = rest.InClusterConfig(); err != nil {
		// If running outside the cluster, use kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Failed to set up Kubernetes config: %v", err)
		}
	}

	// Create the client with the specified scheme
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	log.Println("Starting NimbusPolicyWatcher")
	// Initialize the NimbusPolicyWatcher
	npw := watcher.NewNimbusPolicyWatcher(c)

	// Start watching NimbusPolicies
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	policyChan, err := npw.WatchNimbusPolicies(ctx)
	if err != nil {
		log.Fatalf("NimbusPolicy: Watch Failed %v", err)
	}

	// Keep track of detected policies to prevent duplicate printing
	detectedPolicies := make(map[string]bool)

	// Initialize the PolicyTransformer
	pt := transformer.NewPolicyTransformer(c)

	log.Println("Starting policy processing loop")
	// Process received NimbusPolicies
	for {
		select {
		case policy := <-policyChan:
			policyKey := fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)
			if _, detected := detectedPolicies[policyKey]; !detected {
				if verifier.HandlePolicy(policy) {
					log.Printf("NimbusPolicy: Detected policy: Name: %s, Namespace: %s, ID: %s \n%+v\n", policy.Namespace, policy.Name, getRulesIDs(policy), policy)
				}
				// Mark the policy as detected
				detectedPolicies[policyKey] = true

				// Transform NimbusPolicy to KubeArmorPolicy
				log.Println("Transforming NimbusPolicy to KubeArmorPolicy")
				kubeArmorPolicy, err := pt.Transform(ctx, policy)
				if err != nil {
					log.Printf("Error transforming NimbusPolicy: %v", err)
					continue
				}
				// Log the transformed KubeArmorPolicy
				log.Printf("Transformed KubeArmorPolicy: %+v\n", kubeArmorPolicy)

				// Apply or update the KubeArmorPolicy
				log.Println("Applying KubeArmorPolicy")
				err = pt.ApplyPolicy(ctx, kubeArmorPolicy)
				if err != nil {
					log.Printf("Error applying/updating KubeArmorPolicy: %v", err)
					continue
				} else {
					log.Println("Successfully applied/updated KubeArmorPolicy")
				}
			}
		case <-time.After(120 * time.Second):
			log.Println("NimbusPolicy: No detections for 120 seconds")
		}
	}
}

// Return a list of Rules IDs from the NimbusPolicy as a string
func getRulesIDs(policy v1.NimbusPolicy) string {
	var ruleIDs []string
	for _, rule := range policy.Spec.NimbusRules {
		ruleIDs = append(ruleIDs, rule.Id)
	}
	return fmt.Sprintf("[%s]", strings.Join(ruleIDs, ", "))
}
