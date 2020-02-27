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
	"fmt"
	"time"

	promext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	ev1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//AlertmanagerName returns name of alertmanager objects
func AlertmanagerName(cr *promext.PrometheusExt) string {
	return ObjectName(cr, Alertmanager)
}
func alertmanagerLabels(cr *promext.PrometheusExt) map[string]string {
	labels := make(map[string]string)
	labels[AppLabelKey] = AppLabekValue
	labels["component"] = "alertmanager"
	labels[managedLabelKey()] = managedLabelValue(cr)
	for key, v := range cr.Labels {
		labels[key] = v
	}
	return labels
}

//AlertmanagerConfigSecret create secret object to config alertmanager
func AlertmanagerConfigSecret(cr *promext.PrometheusExt) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alertmanager-" + AlertmanagerName(cr),
			Namespace: cr.Namespace,
			Labels:    alertmanagerLabels(cr),
		},
		Data: map[string][]byte{"alertmanager.yaml": []byte(alertConfigStr)},
	}
	return secret
}

//NewAlertmanager create Alertmanager object
func NewAlertmanager(cr *promext.PrometheusExt) (*promv1.Alertmanager, error) {
	replicas := int32(1)
	pvsize := DefaultPVSize
	if cr.Spec.AlertManagerConfig.PVSize != "" {
		pvsize = cr.Spec.AlertManagerConfig.PVSize
	}
	quantity, err := resource.ParseQuantity(pvsize)
	if err != nil {
		return nil, err
	}
	am := &promv1.Alertmanager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AlertmanagerName(cr),
			Namespace: cr.Namespace,
			Labels:    alertmanagerLabels(cr),
		},
		Spec: promv1.AlertmanagerSpec{
			PodMetadata: &metav1.ObjectMeta{
				Labels:            alertmanagerLabels(cr),
				Annotations:       map[string]string{"pvJob": "true"},
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
			BaseImage:   cr.Spec.AlertManagerConfig.ImageRepo,
			Version:     cr.Spec.AlertManagerConfig.ImageTag,
			Replicas:    &replicas,
			Resources:   cr.Spec.AlertManagerConfig.Resources,
			Secrets:     []string{cr.Spec.Certs.MonitoringSecret, cr.Spec.Certs.MonitoringClientSecret, cr.Spec.CASecret},
			ConfigMaps:  []string{RouterEntryCmName(cr), AlertRouterNgCmName(cr)},
			RoutePrefix: "/alertmanager",
			Containers:  []v1.Container{*NewRouterContainer(cr, Alertmanager)},
			Storage: &promv1.StorageSpec{
				VolumeClaimTemplate: v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: metav1.Time{Time: time.Now()},
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes:      []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
						StorageClassName: &cr.Spec.StorageClassName,
						Resources: v1.ResourceRequirements{
							Requests: map[v1.ResourceName]resource.Quantity{"storage": quantity},
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
		am.Spec.ImagePullSecrets = secrets

	}
	externalHost := LoopBackIP
	externalPort := ExternalPort
	if cr.Spec.ClusterAddress != "" {
		externalHost = cr.Spec.ClusterAddress
	}
	if cr.Spec.ClusterPort != 0 {
		externalPort = fmt.Sprintf("%d", cr.Spec.ClusterPort)
	}
	am.Spec.ExternalURL = "https://" + externalHost + ":" + externalPort + "/alertmanager"
	if cr.Spec.AlertManagerConfig.ServiceAccountName != "" {
		am.Spec.ServiceAccountName = cr.Spec.AlertManagerConfig.ServiceAccountName
	}

	return am, nil
}

//UpdatedAlertmanager create updated Alertmanager object
func UpdatedAlertmanager(cr *promext.PrometheusExt, curr *promv1.Alertmanager) (*promv1.Alertmanager, error) {
	pvsize := "10Gi"
	if cr.Spec.AlertManagerConfig.PVSize != "" {
		pvsize = cr.Spec.AlertManagerConfig.PVSize
	}
	quantity, err := resource.ParseQuantity(pvsize)
	if err != nil {
		return nil, err
	}

	am := curr.DeepCopy()
	am.Labels = alertmanagerLabels(cr)
	am.Spec.PodMetadata.Labels = alertmanagerLabels(cr)
	am.Spec.PodMetadata.Annotations = map[string]string{"pvJob": "true"}
	am.Spec.BaseImage = cr.Spec.AlertManagerConfig.ImageRepo
	am.Spec.Version = cr.Spec.AlertManagerConfig.ImageTag
	am.Spec.Resources = cr.Spec.AlertManagerConfig.Resources
	am.Spec.Secrets = []string{cr.Spec.Certs.MonitoringSecret, cr.Spec.Certs.MonitoringClientSecret, cr.Spec.CASecret}
	am.Spec.ConfigMaps = []string{RouterEntryCmName(cr), AlertRouterNgCmName(cr)}
	am.Spec.Containers = []v1.Container{*NewRouterContainer(cr, Alertmanager)}
	am.Spec.Storage = &promv1.StorageSpec{
		VolumeClaimTemplate: v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: curr.Spec.Storage.VolumeClaimTemplate.CreationTimestamp,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes:      []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				StorageClassName: &cr.Spec.StorageClassName,
				Resources: v1.ResourceRequirements{
					Requests: map[v1.ResourceName]resource.Quantity{"storage": quantity},
				},
			},
		},
	}

	if cr.Spec.ImagePullSecrets != nil && len(cr.Spec.ImagePullSecrets) != 0 {
		var secrets []v1.LocalObjectReference
		for _, secret := range cr.Spec.ImagePullSecrets {
			secrets = append(secrets, v1.LocalObjectReference{Name: secret})
		}
		am.Spec.ImagePullSecrets = secrets

	}
	externalHost := LoopBackIP
	externalPort := ExternalPort
	if cr.Spec.ClusterAddress != "" {
		externalHost = cr.Spec.ClusterAddress
	}
	if cr.Spec.ClusterPort != 0 {
		externalPort = fmt.Sprintf("%d", cr.Spec.ClusterPort)
	}
	am.Spec.ExternalURL = "https://" + externalHost + ":" + externalPort + "/alertmanager"
	if cr.Spec.AlertManagerConfig.ServiceAccountName != "" {
		am.Spec.ServiceAccountName = cr.Spec.AlertManagerConfig.ServiceAccountName
	}

	return am, nil
}

//NewAlertmanagetSvc create Alertmanager service object
func NewAlertmanagetSvc(cr *promext.PrometheusExt) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AlertmanagerName(cr),
			Namespace: cr.Namespace,
			Labels:    alertmanagerLabels(cr),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "web",
					Protocol: "TCP",
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8443,
					},
					Port: cr.Spec.AlertManagerConfig.ServicePort,
				},
			},
			Selector: alertmanagerLabels(cr),
			Type:     v1.ServiceTypeClusterIP,
		},
	}
	return svc
}

//UpdatedAlertmanagetSvc create Alertmanager service object
func UpdatedAlertmanagetSvc(cr *promext.PrometheusExt, curr *v1.Service) *v1.Service {
	svc := curr.DeepCopy()
	svc.Labels = alertmanagerLabels(cr)
	svc.Spec.Selector = alertmanagerLabels(cr)
	svc.Spec.Ports[0].Port = cr.Spec.AlertManagerConfig.ServicePort
	return svc
}

//NewAlertmanagerIngress create ingress for managed alertmanager
func NewAlertmanagerIngress(cr *promext.PrometheusExt) *ev1beta1.Ingress {
	ingress := &ev1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        AlertmanagerName(cr),
			Namespace:   cr.Namespace,
			Labels:      alertmanagerLabels(cr),
			Annotations: ingressAnnotations(cr),
		},
		Spec: ev1beta1.IngressSpec{
			Rules: []ev1beta1.IngressRule{
				{
					IngressRuleValue: ev1beta1.IngressRuleValue{
						HTTP: &ev1beta1.HTTPIngressRuleValue{
							Paths: []ev1beta1.HTTPIngressPath{
								{
									Path: "/alertmanager",
									Backend: ev1beta1.IngressBackend{
										ServiceName: AlertmanagerName(cr),
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: cr.Spec.AlertManagerConfig.ServicePort,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ingress
}

//UpdatedAlertmanagetIngress create Alertmanager ingress object
func UpdatedAlertmanagetIngress(cr *promext.PrometheusExt, curr *ev1beta1.Ingress) *ev1beta1.Ingress {
	ingress := curr.DeepCopy()
	ingress.Labels = alertmanagerLabels(cr)
	ingress.Annotations = ingressAnnotations(cr)
	ingress.Spec = ev1beta1.IngressSpec{
		Rules: []ev1beta1.IngressRule{
			{
				IngressRuleValue: ev1beta1.IngressRuleValue{
					HTTP: &ev1beta1.HTTPIngressRuleValue{
						Paths: []ev1beta1.HTTPIngressPath{
							{
								Path: "/alertmanager",
								Backend: ev1beta1.IngressBackend{
									ServiceName: AlertmanagerName(cr),
									ServicePort: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: cr.Spec.AlertManagerConfig.ServicePort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}