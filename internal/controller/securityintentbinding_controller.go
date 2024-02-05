// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
	"github.com/5GSEC/nimbus/pkg/processor/policybuilder"
)

// SecurityIntentBindingReconciler reconciles a SecurityIntentBinding object
type SecurityIntentBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	sib := &v1.SecurityIntentBinding{}
	err := r.Get(ctx, req.NamespacedName, sib)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("SecurityIntentBinding not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return ctrl.Result{}, err
	}

	if sib.Status.Status == "" {
		sib.Status.Status = StatusCreated
		sib.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, sib); err != nil {
			logger.Error(err, "failed to update SecurityIntentBinding status", "SecurityIntentBinding.Name", sib.Name, "SecurityIntentBinding.Namespace", sib.Namespace)
			return ctrl.Result{}, err
		}
		// Let's re-fetch the SecurityIntentBinding CR after updating the status so that
		// we have the latest state of the resource on the cluster.
		if err := r.Get(ctx, req.NamespacedName, sib); err != nil {
			logger.Error(err, "failed to re-fetch SecurityIntentBinding", "SecurityIntentBinding.Name", sib.Name, "SecurityIntentBinding.Namespace", sib.Namespace)
			return ctrl.Result{}, err
		}
		logger.Info("SecurityIntentBinding found", "SecurityIntentBinding.Name", sib.Name, "SecurityIntentBinding.Namespace", sib.Namespace)
		return ctrl.Result{}, nil
	}

	bindingInfo := intentbinder.MatchAndBindIntents(ctx, r.Client, sib)
	nimbusPolicy, err := policybuilder.BuildNimbusPolicy(ctx, r.Client, r.Scheme, bindingInfo)
	if err != nil {
		logger.Error(err, "failed to build NimbusPolicy")
		return ctrl.Result{}, err
	}

	existingNp := &v1.NimbusPolicy{}
	err = r.Get(ctx, req.NamespacedName, existingNp)

	// NimbusPolicy가 존재하지 않을 때, 즉 새로 NimbusPolicy를 생성해야 할 때
	if err != nil && errors.IsNotFound(err) {
		// NimbusPolicy 생성 로직
		nimbusPolicy.Status.LastUpdated = metav1.Now()
		if err := r.Create(ctx, nimbusPolicy); err != nil {
			logger.Error(err, "failed to create NimbusPolicy", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
			return ctrl.Result{}, err
		}
		// 잠시 대기하여 Kubernetes API 서버가 NimbusPolicy 정보를 업데이트할 시간을 준다.
		time.Sleep(time.Second * 10)
		return ctrl.Result{}, nil
	}

	// NimbusPolicy가 이미 존재할 때, 즉 NimbusPolicy를 업데이트해야 할 때
	if err == nil {
		// 생성된 NimbusPolicy와 기존 NimbusPolicy를 비교하여 업데이트가 필요한지 확인
		if !reflect.DeepEqual(existingNp.Spec, nimbusPolicy.Spec) || !reflect.DeepEqual(existingNp.Status, nimbusPolicy.Status) {
			nimbusPolicy.ObjectMeta.ResourceVersion = existingNp.ObjectMeta.ResourceVersion
			logger.Info("Updating NimbusPolicy due to changes", "NimbusPolicy.Name", existingNp.Name)
			if err := r.Update(ctx, nimbusPolicy); err != nil {
				logger.Error(err, "Failed to update NimbusPolicy", "NimbusPolicy.Name", nimbusPolicy.Name)
				return ctrl.Result{}, err
			}
		}
	}

	//nimbusPolicy.ObjectMeta = existingNp.ObjectMeta
	//nimbusPolicy.Status.LastUpdated = metav1.Now()
	//if err := r.Update(ctx, nimbusPolicy); err != nil {
	//	logger.Error(err, "failed to update NimbusPolicy")
	//	return ctrl.Result{}, err
	//}

	if nimbusPolicy.Status.Status == "" || nimbusPolicy.Status.Status == StatusPending {
		nimbusPolicy.Status.Status = StatusCreated
		nimbusPolicy.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, nimbusPolicy); err != nil {
			logger.Error(err, "failed to update NimbusPolicy status", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
			return ctrl.Result{}, err
		}
		logger.Info("NimbusPolicy created", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntentBinding{}).
		Owns(&v1.NimbusPolicy{}).
		Complete(r)
}
