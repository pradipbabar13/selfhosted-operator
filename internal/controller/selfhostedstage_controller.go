/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	selfhosteddevv1 "github.com/pradipbabar13/selfhosted-operator/api/v1"
)

// SelfhostedstageReconciler reconciles a Selfhostedstage object
type SelfhostedstageReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=selfhosted.dev.my.domain,resources=selfhostedstages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=selfhosted.dev.my.domain,resources=selfhostedstages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=selfhosted.dev.my.domain,resources=selfhostedstages/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Selfhostedstage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SelfhostedstageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("stage", req.NamespacedName)

	// Fetch the Stage instance
	stage := &selfhosteddevv1.Selfhostedstage{}
	err := r.Get(ctx, req.NamespacedName, stage)
	if err != nil {
		if errors.IsNotFound(err) {
			// The Selfhostedstage resource may have been deleted after the reconcile request.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		log.Error(err, "Failed to get Stage")
		return reconcile.Result{}, err
	}
	// Check if namespaceName is specified
	if stage.Spec.Environment == "" {
		log.Info("NamespaceName not specified in Stage spec")
		return reconcile.Result{}, nil
	}
	// Create Namespace object
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: stage.Spec.Environment,
		},
	}

	// Check if the Namespace already exists
	found := &corev1.Namespace{}
	err = r.Get(ctx, types.NamespacedName{Name: namespace.Name}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the Namespace
		log.Info("Creating a new Namespace", "Namespace.Name", namespace.Name)
		err = r.Create(ctx, namespace)
		if err != nil {
			log.Error(err, "Failed to create new Namespace", "Namespace.Name", namespace.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Namespace")
		return reconcile.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SelfhostedstageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&selfhosteddevv1.Selfhostedstage{}).
		Complete(r)
}
