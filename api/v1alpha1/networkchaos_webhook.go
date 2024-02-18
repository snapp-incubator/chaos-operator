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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var networkchaoslog = logf.Log.WithName("networkchaos-resource")
var runtimeClient client.Client

func (r *NetworkChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	runtimeClient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-chaos-snappcloud-io-v1alpha1-networkchaos,mutating=true,failurePolicy=fail,sideEffects=None,groups=chaos.snappcloud.io,resources=networkchaos,verbs=create;update,versions=v1alpha1,name=mnetworkchaos.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &NetworkChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *NetworkChaos) Default() {
	networkchaoslog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-chaos-snappcloud-io-v1alpha1-networkchaos,mutating=false,failurePolicy=fail,sideEffects=None,groups=chaos.snappcloud.io,resources=networkchaos,verbs=create;update,versions=v1alpha1,name=vnetworkchaos.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NetworkChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NetworkChaos) ValidateCreate() (admission.Warnings, error) {
	ctx := context.Background()
	svcNamespacedName := types.NamespacedName{
		Name: r.Spec.Upstream.Name,
		// Assuming the service is in the same namespace as the NetworkChaos object
		Namespace: r.Namespace,
	}
	svc := &v1.Service{}
	if err := runtimeClient.Get(ctx, svcNamespacedName, svc); err != nil {
		return nil, fmt.Errorf("failed to find the specified upstream service (%s) in namespace (%s): %v", r.Spec.Upstream.Name, r.Namespace, err)
	}
	// Check if the specified port is present in the Service
	upstreamPort, _ := strconv.ParseInt(r.Spec.Upstream.Port, 10, 32)
	portFound := false
	for _, p := range svc.Spec.Ports {
		if int32(upstreamPort) == p.Port {
			portFound = true
			break
		}
	}
	if !portFound {
		return nil, fmt.Errorf("specified port %s not found in upstream service %s", r.Spec.Upstream.Port, r.Spec.Upstream.Name)
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NetworkChaos) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldNetworkChaos, _ := old.(*NetworkChaos)

	if r.Spec.Stream != oldNetworkChaos.Spec.Stream {
		return nil, errors.New("modification of Stream field is not allowed")
	}
	if r.Spec.Upstream != oldNetworkChaos.Spec.Upstream {
		return nil, errors.New("modification of Upstream field is not allowed")
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NetworkChaos) ValidateDelete() (admission.Warnings, error) {
	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
