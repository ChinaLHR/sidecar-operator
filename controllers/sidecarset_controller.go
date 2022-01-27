/*
Copyright 2022.

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
	"encoding/json"
	appsv1alpha1 "github.com/ChinaLHR/sidecar-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// SidecarSetReconciler reconciles a SidecarSet object
type SidecarSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.chinalhr.github.io,resources=sidecarsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.chinalhr.github.io,resources=sidecarsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.chinalhr.github.io,resources=sidecarsets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get

func (r *SidecarSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	//next reconcile
	reconcileResult := ctrl.Result{RequeueAfter: time.Second * 5, Requeue: true}

	//fetch the SidecarSet instance
	sidecarSet := &appsv1alpha1.SidecarSet{}
	err := r.Get(context.TODO(), req.NamespacedName, sidecarSet)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "process sidecarSet error : object not found")
			return reconcileResult, nil
		}
		log.Error(err, "process sidecarSet error : reading object error")
		return reconcile.Result{}, err
	}

	log.V(0).Info("begin to process sidecarSet", "name", sidecarSet.Name)
	//list matched pod
	selector, err := metav1.LabelSelectorAsSelector(sidecarSet.Spec.Selector)
	if err != nil {
		log.Error(err, "process sidecarSet error")
		return reconcile.Result{}, err
	}

	listOpts := &client.ListOptions{LabelSelector: selector}

	matchedPods := &v1.PodList{}
	if err := r.Client.List(context.TODO(), matchedPods, listOpts); err != nil {
		log.Error(err, "list matched pods error")
		return reconcile.Result{}, err
	}
	log.V(0).Info("process sidecarSet matchedPods", "pods", matchedPods)

	// ignore inactive pods
	var filteredPods []*v1.Pod
	for i := range matchedPods.Items {
		pod := &matchedPods.Items[i]
		if IsPodActive(pod) {
			filteredPods = append(filteredPods, pod)
		}
	}

	//calculateSidecarSetStatus
	status, err := calculateSidecarSetStatus(sidecarSet, filteredPods)
	if err != nil {
		log.Error(err, "calculateSidecarSetStatus error")
		return reconcile.Result{}, err
	}

	//updateSidecarSetStatus
	err = r.updateSidecarSetStatus(log, sidecarSet, status)
	if err != nil {
		log.Error(err, "updateSidecarSetStatus error")
		return reconcile.Result{}, err
	}
	return reconcileResult, nil
}

func (r *SidecarSetReconciler) updateSidecarSetStatus(log logr.Logger, sidecarSet *appsv1alpha1.SidecarSet, status *appsv1alpha1.SidecarSetStatus) error {
	sidecarSetClone := sidecarSet.DeepCopy()
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		sidecarSetClone.Status = *status

		updateErr := r.Status().Update(context.TODO(), sidecarSetClone)
		if updateErr == nil {
			return nil
		}

		key := types.NamespacedName{
			Name: sidecarSetClone.Name,
		}

		if err := r.Get(context.TODO(), key, sidecarSetClone); err != nil {
			log.V(0).Info("error getting updated sidecarSet from client", "sidecarSet", sidecarSetClone.Name)
		}

		return updateErr
	})

	return err
}

func calculateSidecarSetStatus(sidecarSet *appsv1alpha1.SidecarSet, pods []*v1.Pod) (*appsv1alpha1.SidecarSetStatus, error) {
	var matchedPods, updatedPods, readyPods int32
	matchedPods = int32(len(pods))
	for _, pod := range pods {
		updated, err := isPodSidecarUpdated(sidecarSet, pod)
		if err != nil {
			return nil, err
		}
		if updated {
			updatedPods++
		}

		if isRunningAndReady(pod) {
			readyPods++
		}
	}

	return &appsv1alpha1.SidecarSetStatus{
		MatchedPods: matchedPods,
		UpdatedPods: updatedPods,
		ReadyPods:   readyPods,
	}, nil
}

func isPodSidecarUpdated(sidecarSet *appsv1alpha1.SidecarSet, pod *v1.Pod) (bool, error) {
	hashKey := appsv1alpha1.SidecarAnnotationHashKey
	if pod.Annotations[hashKey] == "" {
		return false, nil
	}

	sidecarSetHash := make(map[string]string)
	if err := json.Unmarshal([]byte(pod.Annotations[hashKey]), &sidecarSetHash); err != nil {
		return false, err
	}

	return sidecarSetHash[sidecarSet.Name] == sidecarSet.Annotations[hashKey], nil
}

func isRunningAndReady(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodRunning
}

func IsPodActive(p *v1.Pod) bool {
	return v1.PodSucceeded != p.Status.Phase &&
		v1.PodFailed != p.Status.Phase &&
		p.DeletionTimestamp == nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SidecarSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SidecarSet{}).
		Complete(r)
}
