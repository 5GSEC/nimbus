package controller

import (
	"context"
	"time"

	nimbusv1 "github.com/5GSEC/nimbus/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	IntentName        = "dns-manipulation"
	IntentBindingName = "dns-manipulation-binding"
)

var _ = It("should create the NimbusPolicy automatically", func() {

	si := nimbusv1.SecurityIntent{
		TypeMeta: v1.TypeMeta{
			Kind:       "SecurityIntent",
			APIVersion: "intent.security.nimbus.com/v1",
		},

		ObjectMeta: v1.ObjectMeta{
			Name:      IntentName,
			Namespace: "default",
		},

		Spec: nimbusv1.SecurityIntentSpec{
			Intent: nimbusv1.Intent{
				ID:          "dnsManipulation",
				Description: "An adversary can manipulate DNS requests to redirect network traffic and potentially reveal end user activity.",
				Action:      "Block",
			},
		},
	}

	sib := nimbusv1.SecurityIntentBinding{

		TypeMeta: v1.TypeMeta{
			Kind:       "SecurityIntentBinding",
			APIVersion: "intent.security.nimbus.com/v1",
		},

		ObjectMeta: v1.ObjectMeta{
			Name:      IntentBindingName,
			Namespace: "default",
		},

		Spec: nimbusv1.SecurityIntentBindingSpec{
			Intents: []nimbusv1.MatchIntent{
				{
					Name: IntentName,
				},
			},

			Selector: nimbusv1.Selector{
				Any: []nimbusv1.ResourceFilter{
					{
						Resources: nimbusv1.Resources{
							Kind:      "Pod",
							Namespace: "default",
							MatchLabels: map[string]string{
								"app": "nginx",
							},
						},
					},
				},
			},
		},
	}

	err := k8sClient.Create(context.TODO(), &si, &client.CreateOptions{})

	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(context.TODO(), &sib, &client.CreateOptions{})

	Expect(err).NotTo(HaveOccurred())

	// Wait for the NimbusPolicy to be created
	waitForIntentBinding(k8sClient, "default", IntentBindingName)
	waitForNimbusPolicy(k8sClient, "default", IntentBindingName)

	matchLabels := map[string]string{
		"app": "nginx",
	}

	nimbusRulesList := []nimbusv1.NimbusRules{
		{
			ID:          "dnsManipulation",
			Description: "An adversary can manipulate DNS requests to redirect network traffic and potentially reveal end user activity.",
			Rule: nimbusv1.Rule{
				RuleAction: "Block",
				Mode:       "best-effort",
			},
		},
	}

	policy := &nimbusv1.NimbusPolicy{
		TypeMeta: v1.TypeMeta{
			Kind:       "NimbusPolicy",
			APIVersion: "intent.security.nimbus.com/v1",
		},

		ObjectMeta: v1.ObjectMeta{
			Name: IntentBindingName,
		},

		Spec: nimbusv1.NimbusPolicySpec{
			Selector: nimbusv1.NimbusSelector{
				MatchLabels: matchLabels,
			},
			NimbusRules: nimbusRulesList,
		},
	}

	nimbusPolicy := nimbusv1.NimbusPolicy{}

	// Retrieve the NimbusPolicy
	err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: "dns-manipulation-binding", Namespace: "default"}, &nimbusPolicy, &client.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	// Assert the content of the NimbusPolicy
	Expect(nimbusPolicy.Spec).To(Equal(policy.Spec))

})

func waitForNimbusPolicy(clientset client.Client, namespace, policyName string) {
	nimbusPolicy := nimbusv1.NimbusPolicy{}
	Eventually(func() bool {
		err := clientset.Get(context.TODO(), types.NamespacedName{Name: IntentBindingName, Namespace: "default"}, &nimbusPolicy, &client.GetOptions{})

		return err == nil
	}, time.Second*10).Should(BeTrue(), "NimbusPolicy not created")

	// clientset.List(context.TODO(), &nimbusPolicy, &client.ListOptions{Namespace: "default"})

	// fmt.Println(nimbusPolicy.Items)
}

func waitForIntentBinding(clientset client.Client, namespace, policyName string) {
	sib := nimbusv1.SecurityIntentBinding{}
	Eventually(func() bool {
		err := clientset.Get(context.TODO(), types.NamespacedName{Name: IntentBindingName, Namespace: "default"}, &sib, &client.GetOptions{})
		return err == nil
	}, time.Second*5).Should(BeTrue(), "sib not created")
}

// var _ = AfterSuite(func() {
// 	By("tearing down the test environment")
// 	err := testEnv.Stop()
// 	Expect(err).NotTo(HaveOccurred())
// })
