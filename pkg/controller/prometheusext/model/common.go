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

package model

import (
	promev1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	monitoringv1alpha1 "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

//ObjectType means if it is prometheus or alertmanager
type ObjectType string

//ManagedLabel return labels tell if it is managed by current CR
func ManagedLabel(cr *monitoringv1alpha1.PrometheusExt) map[string]string {
	return map[string]string{managedLabelKey(): managedLabelValue(cr)}

}

//ObjectName returns name related to current cr
//TODO: howto sync name of grafana?
func ObjectName(cr *monitoringv1alpha1.PrometheusExt, t ObjectType) string {
	return cr.Name + "-" + string(t)
}

//ManagedPrometheus tells if prome is managed by cr
func ManagedPrometheus(cr *monitoringv1alpha1.PrometheusExt, prome *promev1.Prometheus) bool {
	if v, ok := prome.Labels[managedLabelKey()]; ok && v == cr.Name {
		return true
	}
	return false
}

//ManagedAlertmanager tells if prome is managed by cr
func ManagedAlertmanager(cr *monitoringv1alpha1.PrometheusExt, am *promev1.Alertmanager) bool {
	if v, ok := am.Labels[managedLabelKey()]; ok && v == cr.Name {
		return true
	}
	return false
}
func managedLabelKey() string {
	return "managed"
}
func managedLabelValue(cr *monitoringv1alpha1.PrometheusExt) string {
	return cr.Name
}

//MCMCtlDeploymentName gets mcm monitoring controller deployment name
func MCMCtlDeploymentName(cr *monitoringv1alpha1.PrometheusExt) string {
	return cr.Name + "-" + "-mcm-ctl"
}

//IReqeueError defines interface for requeueError
type IReqeueError interface {
	Reason() string
}
type requeueError struct {
	component string
	reason    string
}

// NewRequeueError creats requeueError error
func NewRequeueError(component string, reason string) error {
	return &requeueError{component, reason}

}
func (r *requeueError) Error() string {
	return "Component " + r.component + "requires to be requeued: " + r.reason
}
func (r *requeueError) Reason() string {
	return r.reason
}

//IsRequeueErr tells if error type is requeueError
func IsRequeueErr(e error) bool {
	switch e.(type) {
	case IReqeueError:
		return true
	}
	return false
}
func ingressAnnotations(cr *monitoringv1alpha1.PrometheusExt) map[string]string {
	return map[string]string{
		"kubernetes.io/ingress.class":                    "ibm-icp-management",
		"icp.management.ibm.com/authz-type":              "rbac",
		"icp.management.ibm.com/secure-backends":         "true",
		"icp.management.ibm.com/secure-client-ca-secret": cr.Spec.Certs.MonitoringClientSecret,
		"icp.management.ibm.com/rewrite-target":          "/",
	}
}
