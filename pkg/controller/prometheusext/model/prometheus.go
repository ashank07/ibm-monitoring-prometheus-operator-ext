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
	"bytes"
	"fmt"
	"html/template"
	"time"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	ev1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	exportersv1alpha1 "github.com/IBM/ibm-monitoring-exporters-operator/pkg/apis/monitoring/v1alpha1"
	promext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

//NewPrometheus create new Prometheus object for cr
func NewPrometheus(cr *promext.PrometheusExt) (*promv1.Prometheus, error) {
	spec, err := prometheusSpec(cr)
	if err != nil {
		return nil, err
	}
	prometheus := &promv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PromethuesName(cr),
			Namespace: cr.Namespace,
			Labels:    PrometheusLabels(cr),
		},
		Spec: *spec,
	}
	return prometheus, nil
}

//UpdatedPrometheus update Prometheus object for cr
func UpdatedPrometheus(cr *promext.PrometheusExt, current *promv1.Prometheus) (*promv1.Prometheus, error) {
	spec, err := prometheusSpec(cr)
	if err != nil {
		return nil, err
	}
	prometheus := current.DeepCopy()
	prometheus.Labels = PrometheusLabels(cr)
	prometheus.Spec = *spec
	prometheus.Spec.PodMetadata.CreationTimestamp = current.Spec.PodMetadata.CreationTimestamp
	prometheus.Spec.Storage.VolumeClaimTemplate.CreationTimestamp = current.Spec.Storage.VolumeClaimTemplate.CreationTimestamp
	return prometheus, nil
}

//PromethuesName returns names for prometheus objects
func PromethuesName(cr *promext.PrometheusExt) string {
	return ObjectName(cr, Prometheus)
}

//NewPrometheusSvc create service for managed prometheus
func NewPrometheusSvc(cr *promext.PrometheusExt) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PromethuesName(cr),
			Namespace: cr.Namespace,
			Labels:    PrometheusLabels(cr),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8443,
					},
					Port: cr.Spec.PrometheusConfig.ServicePort,
				},
			},
			Selector: PrometheusLabels(cr),
			Type:     v1.ServiceTypeClusterIP,
		},
	}
	return svc
}

//UpdatedPrometheusSvc update service for managed prometheus
func UpdatedPrometheusSvc(cr *promext.PrometheusExt, currentSvc *v1.Service) *v1.Service {
	svc := currentSvc.DeepCopy()
	svc.Labels = PrometheusLabels(cr)
	svc.Spec.Ports[0].Port = cr.Spec.PrometheusConfig.ServicePort
	svc.Spec.Selector = PrometheusLabels(cr)
	return svc
}

//NewPrometheusIngress create ingress for managed prometheus
func NewPrometheusIngress(cr *promext.PrometheusExt) *ev1beta1.Ingress {
	ingress := &ev1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        PromethuesName(cr),
			Namespace:   cr.Namespace,
			Labels:      PrometheusLabels(cr),
			Annotations: ingressAnnotations(cr),
		},
		Spec: ev1beta1.IngressSpec{
			Rules: []ev1beta1.IngressRule{
				{
					IngressRuleValue: ev1beta1.IngressRuleValue{
						HTTP: &ev1beta1.HTTPIngressRuleValue{
							Paths: []ev1beta1.HTTPIngressPath{
								{
									Path: "/prometheus",
									Backend: ev1beta1.IngressBackend{
										ServiceName: PromethuesName(cr),
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: cr.Spec.PrometheusConfig.ServicePort,
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

//UpdatedPrometheusIngress update ingress for managed prometheus
func UpdatedPrometheusIngress(cr *promext.PrometheusExt, currentIngress *ev1beta1.Ingress) *ev1beta1.Ingress {
	ingress := currentIngress.DeepCopy()
	ingress.Labels = PrometheusLabels(cr)
	ingress.Annotations = ingressAnnotations(cr)
	ingress.Spec = ev1beta1.IngressSpec{
		Rules: []ev1beta1.IngressRule{
			{
				IngressRuleValue: ev1beta1.IngressRuleValue{
					HTTP: &ev1beta1.HTTPIngressRuleValue{
						Paths: []ev1beta1.HTTPIngressPath{
							{
								Path: "/prometheus",
								Backend: ev1beta1.IngressBackend{
									ServiceName: PromethuesName(cr),
									ServicePort: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: cr.Spec.PrometheusConfig.ServicePort,
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

//PrometheusLabels return labels for prometheus objects
func PrometheusLabels(cr *promext.PrometheusExt) map[string]string {
	labels := make(map[string]string)
	labels[AppLabelKey] = AppLabekValue
	labels[Component] = "prometheus"
	labels[HealthCheckKey] = HealthCheckLabelValue
	labels[managedLabelKey()] = managedLabelValue(cr)
	for key, v := range cr.Labels {
		labels[key] = v
	}
	return labels
}

func prometheusSpec(cr *promext.PrometheusExt) (*promv1.PrometheusSpec, error) {
	replicas := int32(1)
	pvsize := DefaultPVSize
	if cr.Spec.PrometheusConfig.PVSize != "" {
		pvsize = cr.Spec.PrometheusConfig.PVSize
	}
	quantity, err := resource.ParseQuantity(pvsize)
	if err != nil {
		return nil, err
	}
	spec := &promv1.PrometheusSpec{
		PodMetadata: &metav1.ObjectMeta{
			Labels:            PrometheusLabels(cr),
			Annotations:       commonPodAnnotations(),
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		BaseImage:              cr.Spec.PrometheusConfig.ImageRepo,
		Version:                cr.Spec.PrometheusConfig.ImageTag,
		Replicas:               &replicas,
		EnableAdminAPI:         true,
		Resources:              cr.Spec.PrometheusConfig.Resources,
		RoutePrefix:            "/prometheus",
		Secrets:                []string{cr.Spec.Certs.MonitoringSecret, cr.Spec.Certs.MonitoringClientSecret},
		ConfigMaps:             []string{ProRouterNgCmName(cr), RouterEntryCmName(cr), ProLuaCmName(cr), ProLuaUtilsCmName(cr)},
		ServiceMonitorSelector: &metav1.LabelSelector{MatchLabels: map[string]string{AppLabelKey: AppLabekValue}},
		AdditionalScrapeConfigs: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: ScrapeTargetsSecretName(cr),
			},
			Key: scrapeTargetsFileName(),
		},
		Containers: []v1.Container{*NewRouterContainer(cr, Prometheus)},
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
	}
	spec.Alerting = &promv1.AlertingSpec{
		Alertmanagers: []promv1.AlertmanagerEndpoints{
			{
				Namespace: cr.Namespace,
				Name:      AlertmanagerName(cr),
				Port:      intstr.IntOrString{Type: intstr.String, StrVal: "web"},
				Scheme:    "https",
				TLSConfig: &promv1.TLSConfig{
					CertFile:           "/etc/prometheus/secrets/" + cr.Spec.MonitoringClientSecret + "/tls.crt",
					KeyFile:            "/etc/prometheus/secrets/" + cr.Spec.MonitoringClientSecret + "/tls.key",
					InsecureSkipVerify: true,
				},
			},
		},
	}
	//Select all rules in current namespace
	//spec.RuleNamespaceSelector = &metav1.LabelSelector{}
	spec.RuleSelector = &metav1.LabelSelector{}

	if cr.Spec.ImagePullSecrets != nil && len(cr.Spec.ImagePullSecrets) != 0 {
		var secrets []v1.LocalObjectReference
		for _, secret := range cr.Spec.ImagePullSecrets {
			secrets = append(secrets, v1.LocalObjectReference{Name: secret})
		}
		spec.ImagePullSecrets = secrets

	}
	if cr.Spec.PrometheusConfig.Retention == "" {
		spec.Retention = "24h"
	} else {
		spec.Retention = cr.Spec.PrometheusConfig.Retention
	}
	if cr.Spec.PrometheusConfig.ScrapeInterval == "" {
		spec.ScrapeInterval = "1m"
	} else {
		spec.ScrapeInterval = cr.Spec.PrometheusConfig.ScrapeInterval
	}
	if cr.Spec.PrometheusConfig.EvaluationInterval == "" {
		spec.EvaluationInterval = "1m"
	} else {
		spec.EvaluationInterval = cr.Spec.PrometheusConfig.EvaluationInterval
	}
	externalHost := LoopBackIP
	externalPort := ExternalPort
	if cr.Spec.ClusterAddress != "" {
		externalHost = cr.Spec.ClusterAddress
	}
	if cr.Spec.ClusterPort != 0 {
		externalPort = fmt.Sprintf("%d", cr.Spec.ClusterPort)
	}
	spec.ExternalURL = "https://" + externalHost + ":" + externalPort + "/prometheus"
	if cr.Spec.PrometheusConfig.ServiceAccountName != "" {
		spec.ServiceAccountName = cr.Spec.PrometheusConfig.ServiceAccountName
	}

	if cr.Spec.PrometheusConfig.LogLevel != "" {
		spec.LogLevel = cr.Spec.PrometheusConfig.LogLevel
	}

	return spec, nil
}

//ScrapeTargetsSecretName return secret name for prometheus scrape targets
func ScrapeTargetsSecretName(cr *promext.PrometheusExt) string {
	return cr.Name + "-scrape-targets"
}

func scrapeTargetsFileName() string {
	return "scrape-targets.yml"
}

//NewScrapeTargetsSecret return secret for prometheus scrape targets
func NewScrapeTargetsSecret(cr *promext.PrometheusExt, exporter *exportersv1alpha1.Exporter) (*v1.Secret, error) {
	var tplBuffer bytes.Buffer

	clusterDomain := defaultClusterDomain
	if cr.Spec.ClusterDomain != "" {
		clusterDomain = cr.Spec.ClusterDomain

	}

	paras := scrapeTargetConfigParas{
		Standalone:       !cr.Spec.MCMMonitor.IsHubCluster,
		CASecretName:     cr.Spec.MonitoringSecret,
		ClientSecretName: cr.Spec.MonitoringClientSecret,
		NodeExporter:     exporter != nil && exporter.Spec.NodeExporter.Enable,
		ClusterDomain:    clusterDomain,
	}
	if err := scrapeTargetsTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ScrapeTargetsSecretName(cr),
			Namespace: cr.Namespace,
			Labels:    PrometheusLabels(cr),
		},
		Data: map[string][]byte{scrapeTargetsFileName(): tplBuffer.Bytes()}, //[]byte(tplBuffer.String())},
	}
	return secret, nil
}

//UpdatedScrapeTargetsSecret return secret for prometheus scrape targets
func UpdatedScrapeTargetsSecret(cr *promext.PrometheusExt, exporter *exportersv1alpha1.Exporter, currentSecret *v1.Secret) (*v1.Secret, error) {
	secret := currentSecret.DeepCopy()
	var tplBuffer bytes.Buffer
	clusterDomain := defaultClusterDomain
	if cr.Spec.ClusterDomain != "" {
		clusterDomain = cr.Spec.ClusterDomain

	}
	paras := scrapeTargetConfigParas{
		Standalone:       !cr.Spec.MCMMonitor.IsHubCluster,
		CASecretName:     cr.Spec.MonitoringSecret,
		ClientSecretName: cr.Spec.MonitoringClientSecret,
		NodeExporter:     exporter != nil && exporter.Spec.NodeExporter.Enable,
		ClusterDomain:    clusterDomain,
	}
	if err := scrapeTargetsTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	secret.Labels = PrometheusLabels(cr)
	secret.Data = map[string][]byte{scrapeTargetsFileName(): tplBuffer.Bytes()} //[]byte(tplBuffer.String())}
	return secret, nil
}

//NewPrometheusRules create default PrometheusRule objects
func NewPrometheusRules(cr *promext.PrometheusExt) []*promv1.PrometheusRule {
	return []*promv1.PrometheusRule{nil}
}

//scrapeTargetConfigParas defines parameters for scrape targets template
type scrapeTargetConfigParas struct {
	Standalone       bool
	CASecretName     string
	ClientSecretName string
	NodeExporter     bool
	ClusterDomain    string
}

var (
	scrapeTargetsTemplate      *template.Template
	prometheusNgConfTemplate   *template.Template
	prometheusLuaTemplate      *template.Template
	prometheusLuaUtilsTemplate *template.Template
)
