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

package reconsiler

import (
	"context"

	promev1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apisv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	exportersv1alpha1 "github.com/IBM/ibm-monitoring-exporters-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/IBM/ibm-monitoring-exporters-operator/pkg/controller/exporter/model"
	monitoringv1alpha1 "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"

	ev1beta1 "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("prometheus-operator-ext.reconsiler")

// Reconsiler sync status of objects
type Reconsiler struct {
	CR           *monitoringv1alpha1.PrometheusExt
	CurrentState *ClusterState
	Schema       *runtime.Scheme
	Context      context.Context
	Client       client.Client
}

// ClusterState store current state of observed objects in the cluster
type ClusterState struct {
	MCMCtrlDeployment             *appsv1.Deployment
	ManagedPrometheus             *promev1.Prometheus   //create by this CR
	ManagedAlertmanager           *promev1.Alertmanager //created by this CR
	PromeSvc                      *v1.Service
	PromeIngress                  *ev1beta1.Ingress
	AlertmanagerSvc               *v1.Service
	AlertManagerIngress           *ev1beta1.Ingress
	Exporter                      *exportersv1alpha1.Exporter
	MonitoringSecret              *v1.Secret
	MonitoringClientSecret        *v1.Secret
	PrometheusScrapeTargetsSecret *v1.Secret
	PromeNgCm                     *v1.ConfigMap
	RouterEntryCm                 *v1.ConfigMap
	ProLuaUtilsCm                 *v1.ConfigMap
	ProLuaCm                      *v1.ConfigMap
	AlertNgCm                     *v1.ConfigMap
	PrometheusOperatorDeployment  *appsv1.Deployment
}

// ReadClusterState Read objects managed by this CR from cluster
func (r *Reconsiler) ReadClusterState() error {
	r.CurrentState = &ClusterState{}
	if err := r.readSecrets(); err != nil {
		return err
	}
	if err := r.readRouterCms(); err != nil {
		return err
	}
	if err := r.readExporter(); err != nil {
		return err
	}
	if err := r.readPrometheus(); err != nil {
		return err
	}
	if err := r.readAlertmanager(); err != nil {
		return err
	}
	if err := r.readMCMCtlDeployment(); err != nil {
		return err
	}
	if err := r.readPrometheusOperatorDeployment(); err != nil {
		return err
	}
	return nil
}

// Sync makes cluster state as expected
func (r *Reconsiler) Sync() error {
	if err := r.syncProOperatorDeployment(); err != nil {
		return err
	}
	if err := r.syncSecrets(); err != nil {
		return err
	}
	if err := r.syncRouterCms(); err != nil {
		return err
	}
	if err := r.syncPrometheus(); err != nil {
		return err
	}
	if err := r.syncAlertmanager(); err != nil {
		return err
	}
	if err := r.syncMCMCtl(); err != nil {
		return err
	}
	return nil
}

func (r *Reconsiler) createObject(obj runtime.Object) error {
	if err := controllerutil.SetControllerReference(r.CR, obj.(apisv1.Object), r.Schema); err != nil {
		return err
	}
	return r.Client.Create(r.Context, obj)
}

func (r *Reconsiler) updateObject(obj runtime.Object) error {
	if err := controllerutil.SetControllerReference(r.CR, obj.(apisv1.Object), r.Schema); err != nil {
		return err
	}
	if err := r.Client.Update(r.Context, obj); err != nil {
		if kerrors.IsConflict(err) {
			return model.NewRequeueError("sync.UpdateObject", "Object version conflict when updating and requeue it")
		}
		return err

	}
	return nil

}
