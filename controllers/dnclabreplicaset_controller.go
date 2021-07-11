/*
Copyright 2021.

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

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mycorev1 "ds.korea.ac.kr/dnclabreplicaset/api/v1"

	corev1 "k8s.io/api/core/v1"
)

// DnclabReplicaSetReconciler reconciles a DnclabReplicaSet object
type DnclabReplicaSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=mycore.ds.korea.ac.kr,resources=dnclabreplicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mycore.ds.korea.ac.kr,resources=dnclabreplicasets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mycore.ds.korea.ac.kr,resources=dnclabreplicasets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DnclabReplicaSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DnclabReplicaSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("DnclabReplicaSet", req.NamespacedName)

	// your logic here
	// Get DnclabReplicaSet
	drs := &mycorev1.DnclabReplicaSet{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, drs)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "failed to get drs")
		return ctrl.Result{}, err
	}

	// Get all pods infos of DnclabReplicaSet
	podList := &corev1.PodList{}
	listOps := []client.ListOption{
		client.InNamespace(req.NamespacedName.Namespace),
		client.MatchingLabels(labelsForDsReplicaSet(drs.Name)),
	}
	err = r.Client.List(context.TODO(), podList, listOps...)
	if err != nil {
		log.Error(err, "failed to get pod list", "DsReplicaset Namespace", drs.Namespace, "name", drs.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DnclabReplicaSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mycorev1.DnclabReplicaSet{}).
		Complete(r)
}

// labelsForDsReplicaSet returns the labels for selecting the resources
// belonging to the given dsreplicaset CR name.
func labelsForDsReplicaSet(name string) map[string]string {
	return map[string]string{"app": "dsreplicaset", "dsreplicaset_cr": name}
}
