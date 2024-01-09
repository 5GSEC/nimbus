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
	"github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/core/enforcer"
	watcher "github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/receiver/nimbuspolicywatcher"
	"github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/receiver/verifier"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
)

// Initialize the global scheme variable
var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(kubearmorv1.AddToScheme(scheme))
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	log.Println("Starting Kubernetes client configuration")

	var cfg *rest.Config
	var err error
	if cfg, err = rest.InClusterConfig(); err != nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Failed to set up Kubernetes config: %v", err)
		}
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	log.Println("Starting NimbusPolicyWatcher")
	npw := watcher.NewNimbusPolicyWatcher(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	policyChan, err := npw.WatchNimbusPolicies(ctx)
	if err != nil {
		log.Fatalf("NimbusPolicy: Watch Failed %v", err)
	}

	detectedPolicies := make(map[string]bool)
	enforcer := enforcer.NewPolicyEnforcer(c)

	log.Println("Starting policy processing loop")
	for {
		select {
		case policy := <-policyChan:
			policyKey := fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)
			if _, detected := detectedPolicies[policyKey]; !detected {
				if verifier.HandlePolicy(policy) {
					log.Printf("NimbusPolicy: Detected policy: Name: %s, Namespace: %s, ID: %s \n%+v\n", policy.Namespace, policy.Name, getRulesIDs(policy), policy)
					detectedPolicies[policyKey] = true

					log.Println("Exporting and Applying NimbusPolicy to KubeArmorPolicy")
					err := enforcer.Enforcer(ctx, policy)
					if err != nil {
						log.Printf("Error exporting NimbusPolicy: %v", err)
					} else {
						log.Println("Successfully exported NimbusPolicy to KubeArmorPolicy")
					}
				}
			}
		case <-time.After(120 * time.Second):
			log.Println("NimbusPolicy: No detections for 120 seconds")
		}
	}
}

func getRulesIDs(policy v1.NimbusPolicy) string {
	var ruleIDs []string
	for _, rule := range policy.Spec.NimbusRules {
		ruleIDs = append(ruleIDs, rule.Id)
	}
	return fmt.Sprintf("[%s]", strings.Join(ruleIDs, ", "))
}
