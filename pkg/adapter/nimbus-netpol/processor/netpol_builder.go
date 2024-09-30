// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"context"
	"strings"

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildNetPolsFrom(logger logr.Logger, np v1alpha1.NimbusPolicy, k8sClient client.Client) []netv1.NetworkPolicy {
	// Build netpols based on given IDs
	var netpols []netv1.NetworkPolicy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		logger.Info(id)
		if idpool.IsIdSupportedBy(id, "netpol") {
			netpol := buildNetPolFor(id, k8sClient, logger)
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

func buildNetPolFor(id string, k8sClient client.Client, logger logr.Logger) netv1.NetworkPolicy {
	switch id {
	case idpool.DNSManipulation:
		return dnsManipulationNetpol(k8sClient, logger)
	case idpool.DenyENAccess:
		return denyExternalNetworkAcessNetpol(k8sClient, logger)
	default:
		return netv1.NetworkPolicy{}
	}
}

func denyExternalNetworkAcessNetpol(k8sClient client.Client, logger logr.Logger) netv1.NetworkPolicy {
	udpProtocol := corev1.ProtocolUDP
	tcpProtocol := corev1.ProtocolTCP
	dnsPort := &intstr.IntOrString{
		Type:   0,
		IntVal: 53,
	}
	froNetpolPeers, err := getPODCIDRs(k8sClient)
	if err != nil {
		logger.Error(err, "Failed to get pod CIDRs")
	}
	staticCIDRs := []netv1.NetworkPolicyPeer{
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
	}

	froNetpolPeers = append(froNetpolPeers, staticCIDRs...)

	toNetPolPeers := []netv1.NetworkPolicyPeer{}

	selector := netv1.NetworkPolicyPeer{
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
	}

	toNetPolPeers = append(toNetPolPeers, selector)
	toNetPolPeers = append(toNetPolPeers, froNetpolPeers...)

	return netv1.NetworkPolicy{
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: froNetpolPeers,
				},
			},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: toNetPolPeers,
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

func dnsManipulationNetpol(k8sClient client.Client, logger logr.Logger) netv1.NetworkPolicy {
	udpProtocol := corev1.ProtocolUDP
	tcpProtocol := corev1.ProtocolTCP
	dnsPort := &intstr.IntOrString{
		Type:   0,
		IntVal: 53,
	}

	netpolPeers, err := getPODCIDRs(k8sClient)
	if err != nil {
		logger.Error(err, "Failed to get pod CIDRs")
	}

	selector := netv1.NetworkPolicyPeer{
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
	}

	netpolPeers = append(netpolPeers, selector)

	return netv1.NetworkPolicy{
		Spec: netv1.NetworkPolicySpec{
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: netpolPeers,
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

func getPODCIDRs(k8sClient client.Client) ([]netv1.NetworkPolicyPeer, error) {
	podCIDRs := []netv1.NetworkPolicyPeer{}
	ctx := context.Background()
	nodes := &corev1.NodeList{}
	if err := k8sClient.List(ctx, nodes); err != nil {
		return nil, err
	}
	for _, node := range nodes.Items {
		netPolPeer := netv1.NetworkPolicyPeer{
			IPBlock: &netv1.IPBlock{
				CIDR: node.Spec.PodCIDR,
			},
		}

		podCIDRs = append(podCIDRs, netPolPeer)

	}

	return podCIDRs, nil
}
