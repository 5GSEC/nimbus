// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"reflect"

	dspv1 "github.com/accuknox/dev2/dsp/pkg/DiscoveredPolicy/api/security.accuknox.com/v1"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/go-logr/logr"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

const (
	StatusActiveDsp   = "Active"
	StatusInactiveDsp = "Inactive"
)

var (
	decoder = k8sscheme.Codecs.UniversalDeserializer()
)

func activateDspsBasedOnNp(ctx context.Context, nimbusPolicyName, namespace string) {
	logger := log.FromContext(ctx)

	var nimbusPolicy v1alpha1.NimbusPolicy
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: nimbusPolicyName, Namespace: namespace}, &nimbusPolicy); err != nil {
		logger.Error(err, "failed to get the NimbusPolicy", "NimbusPolicy.Name", nimbusPolicyName)
		return
	}

	if isNetSegmentId(logger, nimbusPolicy.Spec.NimbusRules) {
		activateOrDeactivateDsps(ctx, nimbusPolicy, true)
	} else {
		deleteDanglingPoliciesIfExist(ctx, nimbusPolicy)
	}
}

func deleteDanglingPoliciesIfExist(ctx context.Context, nimbusPolicy v1alpha1.NimbusPolicy) {
	dsps := getDsps(ctx, nimbusPolicy.Namespace)

	// Iterate using a separate index variable to avoid aliasing
	for idx := range dsps.Items {
		dsp := &dsps.Items[idx]
		// If a nimbusPolicy does not contain the "netSegment" intentID, and if a DSP has
		// a "part-of" label equal to the nimbusPolicy.Name, that DSP needs to be
		// deactivated. This is because the referenced nimbusPolicy has been modified to
		// remove the network segmentation intent.
		if dsp.Labels["app.kubernetes.io/part-of"] == nimbusPolicy.Name && dsp.Spec.PolicyStatus == StatusActiveDsp {
			activateOrDeactivateDsps(ctx, nimbusPolicy, false)
		}
	}
}

func deactivateDspsOnNp(ctx context.Context, deletedNp *unstructured.Unstructured) {
	logger := log.FromContext(ctx)

	var nimbusPolicy v1alpha1.NimbusPolicy
	bytes, err := deletedNp.MarshalJSON()
	if err != nil {
		logger.Error(err, "failed to marshal deleted nimbusPolicy")
		return
	}

	_, _, err = decoder.Decode(bytes, nil, &nimbusPolicy)
	if err != nil {
		logger.Error(err, "failed to decode deleted nimbusPolicy")
		return
	}

	if isNetSegmentId(logger, nimbusPolicy.Spec.NimbusRules) {
		activateOrDeactivateDsps(ctx, nimbusPolicy, false)
	}
}

func matchAndActivateDsp(ctx context.Context, namespace string) {
	logger := log.FromContext(ctx)

	var nimbusPolicies v1alpha1.NimbusPolicyList
	if err := k8sClient.List(ctx, &nimbusPolicies, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "failed to list NimbusPolicies", "namespace", namespace)
		return
	}

	for _, nimbusPolicy := range nimbusPolicies.Items {
		activateDspsBasedOnNp(ctx, nimbusPolicy.Name, nimbusPolicy.Namespace)
	}
}

func activateOrDeactivateDsps(ctx context.Context, nimbusPolicy v1alpha1.NimbusPolicy, activate bool) {
	logger := log.FromContext(ctx)

	dsps := getDsps(ctx, nimbusPolicy.Namespace)
	// Iterate using a separate index variable to avoid aliasing
	for idx := range dsps.Items {
		dsp := &dsps.Items[idx]

		if activate && dsp.Spec.PolicyStatus == StatusActiveDsp {
			continue
		}
		if !activate && dsp.Spec.PolicyStatus == StatusInactiveDsp {
			continue
		}

		_, gvk, err := unstructured.UnstructuredJSONScheme.Decode(dsp.Spec.Policy.Raw, nil, nil)
		if err != nil {
			logger.Error(err, "failed to decode DiscoveredPolicy", "discoveredPolicy.name", dsp.Name, "discoveredPolicy.namespace", dsp.Namespace)
			return
		}

		var currPolicyFullName string
		var needToUpdate bool
		switch gvk.Kind {
		case "KubeArmorPolicy":
			var ksp kubearmorv1.KubeArmorPolicy
			_, _, err = decoder.Decode(dsp.Spec.Policy.Raw, nil, &ksp)
			if err != nil {
				logger.Error(err, "failed to decode KubeArmorPolicy", "discoveredPolicy.name", dsp.Name)
				return
			}
			if reflect.DeepEqual(ksp.Spec.Selector.MatchLabels, nimbusPolicy.Spec.Selector.MatchLabels) {
				if activate {
					dsp.Spec.PolicyStatus = StatusActiveDsp
				} else {
					dsp.Spec.PolicyStatus = StatusInactiveDsp
				}
				currPolicyFullName = ksp.Kind + "/" + ksp.Name
				needToUpdate = true
			}

		case "NetworkPolicy":
			var networkPolicy netv1.NetworkPolicy
			_, _, err = decoder.Decode(dsp.Spec.Policy.Raw, nil, &networkPolicy)
			if err != nil {
				logger.Error(err, "failed to decode NetworkPolicy", "discoveredPolicy.name", dsp.Name)
				return
			}

			if reflect.DeepEqual(networkPolicy.Spec.PodSelector.MatchLabels, nimbusPolicy.Spec.Selector.MatchLabels) {
				if activate {
					dsp.Spec.PolicyStatus = StatusActiveDsp
				} else {
					dsp.Spec.PolicyStatus = StatusInactiveDsp
				}
				currPolicyFullName = networkPolicy.Kind + "/" + networkPolicy.Name
				needToUpdate = true
			}

		case "CiliumNetworkPolicy":
			var ciliumNetworkPolicy ciliumv2.CiliumNetworkPolicy
			_, _, err = decoder.Decode(dsp.Spec.Policy.Raw, nil, &ciliumNetworkPolicy)
			if err != nil {
				logger.Error(err, "failed to decode CiliumNetworkPolicy", "discoveredPolicy.name", dsp.Name)
				return
			}

			if reflect.DeepEqual(ciliumNetworkPolicy.Spec.EndpointSelector.LabelSelector.MatchLabels, nimbusPolicy.Spec.Selector.MatchLabels) {
				if activate {
					dsp.Spec.PolicyStatus = StatusActiveDsp
				} else {
					dsp.Spec.PolicyStatus = StatusInactiveDsp
				}
				currPolicyFullName = ciliumNetworkPolicy.Kind + "/" + ciliumNetworkPolicy.Name
				needToUpdate = true
			}
		}

		if needToUpdate {
			manageLabels(dsp, nimbusPolicy.Name, activate)
			if err = updateDsp(ctx, dsp); err != nil {
				logger.Error(err, "failed to update discoveredPolicy", "discoveredPolicy.name", dsp.Name, "discoveredPolicy.namespace", dsp.Namespace)
			} else {
				// Update nimbusPolicy status based on activate flag:

				// The `adapterutil.UpdateNpStatus` function takes a boolean argument named
				// "decrement" alongside others. Here's how we determine the value to pass:
				//   - If `activate` is true, we negate it to pass `false` as decrement,
				//     indicating an increment for the policies count.
				//   - Otherwise (if `activate` is false), we pass `true` as decrement,
				//     indicating a decrement for the policies count.
				err = adapterutil.UpdateNpStatus(ctx, k8sClient, currPolicyFullName, nimbusPolicy.Name, nimbusPolicy.Namespace, !activate)
				if err != nil {
					logger.Error(err, "failed to update nimbusPolicy status", "nimbusPolicy.name", nimbusPolicy.Name, "nimbusPolicy.namespace", nimbusPolicy.Namespace)
				}
			}
		}
	}
}

func manageLabels(dsp *dspv1.DiscoveredPolicy, nimbusPolicyName string, activate bool) {
	if dsp.ObjectMeta.Labels == nil {
		dsp.ObjectMeta.Labels = map[string]string{}
	}

	// These labels will help in filtering DSPs during deletion of dangling DSPs when
	// a nimbusPolicy is edited.
	if activate {
		dsp.ObjectMeta.Labels["app.kubernetes.io/managed-by"] = "nimbus-de"
		dsp.ObjectMeta.Labels["app.kubernetes.io/part-of"] = nimbusPolicyName
	} else {
		delete(dsp.ObjectMeta.Labels, "app.kubernetes.io/managed-by")
		delete(dsp.ObjectMeta.Labels, "app.kubernetes.io/part-of")
	}
}

func updateDsp(ctx context.Context, dsp *dspv1.DiscoveredPolicy) error {
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestDsp := dspv1.DiscoveredPolicy{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: dsp.Name, Namespace: dsp.Namespace}, &latestDsp); err != nil {
			return err
		}

		latestDsp.Labels = dsp.Labels
		latestDsp.Spec.PolicyStatus = dsp.Spec.PolicyStatus
		if err := k8sClient.Update(ctx, &latestDsp); err != nil {
			return err
		}

		return nil
	}); retryErr != nil {
		return retryErr
	}
	return nil
}

func getDsps(ctx context.Context, namespace string) dspv1.DiscoveredPolicyList {
	logger := log.FromContext(ctx)

	var dsps dspv1.DiscoveredPolicyList
	if err := k8sClient.List(ctx, &dsps, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "failed to get the DSPs", "namespace", namespace)
		return dspv1.DiscoveredPolicyList{}
	}

	return dsps
}

func isNetSegmentId(logger logr.Logger, nimbusRules []v1alpha1.NimbusRules) bool {
	for _, nimbusRule := range nimbusRules {
		if nimbusRule.ID == idpool.NetworkSegmentation {
			return true
		}
		logger.Info("Discovery engine doesn't support this ID", "ID", nimbusRule.ID)
	}
	return false
}
