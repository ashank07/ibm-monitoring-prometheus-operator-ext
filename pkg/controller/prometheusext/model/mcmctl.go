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
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	promext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

var creationTime *metav1.Time = nil

//NewMCMCtlDeployment create new deployment object for mcm controller
func NewMCMCtlDeployment(cr *promext.PrometheusExt) (*appsv1.Deployment, error) {
	creationTime = &metav1.Time{Time: time.Now()}
	spec, err := mcmDeploymentSpec(cr)
	if err != nil {
		return nil, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MCMCtlDeploymentName(cr),
			Namespace: cr.Namespace,
			Labels:    msmcCtrlLabels(cr),
		},
		Spec: *spec,
	}

	return deployment, nil

}

//UpdatedMCMCtlDeployment create updated deployment object for mcm controller
func UpdatedMCMCtlDeployment(cr *promext.PrometheusExt, curr *appsv1.Deployment) (*appsv1.Deployment, error) {
	creationTime = &curr.ObjectMeta.CreationTimestamp
	spec, err := mcmDeploymentSpec(cr)
	if err != nil {
		return nil, err
	}

	deployment := curr.DeepCopy()
	deployment.ObjectMeta.Labels = msmcCtrlLabels(cr)
	deployment.Spec.Selector = spec.Selector

	deployment.Spec.Template.ObjectMeta.Labels = spec.Template.ObjectMeta.Labels
	deployment.Spec.Template.ObjectMeta.Annotations = spec.Template.ObjectMeta.Annotations
	deployment.Spec.Template.Spec.Containers = spec.Template.Spec.Containers
	deployment.Spec.Template.Spec.Volumes = spec.Template.Spec.Volumes
	deployment.Spec.Template.Spec.ImagePullSecrets = spec.Template.Spec.ImagePullSecrets
	deployment.Spec.Template.Spec.ServiceAccountName = spec.Template.Spec.ServiceAccountName

	return deployment, nil

}
func mcmDeploymentSpec(cr *promext.PrometheusExt) (*appsv1.DeploymentSpec, error) {
	replicas := int32(1)
	spec := &appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: msmcCtrlLabels(cr),
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:        MCMCtlDeploymentName(cr),
				Labels:      msmcCtrlLabels(cr),
				Annotations: commonPodAnnotations(),
			},
			Spec: v1.PodSpec{
				HostPID:     false,
				HostIPC:     false,
				HostNetwork: false,
				Volumes: []v1.Volume{
					{
						Name: "monitoring-ca-certs",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: cr.Spec.Certs.MonitoringSecret,
							},
						},
					},
					{
						Name: "monitoring-client-certs",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: cr.Spec.Certs.MonitoringClientSecret,
							},
						},
					},
				},
			},
		},
	}

	if cr.Spec.ImagePullSecrets != nil && len(cr.Spec.ImagePullSecrets) != 0 {
		var secrets []v1.LocalObjectReference
		for _, secret := range cr.Spec.ImagePullSecrets {
			secrets = append(secrets, v1.LocalObjectReference{Name: secret})
		}
		spec.Template.Spec.ImagePullSecrets = secrets

	}
	if len(cr.Spec.MCMMonitor.ServiceAccountName) != 0 {
		spec.Template.Spec.ServiceAccountName = cr.Spec.MCMMonitor.ServiceAccountName
	}

	//container
	container, err := mcmContainer(cr)
	if err != nil {
		return nil, err
	}
	spec.Template.Spec.Containers = []v1.Container{*container}

	return spec, nil

}

func mcmContainer(cr *promext.PrometheusExt) (*v1.Container, error) {
	drops := []v1.Capability{"ALL"}
	pe := false
	p := false
	prometheus, perr := NewPrometheus(cr)
	prometheus.Spec.PodMetadata.CreationTimestamp = *creationTime
	if perr != nil {
		return nil, perr
	}
	prometheus.Name = "ibm-monitoring-prometheus-hub"
	prometheus.ObjectMeta.Labels[Component] = hubPromemetheus
	prometheus.Spec.RuleSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{Component: hubPromemetheus},
	}
	prometheus.Spec.Storage = nil
	prometheus.Spec.AdditionalScrapeConfigs = nil
	prometheus.Spec.PodMetadata.Labels[AppLabelKey] = hubPromemetheus
	prometheusStr, merr := json.Marshal(prometheus)
	if merr != nil {
		return nil, merr
	}

	container := &v1.Container{
		Name:            "mcm",
		Image:           cr.Spec.MCMMonitor.Image,
		ImagePullPolicy: cr.Spec.ImagePolicy,
		SecurityContext: &v1.SecurityContext{
			AllowPrivilegeEscalation: &pe,
			Privileged:               &p,
			Capabilities: &v1.Capabilities{
				Drop: drops,
			},
		},
		Resources: cr.Spec.MCMMonitor.Resources,
		Env: []v1.EnvVar{
			{
				Name:  "NAMESPACES",
				Value: cr.Namespace,
			},
			{
				Name:  "NAMESPACE",
				Value: cr.Namespace,
			},
			{
				Name:  "IS_HUB_CLUSTER",
				Value: fmt.Sprint(cr.Spec.MCMMonitor.IsHubCluster),
			},
			{
				Name:  "GRAFANA_BASE_URL",
				Value: fmt.Sprintf("https://%s:%d/", cr.Spec.GrafanaSvcName, cr.Spec.GrafanaSvcPort),
			},
			{
				Name:  "PROMETHEUS_YAML",
				Value: string(prometheusStr),
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "monitoring-ca-certs",
				MountPath: "/opt/ibm/monitoring/caCerts",
			},
			{
				Name:      "monitoring-client-certs",
				MountPath: "/opt/ibm/monitoring/certs",
			},
		},
	}
	return container, nil

}
func msmcCtrlLabels(cr *promext.PrometheusExt) map[string]string {
	labels := make(map[string]string)
	labels[AppLabelKey] = AppLabelValue
	labels[Component] = "mcm-ctl"
	labels[HealthCheckKey] = HealthCheckLabelValue
	labels[managedLabelKey()] = managedLabelValue(cr)
	for key, v := range cr.Labels {
		labels[key] = v
	}
	return labels

}
