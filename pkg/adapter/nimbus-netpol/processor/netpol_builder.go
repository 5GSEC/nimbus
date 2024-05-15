// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

func BuildNetPolsFrom(logger logr.Logger, np v1.NimbusPolicy) []netv1.NetworkPolicy {
	// Build netpols based on given IDs
	var netpols []netv1.NetworkPolicy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "netpol") {
			netpol := buildNetPolFor(id)
			netpol.Name = np.Name + "-" + strings.ToLower(id)
			netpol.Namespace = np.Namespace
			netpol.Spec.PodSelector.MatchLabels = np.Spec.Selector.MatchLabels
			addManagedByAnnotation(&netpol)
			netpols = append(netpols, netpol)
		} else {
			logger.Info("Network Policy adapter does not support this ID", "ID", id,
				"NimbusPolicy.Name", np.Name, "NimbusPolicy.Namespace", np.Namespace)
		}
	}
	return netpols
}

func buildNetPolFor(id string) netv1.NetworkPolicy {
	switch id {
	case idpool.DNSManipulation:
		return dnsManipulationNetpol()
	default:
		return netv1.NetworkPolicy{}
	}
}

func dnsManipulationNetpol() netv1.NetworkPolicy {
	udpProtocol := corev1.ProtocolUDP
	tcpProtocol := corev1.ProtocolTCP
	dnsPort := &intstr.IntOrString{
		Type:   0,
		IntVal: 53,
	}

	return netv1.NetworkPolicy{
		Spec: netv1.NetworkPolicySpec{
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"k8s-app": "kube-dns",
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "kube-system",
								},
							},
						},
					},
					Ports: []netv1.NetworkPolicyPort{
						{
							Protocol: &udpProtocol,
							Port:     dnsPort,
						},
						{
							Protocol: &tcpProtocol,
							Port:     dnsPort,
						},
					},
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeEgress,
			},
		},
	}
}

func addManagedByAnnotation(netpol *netv1.NetworkPolicy) {
	netpol.Annotations = make(map[string]string)
	netpol.Annotations["app.kubernetes.io/managed-by"] = "nimbus-netpol"
}
