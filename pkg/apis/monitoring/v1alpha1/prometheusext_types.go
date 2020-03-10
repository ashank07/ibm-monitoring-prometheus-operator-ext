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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PrometheusExtSpec defines the desired state of PrometheusExt
type PrometheusExtSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ClusterAddress     string        `json:"clusterAddress"`
	ClusterPort        int32         `json:"clusterPort"`
	ClusterName        string        `json:"clusterName,omitempty"`
	ClusterDomain      string        `json:"clusterDomain,omitempty"`
	ImagePolicy        v1.PullPolicy `json:"imagePolicy,omitempty"`
	ImagePullSecrets   []string      `json:"imagePullSecrets,omitempty"`
	AlertManagerConfig `json:"alertManagerConfig"`
	PrometheusConfig   `json:"prometheusConfig"`
	RouterImage        string `json:"routerImage"`
	StorageClassName   string `json:"storageClassName"`
	MCMMonitor         `json:"mcmMonitor,omitempty"`
	Certs              `json:"certs"`
	IAMProvider        `json:"iamProvider"`
	//Grafana integrated with this CR
	GrafanaSvcName      string `json:"grafanaSvcName"`
	GrafanaSvcPort      int32  `json:"grafanaSvcPort"`
	HelmReleasesMonitor `json:"helmReleasesMonitor,omitempty"`
}

//IAMProvider defines information for iam
type IAMProvider struct {
	Namespace           string `json:"namespace,omitempty"`
	IDProviderSvc       string `json:"idProviderSvc"`
	IDProviderSvcPort   int32  `json:"idProviderSvcPort"`
	IDManagementSvc     string `json:"idManagementSvc"`
	IDManagementSvcPort int32  `json:"idManagementSvcPort"`
}

//HelmReleasesMonitor defines information for heml releases monitoring
type HelmReleasesMonitor struct {
	Namespace string `json:"namespace,omitempty"`
	Port      int32  `json:"port,omitempty"`
}

// PrometheusConfig defines configuration of Prometheus object
type PrometheusConfig struct {
	ServiceAccountName  string                  `json:"serviceAccount,omitempty"`
	ImageRepo           string                  `json:"imageRepo"`
	ImageTag            string                  `json:"imageTag"`
	Retention           string                  `json:"retention,omitempty"`
	ScrapeInterval      string                  `json:"scrapeInterval,omitempty"`
	EvaluationInterval  string                  `json:"evaluationInterval,omitempty"`
	Resources           v1.ResourceRequirements `json:"resource,omitempty"`
	RouterResource      v1.ResourceRequirements `json:"routerResource,omitempty"`
	PVSize              string                  `json:"pvSize,omitempty"`
	ServicePort         int32                   `json:"servicePort"`
	NodeMemoryThreshold int                     `json:"nodeMemoryThreshold"`
	NodeCPUThreshold    int                     `json:"nodeCPUThreshold"`
	LogLevel            string                  `json:"logLevel,omitempty"`
}

// AlertManagerConfig defines configuration of AlertManager object
type AlertManagerConfig struct {
	ServiceAccountName string                  `json:"serviceAccount,omitempty"`
	ImageRepo          string                  `json:"imageRepo"`
	ImageTag           string                  `json:"imageTag"`
	PVSize             string                  `json:"pvSize,omitempty"`
	ServicePort        int32                   `json:"servicePort"`
	Resources          v1.ResourceRequirements `json:"resource,omitempty"`
	LogLevel           string                  `json:"logLevel,omitempty"`
}

// Certs defines certification used by monitoring stack
type Certs struct {
	// Prometheus and AlertManager' tls cert. Define the secret name. It is created by cert manager
	MonitoringSecret string `json:"monitoringSecret"`
	//Define monitoring stack client(prometheus, exporters)'s tls cert secret. It is created by cert manager
	MonitoringClientSecret string `json:"monitoringClientSecret"`
	// The issure name. It is used to generated tls certificates. All tls certificates of monitoring operators need to use same Issuer
	Issuer string `json:"issuer"`
	// If it is false, user can create secret manually before creating CR and operator will not recreate it if secret exists already
	// If it is true, operator will recreate secret if it is not created by certificate (cert-manager)
	AutoClean bool `json:"autoClean,omitempty"`
}

// MCMMonitor defines multimple cloud monitoring related information
type MCMMonitor struct {
	// If it is running on MCM Hub cluster or not
	IsHubCluster bool `json:"isHubCluster,omitempty"`
	// Image for mcm monitoring controller
	Image string `json:"image,omitempty"`
	// MCM helper image for some initiallizing work
	HelperImage        string                  `json:"helpeImage,omitempty"`
	ServiceAccountName string                  `json:"serviceAccount,omitempty"`
	Resources          v1.ResourceRequirements `json:"resource,omitempty"`
}

// PrometheusExtStatus defines the observed state of PrometheusExt
type PrometheusExtStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrometheusExt will start Prometheus and Alertmanager instances with RBAC enabled. It will also enable Multicloud monitoring
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=prometheusexts,scope=Namespaced
type PrometheusExt struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrometheusExtSpec   `json:"spec,omitempty"`
	Status PrometheusExtStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrometheusExtList contains a list of PrometheusExt
type PrometheusExtList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrometheusExt `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrometheusExt{}, &PrometheusExtList{})
}
