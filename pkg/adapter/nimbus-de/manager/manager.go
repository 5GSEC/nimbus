// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"

	dspv1 "github.com/accuknox/dev2/dsp/pkg/DiscoveredPolicy/api/security.accuknox.com/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-de/watcher"
	globalwatcher "github.com/5GSEC/nimbus/pkg/adapter/watcher"
)

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
)

func init() {
	k8sClient = k8s.NewOrDie(scheme)
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(dspv1.AddToScheme(scheme))
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies,verbs=get;list;watch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies/status,verbs=get;update;
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies,verbs=get;list;watch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies/status,verbs=get;update;
//+kubebuilder:rbac:groups=security.accuknox.com,resources=discoveredpolicies,verbs=list;watch;update;

func Run(ctx context.Context) {
	npCh := make(chan common.Request)
	deletedNpCh := make(chan *unstructured.Unstructured)
	go globalwatcher.WatchNimbusPolicies(ctx, npCh, deletedNpCh, "SecurityIntentBinding")

	dspCh := make(chan common.Request)
	go watcher.WatchDsps(ctx, dspCh)

	for {
		select {
		case <-ctx.Done():
			close(npCh)
			close(deletedNpCh)

			close(dspCh)
			return
		case createdNp := <-npCh:
			activateDspsBasedOnNp(ctx, createdNp.Name, createdNp.Namespace)
		case deletedNp := <-deletedNpCh:
			deactivateDspsOnNp(ctx, deletedNp)

		case dsp := <-dspCh:
			matchAndActivateDsp(ctx, dsp.Namespace)
		}
	}
}
