/*
Copyright 2025.

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

package v1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	elasticwebv1 "elasticweb/api/v1"

	corev1 "k8s.io/api/core/v1"
)

// nolint:unused
// log is for logging in this package.
var elasticweblog = logf.Log.WithName("elasticweb-resource")

// SetupElasticWebWebhookWithManager registers the webhook for ElasticWeb in the manager.
func SetupElasticWebWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&elasticwebv1.ElasticWeb{}).
		WithValidator(&ElasticWebCustomValidator{}).
		WithDefaulter(&ElasticWebCustomDefaulter{
			DefaultTotalQPS: 1200,
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-elasticweb-com-bolingcavalry-v1-elasticweb,mutating=true,failurePolicy=fail,sideEffects=None,groups=elasticweb.com.bolingcavalry,resources=elasticwebs,verbs=create;update,versions=v1,name=melasticweb-v1.kb.io,admissionReviewVersions=v1

// ElasticWebCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ElasticWeb when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ElasticWebCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
	DefaultTotalQPS   int32
	DefaultResoureces corev1.ResourceRequirements
}

var _ webhook.CustomDefaulter = &ElasticWebCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ElasticWeb.
func (d *ElasticWebCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	elasticweb, ok := obj.(*elasticwebv1.ElasticWeb)

	if !ok {
		return fmt.Errorf("expected an ElasticWeb object but got %T", obj)
	}
	elasticweblog.Info("Defaulting for ElasticWeb", "name", elasticweb.GetName())

	if elasticweb.Spec.TotalQPS == nil {
		elasticweb.Spec.TotalQPS = new(int32)
		*elasticweb.Spec.TotalQPS = d.DefaultTotalQPS
		elasticweblog.Info("a. TotalQPS is nil, set default value now", "TotalQPS", *elasticweb.Spec.TotalQPS)
	} else {
		elasticweblog.Info("b. TotalQPS exists", "TotalQPS", *elasticweb.Spec.TotalQPS)
	}

	// for i1, v1 := range elasticweb.Spec.Deploy {
	// 	if v1.Resources == nil {
	// 		// Resources: corev1.ResourceRequirements{
	// 		// 	Requests: corev1.ResourceList{
	// 		// 		"cpu":    resource.MustParse(CPU_REQUEST),
	// 		// 		"memory": resource.MustParse(MEM_REQUEST),
	// 		// 	},
	// 		// 	Limits: corev1.ResourceList{
	// 		// 		"cpu":    resource.MustParse(CPU_LIMIT),
	// 		// 		"memory": resource.MustParse(MEM_LIMIT),
	// 		// 	},
	// 		// },
	// 		*(elasticweb.Spec.Deploy[i1].Resources) = corev1.ResourceRequirements{
	// 			Requests: corev1.ResourceList{
	// 				"cpu":    resource.MustParse("1"),
	// 				"memory": resource.MustParse("2Gi"),
	// 			},
	// 			Limits: corev1.ResourceList{
	// 				"cpu":    resource.MustParse("1"),
	// 				"memory": resource.MustParse("2Gi"),
	// 			},
	// 		}
	// 	} else {
	// 		if len(v1.Resources.Requests) == 0 {
	// 			*(&elasticweb.Spec.Deploy[i1].Resources.Requests) = *(&elasticweb.Spec.Deploy[i1].Resources.Limits)
	// 		} else {
	// 			*(&elasticweb.Spec.Deploy[i1].Resources.Limits) = *(&elasticweb.Spec.Deploy[i1].Resources.Requests)
	// 		}
	// 	}
	// }

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-elasticweb-com-bolingcavalry-v1-elasticweb,mutating=false,failurePolicy=fail,sideEffects=None,groups=elasticweb.com.bolingcavalry,resources=elasticwebs,verbs=create;update,versions=v1,name=velasticweb-v1.kb.io,admissionReviewVersions=v1

// ElasticWebCustomValidator struct is responsible for validating the ElasticWeb resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ElasticWebCustomValidator struct {
	//TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &ElasticWebCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ElasticWeb.
func (v *ElasticWebCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	elasticweb, ok := obj.(*elasticwebv1.ElasticWeb)
	if !ok {
		return nil, fmt.Errorf("expected a ElasticWeb object but got %T", obj)
	}
	elasticweblog.Info("Validation for ElasticWeb upon creation", "name", elasticweb.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, validateElasticWeb(elasticweb)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ElasticWeb.
func (v *ElasticWebCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	elasticweb, ok := newObj.(*elasticwebv1.ElasticWeb)
	if !ok {
		return nil, fmt.Errorf("expected a ElasticWeb object for the newObj but got %T", newObj)
	}
	elasticweblog.Info("Validation for ElasticWeb upon update", "name", elasticweb.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, validateElasticWeb(elasticweb)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ElasticWeb.
func (v *ElasticWebCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	elasticweb, ok := obj.(*elasticwebv1.ElasticWeb)
	if !ok {
		return nil, fmt.Errorf("expected a ElasticWeb object but got %T", obj)
	}
	elasticweblog.Info("Validation for ElasticWeb upon deletion", "name", elasticweb.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

func validateElasticWeb(r *elasticwebv1.ElasticWeb) error {
	var allErrs field.ErrorList

	if *r.Spec.SinglePodQPS > 1000 {
		elasticweblog.Info("c. Invalid SinglePodQPS")

		err := field.Invalid(field.NewPath("spec").Child("singlePodQPS"),
			*r.Spec.SinglePodQPS,
			"d. must be less than 1000")

		allErrs = append(allErrs, err)

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elasticweb.com.bolingcavalry", Kind: "ElasticWeb"},
			r.Name,
			allErrs)
	} else {
		elasticweblog.Info("e. SinglePodQPS is valid")
		return nil
	}
}
