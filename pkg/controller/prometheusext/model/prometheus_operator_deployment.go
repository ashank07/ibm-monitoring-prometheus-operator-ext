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
	"os"

	"github.com/prometheus/common/log"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	promext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

//PrometheusOperatorName return name of prometheus operator deployment
func PrometheusOperatorName(cr *promext.PrometheusExt) string {
	return cr.Name + "-prometheus-operator"

}

//NewProOperatorDeployment create new deployment for prometheus operator
func NewProOperatorDeployment(cr *promext.PrometheusExt) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PrometheusOperatorName(cr),
			Namespace: cr.Namespace,
			Labels:    proOperatorLabels(cr),
		},
		Spec: *promeDeploymentSpec(cr),
	}

	return deployment
}

func promeDeploymentSpec(cr *promext.PrometheusExt) *appsv1.DeploymentSpec {
	spec := &appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: proOperatorLabels(cr),
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:        PrometheusOperatorName(cr),
				Labels:      proOperatorLabels(cr),
				Annotations: commonPodAnnotations(),
			},
			Spec: v1.PodSpec{
				HostPID:            false,
				HostIPC:            false,
				HostNetwork:        false,
				ServiceAccountName: "prometheus-operator",
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

	//container
	container := prometneusOperatorContainer(cr)
	spec.Template.Spec.Containers = []v1.Container{*container}

	return spec

}
func prometneusOperatorContainer(cr *promext.PrometheusExt) *v1.Container {
	pe := false
	p := false
	var cpuLimit resource.Quantity
	var memLimit resource.Quantity
	var cpuReq resource.Quantity
	var memReq resource.Quantity
	var err error
	cpuLimit, err = resource.ParseQuantity("200m")
	if err != nil {
		log.Error(err, "")
	}
	memLimit, err = resource.ParseQuantity("256Mi")
	if err != nil {
		log.Error(err, "")
	}
	cpuReq, err = resource.ParseQuantity("100m")
	if err != nil {
		log.Error(err, "")
	}
	memReq, err = resource.ParseQuantity("50Mi")
	if err != nil {
		log.Error(err, "")
	}

	container := &v1.Container{
		Name:            "prometheus-operator",
		Image:           *imageName(os.Getenv(promeOPImageEnv), cr.Spec.PrometheusOperator.Image),
		ImagePullPolicy: cr.Spec.ImagePolicy,
		Args: []string{
			"-namespaces=" + cr.Namespace,
			"-manage-crds=false",
			"-logtostderr=true",
			"--config-reloader-image=" + *imageName(os.Getenv(cmReloadImageEnv), cr.Spec.PrometheusOperator.ConfigmapReloadImage),
			"--prometheus-config-reloader=" + *imageName(os.Getenv(promeConfImageEnv), cr.Spec.PrometheusConfigImage),
		},
		Env: []v1.EnvVar{
			{
				Name:  "NAMESPACE",
				Value: cr.Namespace,
			},
		},
		Ports: []v1.ContainerPort{{
			Name:          "http",
			ContainerPort: 8080,
		}},
		SecurityContext: &v1.SecurityContext{
			AllowPrivilegeEscalation: &pe,
			Privileged:               &p,
		},
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{

				v1.ResourceCPU:    cpuLimit,
				v1.ResourceMemory: memLimit,
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    cpuReq,
				v1.ResourceMemory: memReq,
			},
		},
	}
	return container

}

//UpdatedProOperatorDeployment create new deployment for prometheus operator
func UpdatedProOperatorDeployment(cr *promext.PrometheusExt, curr *appsv1.Deployment) *appsv1.Deployment {
	deployment := curr.DeepCopy()
	deployment.ObjectMeta.Labels = proOperatorLabels(cr)
	ann := commonPodAnnotations()
	for k, v := range ann {
		deployment.ObjectMeta.Annotations[k] = v

	}
	spec := promeDeploymentSpec(cr)
	deployment.Spec.Template.ObjectMeta.Labels = spec.Template.ObjectMeta.Labels
	deployment.Spec.Template.ObjectMeta.Annotations = spec.Template.ObjectMeta.Annotations
	deployment.Spec.Template.Spec.Containers = spec.Template.Spec.Containers

	return deployment
}

func proOperatorLabels(cr *promext.PrometheusExt) map[string]string {
	labels := make(map[string]string)
	labels[AppLabelKey] = AppLabelValue
	labels[Component] = "prometheus-operator"
	labels[HealthCheckKey] = HealthCheckLabelValue
	labels[managedLabelKey()] = managedLabelValue(cr)
	for key, v := range cr.Labels {
		labels[key] = v
	}
	return labels

}
