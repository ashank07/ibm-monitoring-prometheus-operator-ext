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
	"fmt"
	"strings"

	promev1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apisv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ev1beta1 "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	exportersv1alpha1 "github.com/IBM/ibm-monitoring-exporters-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/IBM/ibm-monitoring-exporters-operator/pkg/controller/exporter/model"
	monitoringv1alpha1 "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
	promodel "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/model"
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
	r.updateStatus()
	if err := r.syncStorageClass(); err != nil {
		return err
	}
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
func (r *Reconsiler) updateStatus() {
	if r.CurrentState.PrometheusOperatorDeployment != nil {
		r.CR.Status.PrometheusOperator = fmt.Sprintf("%d desired | %d updated | %d total | %d available | %d unavailable",
			r.CurrentState.PrometheusOperatorDeployment.Status.Replicas,
			r.CurrentState.PrometheusOperatorDeployment.Status.UpdatedReplicas,
			r.CurrentState.PrometheusOperatorDeployment.Status.ReadyReplicas,
			r.CurrentState.PrometheusOperatorDeployment.Status.AvailableReplicas,
			r.CurrentState.PrometheusOperatorDeployment.Status.UnavailableReplicas)
	}
	if r.CurrentState.ManagedPrometheus == nil {
		r.CR.Status.Prometheus = model.NotReady
	} else {
		r.CR.Status.Prometheus = r.CurrentState.ManagedPrometheus.ObjectMeta.Name
	}
	if r.CurrentState.ManagedAlertmanager == nil {
		r.CR.Status.Alertmanager = model.NotReady
	} else {
		r.CR.Status.Alertmanager = r.CurrentState.ManagedAlertmanager.ObjectMeta.Name
	}
	if r.CurrentState.Exporter == nil {
		r.CR.Status.Exporter = model.NotReady
	} else {
		r.CR.Status.Exporter = r.CurrentState.Exporter.ObjectMeta.Name
	}
	r.CR.Status.Configmaps = r.cmStatus()
	r.CR.Status.Secrets = r.secretStatus()
	if err := r.Client.Status().Update(r.Context, r.CR); err != nil {
		log.Error(err, "Failed to update status")
	}

}
func (r *Reconsiler) cmStatus() string {
	var ready []string
	var notReady []string
	if r.CurrentState.PromeNgCm != nil {
		ready = append(ready, promodel.ProRouterNgCmName(r.CR))
	} else {
		notReady = append(notReady, promodel.ProRouterNgCmName(r.CR))
	}
	if r.CurrentState.RouterEntryCm != nil {
		ready = append(ready, promodel.RouterEntryCmName(r.CR))
	} else {
		notReady = append(notReady, promodel.RouterEntryCmName(r.CR))
	}
	if r.CurrentState.ProLuaUtilsCm != nil {
		ready = append(ready, promodel.ProLuaUtilsCmName(r.CR))
	} else {
		notReady = append(notReady, promodel.ProLuaUtilsCmName(r.CR))
	}

	if r.CurrentState.ProLuaCm != nil {
		ready = append(ready, promodel.ProLuaCmName(r.CR))
	} else {
		notReady = append(notReady, promodel.ProLuaCmName(r.CR))
	}

	if r.CurrentState.AlertNgCm != nil {
		ready = append(ready, promodel.AlertRouterNgCmName(r.CR))
	} else {
		notReady = append(notReady, promodel.AlertRouterNgCmName(r.CR))
	}
	readyStr := strings.Join(ready, " ")
	if strings.TrimSpace(readyStr) == "" {
		readyStr = promodel.None

	}
	notReadyStr := strings.Join(notReady, " ")
	if strings.TrimSpace(notReadyStr) == "" {
		notReadyStr = promodel.None

	}

	return model.Ready + ": " + readyStr + ", " + model.NotReady + ": " + notReadyStr
}
func (r *Reconsiler) secretStatus() string {
	var ready []string
	var notReady []string
	if r.CurrentState.MonitoringSecret != nil {
		ready = append(ready, r.CR.Spec.Certs.MonitoringSecret)
	} else {
		notReady = append(notReady, r.CR.Spec.Certs.MonitoringSecret)
	}
	if r.CurrentState.MonitoringClientSecret != nil {
		ready = append(ready, r.CR.Spec.Certs.MonitoringClientSecret)
	} else {
		notReady = append(notReady, r.CR.Spec.Certs.MonitoringClientSecret)
	}
	if r.CurrentState.PrometheusScrapeTargetsSecret != nil {
		ready = append(ready, promodel.ScrapeTargetsSecretName(r.CR))
	} else {
		notReady = append(notReady, promodel.ScrapeTargetsSecretName(r.CR))
	}
	readyStr := strings.Join(ready, " ")
	if strings.TrimSpace(readyStr) == "" {
		readyStr = promodel.None

	}
	notReadyStr := strings.Join(notReady, " ")
	if strings.TrimSpace(notReadyStr) == "" {
		notReadyStr = promodel.None

	}

	return model.Ready + ": " + readyStr + ", " + model.NotReady + ": " + notReadyStr
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
