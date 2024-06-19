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

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

func BuildNetPolsFrom(logger logr.Logger, np v1alpha1.NimbusPolicy) []netv1.NetworkPolicy {
	// Build netpols based on given IDs
	var netpols []netv1.NetworkPolicy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		logger.Info(id)
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
	case idpool.DenyENAccess:
		return denyExternalNetworkAcessNetpol()
	default:
		return netv1.NetworkPolicy{}
	}
}

func denyExternalNetworkAcessNetpol() netv1.NetworkPolicy {
	udpProtocol := corev1.ProtocolUDP
	tcpProtocol := corev1.ProtocolTCP
	dnsPort := &intstr.IntOrString{
		Type:   0,
		IntVal: 53,
	}

	return netv1.NetworkPolicy{
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "10.0.0.0/8",
							},
						},
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "172.16.0.0/12",
							},
						},
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "192.168.0.0/16",
							},
						},
					},
				},
			},
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

						{
							IPBlock: &netv1.IPBlock{
								CIDR: "10.0.0.0/8",
							},
						},
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "172.16.0.0/12",
							},
						},
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "192.168.0.0/16",
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
				netv1.PolicyTypeIngress,
			},
		},
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
