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

	//Host value of route cp-console
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ClusterAddress string `json:"clusterAddress,omitempty"`
	//Port value of route cp-console
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ClusterPort int32 `json:"clusterPort,omitempty"`
	//Cluster name, mycluster by default
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ClusterName string `json:"clusterName,omitempty"`
	//Cluster domain name, cluster.local by default
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// Image pull policy
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ImagePolicy v1.PullPolicy `json:"imagePolicy,omitempty"`
	// Extra image pull secrets
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	//Configurations for alertmanager
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	AlertManagerConfig `json:"alertManagerConfig"`
	//Configurations for prometheus
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	PrometheusConfig `json:"prometheusConfig"`
	//repo:tag for router image
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	RouterImage string `json:"routerImage"`
	//Storage class name used by Prometheus and Alertmanager
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	StorageClassName string `json:"storageClassName"`
	//Configurations for mcm monitor controller
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	MCMMonitor `json:"mcmMonitor,omitempty"`
	//Configurations for tls certification
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Certs `json:"certs"`
	//Configurations for IAM Provider
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	IAMProvider `json:"iamProvider"`
	//Grafana service name trusted by prometheus
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	GrafanaSvcName string `json:"grafanaSvcName"`
	//Grafana service port truested by prometheus
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	GrafanaSvcPort int32 `json:"grafanaSvcPort"`
	//Helm API service information
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	HelmReleasesMonitor `json:"helmReleasesMonitor,omitempty"`

	PrometheusOperator `json:"prometheusOperator,omitempty"`
}

//PrometheusOperator defines inforamtion for prometheus operator deployment
type PrometheusOperator struct {
	//Image of prometheus
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Image string `json:"image,omitempty"`
	//Image of configmap reloader
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ConfigmapReloadImage string `json:"configmapReloadImage,omitempty"`
	//Image of prometheus config reloader
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	PrometheusConfigImage string `json:"prometheusConfigImage,omitempty"`
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
	Image               string                  `json:"image,omitempty"`
	ImageRepo           string                  `json:"imageRepo,omitempty"`
	ImageTag            string                  `json:"imageTag,omitempty"`
	ImageSHA            string                  `json:"imageSHA,omitempty"`
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
	Image              string                  `json:"image,omitempty"`
	ImageRepo          string                  `json:"imageRepo,omitempty"`
	ImageTag           string                  `json:"imageTag,omitempty"`
	ImageSHA           string                  `json:"imageSHA,omitempty"`
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

	//Status of prometheus operator deployment
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	PrometheusOperator string `json:"prometheusOperator,omitempty"`
	//Status of the prometheus CR, created or not
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Prometheus string `json:"prometheus,omitempty"`
	//Status of the alert manager CR, created or not
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Alertmanager string `json:"alertmanager,omitempty"`
	//Status of the exporter CR, created or not
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Exporter string `json:"exporter,omitempty"`
	//Status of required secrets, created or not
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Secrets string `json:"secrets,omitempty"`
	//Status of required configmaps, created or not
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Configmaps string `json:"configmaps,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrometheusExt will start Prometheus and Alertmanager instances with RBAC enabled. It will also enable Multicloud monitoring
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=prometheusexts,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Prometheus Operator Extension"
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
