package nimbus_test

// import (
// 	"context"
// 	"fmt"
// 	"path/filepath"
// 	"runtime"
// 	"testing"
// 	"time"

// 	nimbusv1 "github.com/5GSEC/nimbus/api/v1"
// 	controllers "github.com/5GSEC/nimbus/internal/controller"
// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// 	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/client-go/kubernetes/scheme"
// 	"k8s.io/client-go/rest"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/envtest"
// 	logf "sigs.k8s.io/controller-runtime/pkg/log"
// 	"sigs.k8s.io/controller-runtime/pkg/log/zap"
// )

// const (
// 	IntentName        = "dns-manipulation"
// 	IntentBindingName = "dns-manipulation-binding"
// )

// var cfg *rest.Config
// var k8sClient client.Client
// var testEnv *envtest.Environment

// func TestNimbus(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "Nimbus Suite")
// }

// var _ = BeforeSuite(func() {
// 	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

// 	By("bootstrapping test environment")
// 	testEnv = &envtest.Environment{
// 		CRDDirectoryPaths:     []string{filepath.Join("config", "crd", "bases")},
// 		ErrorIfCRDPathMissing: true,

// 		// The BinaryAssetsDirectory is only required if you want to run the tests directly
// 		// without call the makefile target test. If not informed it will look for the
// 		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
// 		// Note that you must have the required binaries setup under the bin directory to perform
// 		// the tests directly. When we run make test it will be setup and used automatically.
// 		BinaryAssetsDirectory: filepath.Join("bin", "k8s",
// 			fmt.Sprintf("1.28.3-%s-%s", runtime.GOOS, runtime.GOARCH)),
// 	}

// 	var err error
// 	// cfg is defined in this file globally.
// 	cfg, err = testEnv.Start()
// 	Expect(err).NotTo(HaveOccurred())
// 	Expect(cfg).NotTo(BeNil())

// 	err = nimbusv1.AddToScheme(scheme.Scheme)
// 	Expect(err).NotTo(HaveOccurred())

// 	//+kubebuilder:scaffold:scheme

// 	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
// 	Expect(err).NotTo(HaveOccurred())
// 	Expect(k8sClient).NotTo(BeNil())

// 	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
// 		Scheme: scheme.Scheme,
// 	})

// 	err = (&controllers.SecurityIntentReconciler{
// 		Client: k8sManager.GetClient(),
// 		Scheme: scheme.Scheme,
// 	}).SetupWithManager(k8sManager)
// 	Expect(err).ToNot(HaveOccurred())

// 	err = (&controllers.SecurityIntentBindingReconciler{
// 		Client: k8sManager.GetClient(),
// 		Scheme: scheme.Scheme,
// 	}).SetupWithManager(k8sManager)
// 	Expect(err).ToNot(HaveOccurred())

// 	go func() {
// 		defer GinkgoRecover()
// 		err = k8sManager.Start(ctrl.SetupSignalHandler())
// 		Expect(err).ToNot(HaveOccurred())
// 	}()

// })

// var _ = It("should create the NimbusPolicy automatically", func() {

// 	si := nimbusv1.SecurityIntent{
// 		TypeMeta: v1.TypeMeta{
// 			Kind:       "SecurityIntent",
// 			APIVersion: "intent.security.nimbus.com/v1",
// 		},

// 		ObjectMeta: v1.ObjectMeta{
// 			Name:      IntentName,
// 			Namespace: "default",
// 		},

// 		Spec: nimbusv1.SecurityIntentSpec{
// 			Intent: nimbusv1.Intent{
// 				ID:          "dnsManipulation",
// 				Description: "An adversary can manipulate DNS requests to redirect network traffic and potentially reveal end user activity.",
// 				Action:      "Block",
// 			},
// 		},
// 	}

// 	sib := nimbusv1.SecurityIntentBinding{

// 		TypeMeta: v1.TypeMeta{
// 			Kind:       "SecurityIntentBinding",
// 			APIVersion: "intent.security.nimbus.com/v1",
// 		},

// 		ObjectMeta: v1.ObjectMeta{
// 			Name:      IntentBindingName,
// 			Namespace: "default",
// 		},

// 		Spec: nimbusv1.SecurityIntentBindingSpec{
// 			Intents: []nimbusv1.MatchIntent{
// 				{
// 					Name: IntentName,
// 				},
// 			},

// 			Selector: nimbusv1.Selector{
// 				Any: []nimbusv1.ResourceFilter{
// 					{
// 						Resources: nimbusv1.Resources{
// 							Kind:      "Pod",
// 							Namespace: "default",
// 							MatchLabels: map[string]string{
// 								"app": "nginx",
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	err := k8sClient.Create(context.TODO(), &si, &client.CreateOptions{})

// 	Expect(err).NotTo(HaveOccurred())

// 	err = k8sClient.Create(context.TODO(), &sib, &client.CreateOptions{})

// 	Expect(err).NotTo(HaveOccurred())

// 	// Wait for the NimbusPolicy to be created
// 	waitForIntentBinding(k8sClient, "default", IntentBindingName)
// 	waitForNimbusPolicy(k8sClient, "default", IntentBindingName)

// 	matchLabels := map[string]string{
// 		"app": "nginx",
// 	}

// 	nimbusRulesList := []nimbusv1.NimbusRules{
// 		{
// 			ID:          "dnsManipulation",
// 			Description: "An adversary can manipulate DNS requests to redirect network traffic and potentially reveal end user activity.",
// 			Rule: nimbusv1.Rule{
// 				RuleAction: "Block",
// 				Mode:       "best-effort",
// 			},
// 		},
// 	}

// 	policy := &nimbusv1.NimbusPolicy{
// 		TypeMeta: v1.TypeMeta{
// 			Kind:       "NimbusPolicy",
// 			APIVersion: "intent.security.nimbus.com/v1",
// 		},

// 		ObjectMeta: v1.ObjectMeta{
// 			Name: IntentBindingName,
// 		},

// 		Spec: nimbusv1.NimbusPolicySpec{
// 			Selector: nimbusv1.NimbusSelector{
// 				MatchLabels: matchLabels,
// 			},
// 			NimbusRules: nimbusRulesList,
// 		},
// 	}

// 	nimbusPolicy := nimbusv1.NimbusPolicy{}

// 	// Retrieve the NimbusPolicy
// 	err = k8sClient.Get(context.TODO(), types.NamespacedName{Name: "dns-manipulation-binding", Namespace: "default"}, &nimbusPolicy, &client.GetOptions{})
// 	Expect(err).NotTo(HaveOccurred())

// 	// Assert the content of the NimbusPolicy
// 	Expect(nimbusPolicy.Spec).To(Equal(policy.Spec))

// })

// func waitForNimbusPolicy(clientset client.Client, namespace, policyName string) {
// 	nimbusPolicy := nimbusv1.NimbusPolicy{}
// 	Eventually(func() bool {
// 		err := clientset.Get(context.TODO(), types.NamespacedName{Name: IntentBindingName, Namespace: "default"}, &nimbusPolicy, &client.GetOptions{})

// 		return err == nil
// 	}, time.Second*10).Should(BeTrue(), "NimbusPolicy not created")

// 	// clientset.List(context.TODO(), &nimbusPolicy, &client.ListOptions{Namespace: "default"})

// 	// fmt.Println(nimbusPolicy.Items)
// }

// func waitForIntentBinding(clientset client.Client, namespace, policyName string) {
// 	sib := nimbusv1.SecurityIntentBinding{}
// 	Eventually(func() bool {
// 		err := clientset.Get(context.TODO(), types.NamespacedName{Name: IntentBindingName, Namespace: "default"}, &sib, &client.GetOptions{})
// 		return err == nil
// 	}, time.Second*5).Should(BeTrue(), "sib not created")
// }

// // var _ = AfterSuite(func() {
// // 	By("tearing down the test environment")
// // 	err := testEnv.Stop()
// // 	Expect(err).NotTo(HaveOccurred())
// // })
