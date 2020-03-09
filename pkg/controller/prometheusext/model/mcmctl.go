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

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	promext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

//NewMCMCtlDeployment create new deployment object for mcm controller
func NewMCMCtlDeployment(cr *promext.PrometheusExt) (*appsv1.Deployment, error) {
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
	spec, err := mcmDeploymentSpec(cr)
	if err != nil {
		return nil, err
	}
	deployment := curr.DeepCopy()
	deployment.ObjectMeta.Labels = msmcCtrlLabels(cr)

	deployment.Spec = *spec
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
	configmaps := map[string]string{
		"/opt/ibm/router/conf":                           ProRouterNgCmName(cr),
		"/opt/ibm/router/entry":                          RouterEntryCmName(cr),
		"/opt/ibm/router/nginx/conf/prom.lua":            ProLuaCmName(cr),
		"/opt/ibm/router/nginx/conf/monitoring-util.lua": ProLuaUtilsCmName(cr),
	}
	cmStr, err := json.Marshal(configmaps)
	if err != nil {
		return nil, err
	}
	secrets := map[string]string{
		"/opt/ibm/router/caCerts": cr.Spec.MonitoringSecret,
		"/opt/ibm/router/certs":   cr.Spec.Certs.MonitoringSecret,
		"":                        cr.Spec.MonitoringClientSecret,
	}
	secretStr, err := json.Marshal(secrets)
	if err != nil {
		return nil, err
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
				Name:  "INIT_IMAGE",
				Value: cr.Spec.MCMMonitor.HelperImage,
			},
			{
				Name:  "NAMESPACES",
				Value: cr.Namespace,
			},
			{
				Name:  "NAMESPACE",
				Value: cr.Namespace,
			},
			{
				Name:  "CLUSTER_ADDRESS",
				Value: cr.Spec.ClusterAddress,
			},
			{
				Name:  "CLUSTER_PORT",
				Value: fmt.Sprint(cr.Spec.ClusterPort),
			},
			{
				Name:  "IS_HUB_CLUSTER",
				Value: fmt.Sprint(cr.Spec.MCMMonitor.IsHubCluster),
			},
			{
				Name:  "ALERTMANAGER_NAME",
				Value: AlertmanagerName(cr),
			},
			{
				Name:  "CONFIGMAPS",
				Value: string(cmStr),
			},
			{
				Name:  "SECRETS",
				Value: string(secretStr),
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
	labels[AppLabelKey] = AppLabekValue
	labels[Component] = "mcm-ctl"
	labels[MeteringLabelKey] = MetringLabelValue
	labels[managedLabelKey()] = managedLabelValue(cr)
	for key, v := range cr.Labels {
		labels[key] = v
	}
	return labels

}
