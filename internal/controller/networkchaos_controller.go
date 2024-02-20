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
	"strconv"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/pingcap/errors"
	chaosv1alpha1 "github.com/snapp-incubator/toxiproxy-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NetworkChaosReconciler reconciles a NetworkChaos object
type NetworkChaosReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Constants for Toxiproxy configurations
const (
	toxiproxyImage  = "ghcr.io/shopify/toxiproxy" // The Docker image for Toxiproxy
	toxiproxyPort   = 8474                        // Default port for Toxiproxy
	portFormatIndex = 5
	chaosFinalizer  = "chaos.snappcloud.io/cleanup-chaos"
)

//+kubebuilder:rbac:groups=chaos.snappcloud.io,resources=networkchaos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=chaos.snappcloud.io,resources=networkchaos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=chaos.snappcloud.io,resources=networkchaos/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NetworkChaos object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *NetworkChaosReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch NetworkChaos object
	networkChaos := &chaosv1alpha1.NetworkChaos{}
	if err := r.Client.Get(ctx, req.NamespacedName, networkChaos); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("NetworkChaos resource not found. Ignoring since the object must be deleted", "NetworkChaos", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get NetworkChaos")
		return ctrl.Result{}, err
	}

	if networkChaos.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(networkChaos, chaosFinalizer) {
			controllerutil.AddFinalizer(networkChaos, chaosFinalizer)
			if err := r.Update(ctx, networkChaos); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(networkChaos, chaosFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.finalizeNetworkChaos(ctx, req, networkChaos); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried.
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(networkChaos, chaosFinalizer)
			if err := r.Update(ctx, networkChaos); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Ensure toxiproxy Deployment is created
	if err := r.ensureToxiproxyDeployment(ctx, req, networkChaos); err != nil {
		return ctrl.Result{}, err
	}
	// Ensure toxiproxy Service is created
	if err := r.ensureToxiproxyService(ctx, req, networkChaos); err != nil {
		return ctrl.Result{}, err
	}
	// Manage Toxiproxy Proxies and Toxics
	if err := r.manageToxiproxyProxies(ctx, req, networkChaos); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NetworkChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Predicate to filter Pods with the label app=toxiproxy
	labelPredicate := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetLabels()["app"] == "toxiproxy"
	})

	// Map a Pod deletion event to reconcile requests for all NetworkChaos resources
	mapFunc := func(c context.Context, a client.Object) []reconcile.Request {
		ctx := context.Background()
		log := log.FromContext(ctx)

		var requests []reconcile.Request

		var networkChaosList chaosv1alpha1.NetworkChaosList
		if err := mgr.GetClient().List(ctx, &networkChaosList, &client.ListOptions{}); err != nil {
			// Handle error
			log.Error(err, "Unable to list NetworkChaos resources")
			return nil
		}

		for _, networkChaos := range networkChaosList.Items {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      networkChaos.Name,
					Namespace: networkChaos.Namespace,
				},
			})
		}

		return requests
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&chaosv1alpha1.NetworkChaos{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(mapFunc),
			builder.WithPredicates(labelPredicate),
		).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *NetworkChaosReconciler) ensureToxiproxyDeployment(ctx context.Context, req ctrl.Request, networkChaos *chaosv1alpha1.NetworkChaos) error {
	log := log.FromContext(ctx)

	deployment := &appsv1.Deployment{}

	deploymentName := "toxiproxy-" + req.Namespace

	// Try to get the Deployment if it exists
	if err := r.Client.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: req.Namespace}, deployment); err != nil {
		if errors.IsNotFound(err) {
			dep := r.createToxiproxyDeployment(req.Namespace)
			if err = r.Client.Create(ctx, dep); err != nil {
				log.Error(err, "Failed to create toxiproxy Deployment")
				return err
			}
			log.Info("Toxiproxy Deployment created successfully")
		} else {
			log.Error(err, "Failed to get toxiproxy Deployment")
			return err
		}
	}
	return nil
}
func (r *NetworkChaosReconciler) ensureToxiproxyService(ctx context.Context, req ctrl.Request, networkChaos *chaosv1alpha1.NetworkChaos) error {

	log := log.FromContext(ctx)
	svc := &corev1.Service{}

	svcName := "toxiproxy-" + req.Namespace
	selector := "toxiproxy"

	// Try to get the Service if it exists
	if err := r.Client.Get(ctx, types.NamespacedName{Name: svcName, Namespace: req.Namespace}, svc); err != nil {
		if errors.IsNotFound(err) {
			ser := r.createToxiproxyService(req.Namespace, svcName, selector, toxiproxyPort, toxiproxyPort)
			if err = r.Client.Create(ctx, ser); err != nil {
				log.Error(err, "Failed to create toxiproxy Service for TOXIPROXY")
				return err
			}
			log.Info("toxiproxy Service created successfully")
		} else {
			log.Error(err, "Failed to get toxiproxy Service")
			return err
		}
	}
	return nil
}
func (r *NetworkChaosReconciler) createToxiproxyDeployment(ns string) *appsv1.Deployment {
	// Define labels
	labels := map[string]string{"app": "toxiproxy"}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "toxiproxy-" + ns,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "proxy",
						Image: toxiproxyImage,
						// Define other container attributes (ports, env vars, etc.)
					}},
				},
			},
		},
	}

	return dep
}
func (r *NetworkChaosReconciler) createToxiproxyService(ns string, name string, selector string, port int, targetPort int) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": selector, // This should match labels of the Pods in the Deployment
			},
			Ports: []corev1.ServicePort{{
				Port:       int32(port),                // The port that the service should serve on
				TargetPort: intstr.FromInt(targetPort), // The target port on the pod to forward to
			}},
		},
	}

	return svc
}
func (r *NetworkChaosReconciler) manageToxiproxyProxies(ctx context.Context, req ctrl.Request, networkChaos *chaosv1alpha1.NetworkChaos) error {

	// Create a new Toxiproxy client
	toxiproxyClient := toxiproxy.NewClient("toxiproxy-" + req.Namespace + "." + req.Namespace + ".svc.cluster.local:8474")

	// Attempt to retrieve an existing proxy
	proxy, err := r.getOrCreateProxy(ctx, req, toxiproxyClient, networkChaos)
	if err != nil {
		return err
	}
	if err := r.ensureToxiproxyServiceForProxy(ctx, req, proxy, networkChaos); err != nil {
		return err
	}

	if err := r.manageToxics(ctx, req, proxy, networkChaos); err != nil {
		return err
	}

	return nil
}

func (r *NetworkChaosReconciler) getOrCreateProxy(ctx context.Context, req ctrl.Request, toxiproxyClient *toxiproxy.Client, networkChaos *chaosv1alpha1.NetworkChaos) (*toxiproxy.Proxy, error) {
	log := log.FromContext(ctx)

	proxy, err := toxiproxyClient.Proxy(networkChaos.GetName())
	listen := ""

	if err != nil {
		//before creating proxy check if svc exists, get the svc' port to create proxy with this port
		svc := &corev1.Service{}
		if err = r.Client.Get(ctx, types.NamespacedName{Name: "toxiproxy-" + networkChaos.GetName() + "-" + networkChaos.Spec.Upstream.Name, Namespace: req.Namespace}, svc); err == nil {
			if len(svc.Spec.Ports) > 0 {
				listen = "0.0.0.0:" + strconv.Itoa(int(svc.Spec.Ports[0].Port))
			} else {
				log.Error(err, "Service does not expose any ports")
			}
		}
		// if Proxy does not exist, create a new one
		proxy, err = toxiproxyClient.CreateProxy(networkChaos.GetName(), listen, networkChaos.Spec.Upstream.Name+":"+networkChaos.Spec.Upstream.Port)
		if err != nil {
			log.Error(err, "Failed to create proxy")
			return proxy, err
		}
		log.Info("proxy for service " + networkChaos.Spec.Upstream.Name + " created successfully in namespace " + req.Namespace)

	}
	return proxy, nil
}
func (r *NetworkChaosReconciler) ensureToxiproxyServiceForProxy(ctx context.Context, req ctrl.Request, proxy *toxiproxy.Proxy, networkChaos *chaosv1alpha1.NetworkChaos) error {
	log := log.FromContext(ctx)

	proxyPort := proxy.Listen[portFormatIndex:] // the format is " [::]:port"
	port, err := strconv.Atoi(proxyPort)
	if err != nil {
		log.Error(err, "its empty")
	}
	svc := &corev1.Service{}
	if err = r.Client.Get(ctx, types.NamespacedName{Name: "toxiproxy-" + networkChaos.GetName() + "-" + networkChaos.Spec.Upstream.Name, Namespace: req.Namespace}, svc); err != nil {
		if errors.IsNotFound(err) {
			svc = r.createToxiproxyService(req.Namespace, "toxiproxy-"+networkChaos.GetName()+"-"+networkChaos.Spec.Upstream.Name, "toxiproxy", port, port)
			if err = controllerutil.SetControllerReference(networkChaos, svc, r.Scheme); err != nil {
				log.Error(err, "Failed to add owner refrence")
				return err
			}
			if err = r.Client.Create(ctx, svc); err != nil {
				log.Error(err, "Failed to create Service for proxy")
				return err
			}
			log.Info("Service created successfully and Owner refrence added")
		} else {
			// Error other than NotFound
			log.Error(err, "Failed to get Service")
			return err
		}
	}
	return nil
}

func (r *NetworkChaosReconciler) manageToxics(ctx context.Context, req ctrl.Request, proxy *toxiproxy.Proxy, networkChaos *chaosv1alpha1.NetworkChaos) error {
	log := log.FromContext(ctx)

	// Check if the toxic already exists
	latencyExists := false
	timeoutExists := false

	toxics, err := proxy.Toxics()
	if err != nil {
		log.Error(err, "Failed to get toxics")
		return err
	}

	for _, toxic := range toxics {
		if toxic.Name == networkChaos.GetName()+"-latency" {
			latencyExists = true
			break
		}
	}

	// Update the toxic if it exists
	if latencyExists {
		_, err = proxy.UpdateToxic(networkChaos.GetName()+"-latency", networkChaos.Spec.LatencyToxic.Probability, toxiproxy.Attributes{
			"latency": networkChaos.Spec.LatencyToxic.Latency,
			"jitter":  networkChaos.Spec.LatencyToxic.Jitter,
		})
		if err != nil {
			log.Error(err, "Failed to update toxic")
			return err
		}
		log.Info("Toxic " + networkChaos.GetName() + " updated on " + proxy.Name + " proxy ")
	} else {

		// Add the toxic if it doesn't exist
		if networkChaos.Spec.LatencyToxic.Latency != 0 {

			_, err = proxy.AddToxic(networkChaos.GetName()+"-latency", "latency", networkChaos.Spec.Stream, networkChaos.Spec.LatencyToxic.Probability, toxiproxy.Attributes{
				"latency": networkChaos.Spec.LatencyToxic.Latency,
				"jitter":  networkChaos.Spec.LatencyToxic.Jitter,
			})
			if err != nil {
				log.Error(err, "Failed to create latency toxic")
				return err
			}
			log.Info("Latency toxic " + networkChaos.GetName() + " added on " + proxy.Name + " proxy ")
		}
	}
	for _, toxic := range toxics {
		if toxic.Name == networkChaos.GetName()+"-timeout" {
			timeoutExists = true
			break
		}
	}
	// Update the toxic if it exists
	if timeoutExists {
		_, err = proxy.UpdateToxic(networkChaos.GetName()+"-timeout", networkChaos.Spec.LatencyToxic.Probability, toxiproxy.Attributes{
			"timeout": networkChaos.Spec.TimeoutToxic.Timeout,
		})
		if err != nil {
			log.Error(err, "Failed to update toxic")
			return err
		}
		log.Info("Timeout toxic " + networkChaos.GetName() + " updated on " + proxy.Name + " proxy ")
	} else {
		if networkChaos.Spec.TimeoutToxic.Timeout != 0 {

			_, err = proxy.AddToxic(networkChaos.GetName()+"-timeout", "timeout", networkChaos.Spec.Stream, networkChaos.Spec.TimeoutToxic.Probability, toxiproxy.Attributes{
				"timeout": networkChaos.Spec.TimeoutToxic.Timeout,
			})
			if err != nil {
				log.Error(err, "Failed to create timeout toxic")
				return err
			}
			log.Info("Timeout toxic " + networkChaos.GetName() + " added on " + proxy.Name + " proxy ")
		}
	}
	// Disable the proxy if NetworkChaosSpec.Enable is false
	if !networkChaos.Spec.Enabled {
		if err := proxy.Disable(); err != nil {
			log.Error(err, "Failed to disable the proxy")
			return err
		}
		log.Info("Proxy Updated and disabled")
		return nil
	}
	return nil
}

func (r *NetworkChaosReconciler) finalizeNetworkChaos(ctx context.Context, req ctrl.Request, networkChaos *chaosv1alpha1.NetworkChaos) error {
	log := log.FromContext(ctx)

	// Initialize Toxiproxy client
	toxiproxyClient := toxiproxy.NewClient("toxiproxy-" + req.Namespace + "." + req.Namespace + ".svc.cluster.local:8474")

	proxyName := networkChaos.GetName()

	// Delete the proxy
	proxy, err := toxiproxyClient.Proxy(proxyName)
	if err != nil {
		log.Error(err, "Failed to get proxy")
		return err

	}
	err = proxy.Delete()
	if err != nil {
		log.Error(err, "Failed to delete Toxiproxy proxy")
		return err
	}
	log.Info("Successfully finalized and deleted Toxiproxy proxy")
	return nil

}
