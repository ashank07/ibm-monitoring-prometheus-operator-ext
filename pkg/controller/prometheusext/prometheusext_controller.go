//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package prometheusext

import (
	"context"
	"time"

	promev1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	exportersv1alpha1 "github.com/IBM/ibm-monitoring-exporters-operator/pkg/apis/monitoring/v1alpha1"
	monitoringv1alpha1 "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
	"github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/model"
	"github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/reconsiler"
)

var log = logf.Log.WithName("controller_prometheusext")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PrometheusExt Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePrometheusExt{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("prometheusext-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PrometheusExt
	err = c.Watch(&source.Kind{Type: &monitoringv1alpha1.PrometheusExt{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch deployment - for mcm monitor controller
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &monitoringv1alpha1.PrometheusExt{},
	})
	if err != nil {
		return err
	}

	// Watch Prometheus
	// There might be multilple Prometheus instances for MCM Hub.
	// We watch only the one created indirectly by PrometheusExt CR
	err = c.Watch(&source.Kind{Type: &promev1.Prometheus{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &monitoringv1alpha1.PrometheusExt{},
	})
	if err != nil {
		return err
	}
	// Watch Alertmanager
	err = c.Watch(&source.Kind{Type: &promev1.Alertmanager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &monitoringv1alpha1.PrometheusExt{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to Exporter object for scrape configuration no matter if it is created by us or not
	err = c.Watch(&source.Kind{Type: &exportersv1alpha1.Exporter{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePrometheusExt implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePrometheusExt{}

// ReconcilePrometheusExt reconciles a PrometheusExt object
type ReconcilePrometheusExt struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PrometheusExt object and makes changes based on the state read
// and what is in the PrometheusExt.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePrometheusExt) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PrometheusExt")

	// Fetch the PrometheusExt instance
	instance := &monitoringv1alpha1.PrometheusExt{}
	ctx := context.Background()
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	reconsiler := reconsiler.Reconsiler{
		Client:  r.client,
		CR:      instance.DeepCopy(),
		Schema:  r.scheme,
		Context: ctx,
	}
	if err := reconsiler.ReadClusterState(); err != nil {
		return reconcile.Result{}, err
	}
	if err := reconsiler.Sync(); err != nil {
		if !model.IsRequeueErr(err) {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second}, nil
	}

	return reconcile.Result{}, nil
}
