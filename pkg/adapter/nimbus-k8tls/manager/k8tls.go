// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/api/v1alpha1"
)

//+kubebuilder:rbac:groups="",resources=namespaces;serviceaccounts;configmaps,verbs=get;create;delete;update
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings,verbs=get;create;delete;update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list

func setupK8tlsEnv(ctx context.Context, cwnp v1alpha1.ClusterNimbusPolicy, scheme *runtime.Scheme, k8sClient client.Client) error {
	logger := log.FromContext(ctx)

	// Retrieve the namespace
	ns := &corev1.Namespace{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: NamespaceName}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "failed to fetch Namespace", "Namespace.Name", NamespaceName)
		}
		return err
	}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "fips-config",
			Namespace:   NamespaceName,
			Labels:      ns.Labels,
			Annotations: ns.Annotations,
		},
		Data: map[string]string{
			"fips-140-3.json": `
{
    "TLS_versions": [
        {
            "TLS_version": "TLSv1.0_1.1",
            "cipher_suites": [
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
                },
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
                }
            ]
        },
        {
            "TLS_version": "TLSv1.2",
            "cipher_suites": [
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
                },
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
                },
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
                },
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_256_CCM"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_128_CCM"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_256_CCM_8"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8"
                },
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384"
                },
                {
                    "cipher_suite": "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384"
                },
                {
                    "cipher_suite": "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256"
                }
            ]
        },
        {
            "TLS_version": "TLSv1.3",
            "cipher_suites": [
                {
                    "cipher_suite": "TLS_AES_256_GCM_SHA384"
                },
                {
                    "cipher_suite": "TLS_AES_128_GCM_SHA256"
                },
                {
                    "cipher_suite": "TLS_AES_128_CCM_SHA256"
                },
                {
                    "cipher_suite": "TLS_AES_128_CCM_8_SHA256"
                }
            ]
        }
    ]
}`,
		},
	}

	objectMeta := metav1.ObjectMeta{
		Name:        ns.Name,
		Namespace:   ns.Name,
		Labels:      ns.Labels,
		Annotations: ns.Annotations,
	}

	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: objectMeta,
	}

	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRole",
		},
		ObjectMeta: objectMeta,
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list"},
				APIGroups: []string{""},
				Resources: []string{"services"},
			},
		},
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: objectMeta,
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				APIGroup:  "",
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
	}

	objs := []client.Object{ns, cm, sa, clusterRole, clusterRoleBinding}
	for idx := range objs {
		objToCreate := objs[idx]

		// Don't set owner ref on namespace. In environments with configured Pod Security
		// Standards labelling namespaces becomes a requirement. However, on deletion of
		// CWNP a namespace with ownerReferences set also gets deleted. Since we need to
		// keep the nimbus-k8tls-env namespace labeled, removing the ownerReferences
		// prevents this deletion.
		if idx != 0 {
			if err := ctrl.SetControllerReference(&cwnp, objToCreate, scheme); err != nil {
				return err
			}
		}

		var existingObj client.Object

		// Set the type of object, otherwise existingObj will always remain nil.
		switch objToCreate.(type) {
		case *corev1.Namespace:
			existingObj = &corev1.Namespace{}
		case *corev1.ConfigMap:
			existingObj = &corev1.ConfigMap{}
		case *corev1.ServiceAccount:
			existingObj = &corev1.ServiceAccount{}
		case *rbacv1.ClusterRole:
			existingObj = &rbacv1.ClusterRole{}
		case *rbacv1.ClusterRoleBinding:
			existingObj = &rbacv1.ClusterRoleBinding{}
		}

		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(objToCreate), existingObj)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		objKind := strings.ToLower(objToCreate.GetObjectKind().GroupVersionKind().Kind)
		if err != nil {
			if errors.IsNotFound(err) {
				if err := k8sClient.Create(ctx, objToCreate); err != nil {
					return err
				}
				logger.Info(fmt.Sprintf("created %s/%s", objKind, objToCreate.GetName()))
			}
		} else {
			objToCreate.SetResourceVersion(existingObj.GetResourceVersion())
			if err := k8sClient.Update(ctx, objToCreate); err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("configured %s/%s", objKind, objToCreate.GetName()))
		}
	}

	return nil
}
