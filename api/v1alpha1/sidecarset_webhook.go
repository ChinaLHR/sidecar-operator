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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var sidecarsetlog = logf.Log.WithName("sidecarset-resource")

func (r *SidecarSet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-apps-chinalhr-github-io-v1alpha1-sidecarset,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.chinalhr.github.io,resources=sidecarsets,verbs=create;update,versions=v1alpha1,name=vsidecarset.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &SidecarSet{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SidecarSet) ValidateCreate() error {
	sidecarsetlog.Info("validate create", "name", r.Name)
	allErrs := r.validateSpec()
	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "apps.chinalhr.github.io", Kind: "SidecarSet"},
			r.Name, allErrs)
	}
	return nil
}

func (r *SidecarSet) validateSpec() field.ErrorList {
	spec := r.Spec
	var allErrs field.ErrorList

	if spec.Selector == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("selector"), "no selector defined for sidecarset"))
	} else {
		if len(spec.Selector.MatchLabels)+len(spec.Selector.MatchExpressions) == 0 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("selector"), spec.Selector, "empty selector is not valid for sidecarset."))
		}
	}
	return allErrs
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SidecarSet) ValidateUpdate(old runtime.Object) error {
	sidecarsetlog.Info("validate update", "name", r.Name)
	allErrs := r.validateSpec()
	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "apps.chinalhr.github.io", Kind: "SidecarSet"},
			r.Name, allErrs)
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SidecarSet) ValidateDelete() error {
	sidecarsetlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
