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
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"
	"strconv"

	mycorev1 "ds.korea.ac.kr/dnclabreplicaset/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
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

func (r *DnclabReplicaSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	_ = context.Background()
	drsLogger := r.Log.WithValues("dnclabreplicaset", req.NamespacedName)

	// your logic here
	// Get DnclabReplicaSet
	drs := &mycorev1.DnclabReplicaSet{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, drs)
	if err != nil {
		if errors.IsNotFound(err) {
			drsLogger.Info("dnclabreplicaset resource not found")
			return ctrl.Result{}, nil
		}

		drsLogger.Error(err, "failed to get drs")
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
		drsLogger.Error(err, "failed to get pod list", "DsReplicaset Namespace", drs.Namespace, "name", drs.Name)
		return ctrl.Result{}, err
	}

	// Create or Delete Pod
	replicaDiff := getNumOfPods(drsLogger) - len(podList.Items)
	replicaDiffStrMsg := "replicaDiff : " + strconv.Itoa(replicaDiff)
	drsLogger.Info(replicaDiffStrMsg)

	if replicaDiff > 0 {
		for i := 0; i < replicaDiff; i++ {
			// set pod
			pod := &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: drs.Namespace,
					Name:      getRandomPodName(drs.Name),
					Labels:    labelsForDsReplicaSet(drs.Name),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  drs.Spec.Name,
						Image: drs.Spec.Image,
					}},
				},
			}
			if err := ctrl.SetControllerReference(drs, pod, r.Scheme); err != nil {
				drsLogger.Error(err, "failed to set controller reference for the pod.",
					"DnclabReplicaSet Namespace", drs.Namespace, "Name", drs.Name)
				return ctrl.Result{}, err
			}

			// Create Pod
			if err = r.Client.Create(context.TODO(), pod); err != nil {
				drsLogger.Error(err, "failed to create the pod.",
					"DnclabReplicaSet Namespace", drs.Namespace, "Name", drs.Name)
				return ctrl.Result{}, err
			}
		}

		return reconcile.Result{Requeue: true}, nil

	} else if replicaDiff < 0 {
		podNames := getSortedPodNames(podList.Items)
		for i := 0; i < -replicaDiff; i++ {
			pod := &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: drs.Namespace,
					Name:      podNames[i],
				},
			}

			// delete pod
			err = r.Client.Delete(context.TODO(), pod)
			if err != nil {
				if !errors.IsNotFound(err) {
					drsLogger.Error(err, "failed to delete the pod.",
						"DnclabReplicaSet Namespace", drs.Namespace, "Name", drs.Name)
					return ctrl.Result{}, err
				}
			}
		}

		return reconcile.Result{Requeue: true}, nil
	}

	// update DnclabReplicaSet status
	podNames := getSortedPodNames(podList.Items)

	if !reflect.DeepEqual(podNames, drs.Status.PodNames) {
		drs.Status.PodNames = podNames
		err := r.Client.Status().Update(context.TODO(), drs)
		if err != nil {
			drsLogger.Error(err, "failed to update DnclabReplicaset status")
			return ctrl.Result{}, err
		}
	}

	drsLogger.Info("DnclabReplicSet controller is successfully reconciled!")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DnclabReplicaSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mycorev1.DnclabReplicaSet{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

// labelsForDsReplicaSet returns the labels for selecting the resources
// belonging to the given dsreplicaset CR name.
func labelsForDsReplicaSet(name string) map[string]string {
	return map[string]string{"dnclabreplicaset": name}
}

func getRandomPodName(name string) string {
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return "dnclabreplicaset-" + name + "-" + string(b)
}

func getSortedPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	sort.Strings(podNames)
	return podNames
}

// pod ?????? ????????? ?????? ????????? ????????? ??????
func getNumOfPods(drsLogger logr.Logger) int {
	numOfPods := rand.Intn(5) + 1
	//numOfPods := 5
	numOfPodsStr := "?????? Pod ?????? : " + strconv.Itoa(numOfPods)
	drsLogger.Info(numOfPodsStr)
	return numOfPods
}
