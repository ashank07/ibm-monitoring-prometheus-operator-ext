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
	"os"

	"html/template"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	promext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

//ProRouterNgCmName returns prometheus router nginx configmap name
func ProRouterNgCmName(cr *promext.PrometheusExt) string {
	return cr.Name + "-prometheus-router-ng"
}

//AlertRouterNgCmName returns alertmanaget router nginx configmap name
func AlertRouterNgCmName(cr *promext.PrometheusExt) string {
	return cr.Name + "-alertmanager-router-ng"
}

//RouterEntryCmName returns router entrypoint configmap name
func RouterEntryCmName(cr *promext.PrometheusExt) string {
	return cr.Name + "-prometheus-router-entry"
}

type proRouterNgParas struct {
	Managed    bool //is it managed instance? true for now
	Openshift  bool //installed on top openshift? true for now
	Standalone bool
}

//NewAlertmanagerRouterNgCm returns configmap for router nginx
func NewAlertmanagerRouterNgCm(cr *promext.PrometheusExt) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AlertRouterNgCmName(cr),
			Namespace: cr.Namespace,
			Labels:    alertmanagerLabels(cr),
		},
		Data: map[string]string{"nginx.conf": alertRouterConfig},
	}
	return cm

}

//UpdatedAlertRouterNgcm update configmap for alertmanager router nginx onfig
func UpdatedAlertRouterNgcm(cr *promext.PrometheusExt, curr *v1.ConfigMap) *v1.ConfigMap {
	cm := curr.DeepCopy()
	cm.Labels = alertmanagerLabels(cr)
	cm.Data = map[string]string{"nginx.conf": alertRouterConfig}
	return cm
}

//NewProRouterNgCm returns configmap for router nginx
func NewProRouterNgCm(cr *promext.PrometheusExt) (*v1.ConfigMap, error) {
	var tplBuffer bytes.Buffer
	paras := proRouterNgParas{
		Managed:    true,
		Openshift:  true,
		Standalone: !cr.Spec.MCMMonitor.IsHubCluster,
	}
	if err := prometheusNgConfTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProRouterNgCmName(cr),
			Namespace: cr.Namespace,
			Labels:    PrometheusLabels(cr),
		},
		Data: map[string]string{"nginx.conf": tplBuffer.String()},
	}
	return cm, nil
}

//ProLuaCmName return configmap name for prometheus lua script
func ProLuaCmName(cr *promext.PrometheusExt) string {
	return cr.Name + "-prometheus-lua"
}

//ProLuaUtilsCmName return configmap name for prometheus lua utils script
func ProLuaUtilsCmName(cr *promext.PrometheusExt) string {
	return cr.Name + "-prometheus-lua-utils"
}

//UpdatedProLuaCm return cupdated onfigmap for prometheus lua script
func UpdatedProLuaCm(cr *promext.PrometheusExt, curr *v1.ConfigMap) (*v1.ConfigMap, error) {
	cm := curr.DeepCopy()
	var tplBuffer bytes.Buffer
	helmNamespace := cr.Namespace
	if cr.Spec.HelmReleasesMonitor.Namespace != "" {
		helmNamespace = cr.Spec.HelmReleasesMonitor.Namespace
	}
	helmPort := defaultHelmPort
	if cr.Spec.HelmReleasesMonitor.Port != 0 {
		helmPort = cr.Spec.HelmReleasesMonitor.Port

	}
	clusterDomain := defaultClusterDomain
	if cr.Spec.ClusterDomain != "" {
		clusterDomain = cr.Spec.ClusterDomain
	}
	paras := luaParas{
		Standalone:          !cr.Spec.MCMMonitor.IsHubCluster,
		Managed:             true,
		Openshift:           true,
		AlertmanagerSvcName: AlertmanagerName(cr),
		AlertmanagerSvcPort: fmt.Sprintf("%d", cr.Spec.AlertManagerConfig.ServicePort),
		HelmNamespace:       helmNamespace,
		HelmPort:            helmPort,
		ClusterDomain:       clusterDomain,
	}
	if err := prometheusLuaTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	cm.Labels = PrometheusLabels(cr)
	cm.Data = map[string]string{"prom.lua": tplBuffer.String()}
	return cm, nil
}

type luaParas struct {
	Openshift           bool
	Managed             bool
	Standalone          bool
	AlertmanagerSvcName string
	AlertmanagerSvcPort string
	HelmNamespace       string
	HelmPort            int32
	ClusterDomain       string
}

//NewProLuaCm return configmap for prometheus lua script
func NewProLuaCm(cr *promext.PrometheusExt) (*v1.ConfigMap, error) {

	var tplBuffer bytes.Buffer

	helmNamespace := cr.Namespace
	if cr.Spec.HelmReleasesMonitor.Namespace != "" {
		helmNamespace = cr.Spec.HelmReleasesMonitor.Namespace
	}
	helmPort := defaultHelmPort
	if cr.Spec.HelmReleasesMonitor.Port != 0 {
		helmPort = cr.Spec.HelmReleasesMonitor.Port

	}
	clusterDomain := defaultClusterDomain
	if cr.Spec.ClusterDomain != "" {
		clusterDomain = cr.Spec.ClusterDomain
	}
	paras := luaParas{
		Standalone:          !cr.Spec.MCMMonitor.IsHubCluster,
		Managed:             true,
		Openshift:           true,
		AlertmanagerSvcName: AlertmanagerName(cr),
		AlertmanagerSvcPort: fmt.Sprintf("%d", cr.Spec.AlertManagerConfig.ServicePort),
		HelmNamespace:       helmNamespace,
		HelmPort:            helmPort,
		ClusterDomain:       clusterDomain,
	}
	if err := prometheusLuaTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProLuaCmName(cr),
			Namespace: cr.Namespace,
			Labels:    PrometheusLabels(cr),
		},
		Data: map[string]string{"prom.lua": tplBuffer.String()},
	}
	return cm, nil
}

//UpdatedProLuaUtilsCm return updated configmap for prometheus lua utils script
func UpdatedProLuaUtilsCm(cr *promext.PrometheusExt, curr *v1.ConfigMap) (*v1.ConfigMap, error) {
	var tplBuffer bytes.Buffer
	iamNS := cr.Namespace
	if cr.Spec.IAMProvider.Namespace != "" {
		iamNS = cr.Spec.IAMProvider.Namespace
	}

	clusterName := defaultClusterName
	if cr.Spec.ClusterName != "" {
		clusterName = cr.Spec.ClusterName

	}
	clusterDomain := defaultClusterDomain
	if cr.Spec.ClusterDomain != "" {
		clusterDomain = cr.Spec.ClusterDomain

	}

	paras := luaUtilsParas{
		ClusterName:          clusterName,
		ClusterDomain:        clusterDomain,
		Namespace:            cr.Namespace,
		PrometheusSvcName:    PromethuesName(cr),
		PrometheusSvcPort:    fmt.Sprintf("%d", cr.Spec.PrometheusConfig.ServicePort),
		GrafanaSvcName:       cr.Spec.GrafanaSvcName,
		GrafanaSvcPort:       fmt.Sprintf("%d", cr.Spec.GrafanaSvcPort),
		IAMNamespace:         iamNS,
		IAMProviderSvcName:   cr.Spec.IAMProvider.IDProviderSvc,
		IAMProviderSvcPort:   fmt.Sprintf("%d", cr.Spec.IAMProvider.IDProviderSvcPort),
		IAMManagementSvcName: cr.Spec.IAMProvider.IDManagementSvc,
		IAMManagementSvcPort: fmt.Sprintf("%d", cr.Spec.IAMProvider.IDManagementSvcPort),
	}
	if err := prometheusLuaUtilsTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	cm := curr.DeepCopy()
	cm.Labels = PrometheusLabels(cr)
	cm.Data = map[string]string{"monitoring-util.lua": tplBuffer.String()}
	return cm, nil
}

type luaUtilsParas struct {
	ClusterName       string
	Namespace         string
	PrometheusSvcName string
	PrometheusSvcPort string
	ClusterDomain     string
	//TODO: are the two parameters right? same as NewProLuaCm
	GrafanaSvcName string
	GrafanaSvcPort string

	IAMNamespace         string
	IAMProviderSvcName   string //platform-identity-provider
	IAMProviderSvcPort   string //4300
	IAMManagementSvcName string //platform-identity-management
	IAMManagementSvcPort string //4500

}

//NewProLuaUtilsCm return configmap for prometheus lua utils script
func NewProLuaUtilsCm(cr *promext.PrometheusExt) (*v1.ConfigMap, error) {

	var tplBuffer bytes.Buffer
	clusterDomain := defaultClusterDomain
	if cr.Spec.ClusterDomain != "" {
		clusterDomain = cr.Spec.ClusterDomain

	}
	clusterName := "mycluster"
	if cr.Spec.ClusterName != "" {
		clusterName = cr.Spec.ClusterName

	}
	iamNS := cr.Namespace
	if cr.Spec.IAMProvider.Namespace != "" {
		iamNS = cr.Spec.IAMProvider.Namespace
	}
	paras := luaUtilsParas{
		ClusterDomain:        clusterDomain,
		ClusterName:          clusterName,
		Namespace:            cr.Namespace,
		PrometheusSvcName:    PromethuesName(cr),
		PrometheusSvcPort:    fmt.Sprintf("%d", cr.Spec.PrometheusConfig.ServicePort),
		GrafanaSvcName:       cr.Spec.GrafanaSvcName,
		GrafanaSvcPort:       fmt.Sprintf("%d", cr.Spec.GrafanaSvcPort),
		IAMNamespace:         iamNS,
		IAMProviderSvcName:   cr.Spec.IAMProvider.IDProviderSvc,
		IAMProviderSvcPort:   fmt.Sprintf("%d", cr.Spec.IAMProvider.IDProviderSvcPort),
		IAMManagementSvcName: cr.Spec.IAMProvider.IDManagementSvc,
		IAMManagementSvcPort: fmt.Sprintf("%d", cr.Spec.IAMProvider.IDManagementSvcPort),
	}
	if err := prometheusLuaUtilsTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProLuaUtilsCmName(cr),
			Namespace: cr.Namespace,
			Labels:    PrometheusLabels(cr),
		},
		Data: map[string]string{"monitoring-util.lua": tplBuffer.String()},
	}
	return cm, nil
}

//UpdatedProRouterNgCm creates updated configmap for prometheus router nginx config
func UpdatedProRouterNgCm(cr *promext.PrometheusExt, curr *v1.ConfigMap) (*v1.ConfigMap, error) {
	cm := curr.DeepCopy()
	var tplBuffer bytes.Buffer
	paras := proRouterNgParas{
		Managed:    true,
		Openshift:  true,
		Standalone: !cr.Spec.MCMMonitor.IsHubCluster,
	}
	if err := prometheusNgConfTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	cm.Labels = PrometheusLabels(cr)
	cm.Data = map[string]string{"nginx.conf": tplBuffer.String()}
	return cm, nil
}

type routerEntryParas struct {
	Openshift bool
	Managed   bool
}

//NewRouterEntryCm returns configmap for router entrypoint
func NewRouterEntryCm(cr *promext.PrometheusExt) (*v1.ConfigMap, error) {
	var tplBuffer bytes.Buffer
	paras := routerEntryParas{
		Managed:   true,
		Openshift: true,
	}
	if err := routerEntrypointTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RouterEntryCmName(cr),
			Namespace: cr.Namespace,
			Labels:    PrometheusLabels(cr),
		},
		Data: map[string]string{"entrypoint.sh": tplBuffer.String()},
	}, nil
}

//UpdatedRouterEntryCm returns updated configmap for router entrypoint
func UpdatedRouterEntryCm(cr *promext.PrometheusExt, curr *v1.ConfigMap) (*v1.ConfigMap, error) {

	var tplBuffer bytes.Buffer
	paras := routerEntryParas{
		Managed:   true,
		Openshift: true,
	}
	if err := routerEntrypointTemplate.Execute(&tplBuffer, paras); err != nil {
		return nil, err
	}
	cm := curr.DeepCopy()
	cm.Labels = PrometheusLabels(cr)
	cm.Data = map[string]string{"entrypoint.sh": tplBuffer.String()}

	return cm, nil
}

//NewRouterContainer returns router container
func NewRouterContainer(cr *promext.PrometheusExt, ot ObjectType) *v1.Container {
	rofs := false
	adds := []v1.Capability{"CHOWN", "NET_ADMIN", "NET_RAW", "LEASE", "SETGID", "SETUID"}

	container := &v1.Container{
		Name:            "router",
		Image:           *imageName(os.Getenv(routerImageEnv), cr.Spec.RouterImage),
		ImagePullPolicy: cr.Spec.ImagePolicy,
		SecurityContext: &v1.SecurityContext{
			ReadOnlyRootFilesystem: &rofs,
			Capabilities: &v1.Capabilities{
				Add: adds,
			},
		},
		Resources: cr.Spec.PrometheusConfig.RouterResource,
		Ports: []v1.ContainerPort{{
			Name:          "router",
			ContainerPort: 8080,
		}},
		Command: []string{"/bin/sh", "-c", "cp /opt/ibm/router/entry/entrypoint.sh /opt/ibm/router/; chmod 744 /opt/ibm/router/entrypoint.sh;exec /opt/ibm/router/entrypoint.sh"},
	}
	//probes are ready for prometheus only

	if ot == Prometheus {
		iamNS := cr.Namespace
		if cr.Spec.IAMProvider.Namespace != "" {
			iamNS = cr.Spec.IAMProvider.Namespace
		}
		clusterDomain := defaultClusterDomain
		if cr.Spec.ClusterDomain != "" {
			clusterDomain = cr.Spec.ClusterDomain

		}

		command := fmt.Sprintf("wget --spider --no-check-certificate -S 'https://%s.%s.svc.%s:%d/v1/info'",
			cr.Spec.IAMProvider.IDProviderSvc,
			iamNS,
			clusterDomain,
			cr.Spec.IAMProvider.IDProviderSvcPort)
		rprobe := &v1.Probe{
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{"sh", "-c", command},
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
		}
		lprobe := rprobe.DeepCopy()
		lprobe.PeriodSeconds = 20
		container.ReadinessProbe = rprobe
		container.LivenessProbe = lprobe
	}
	//
	container.VolumeMounts = []v1.VolumeMount{
		{
			Name:      "secret-" + cr.Spec.Certs.MonitoringSecret,
			MountPath: "/opt/ibm/router/caCerts",
		},
		{
			Name:      "secret-" + cr.Spec.Certs.MonitoringSecret,
			MountPath: "/opt/ibm/router/certs",
		},
		{
			Name:      "configmap-" + RouterEntryCmName(cr),
			MountPath: "/opt/ibm/router/entry",
		},
	}
	if ot == Prometheus {
		container.VolumeMounts = append(
			container.VolumeMounts,
			v1.VolumeMount{
				Name:      "configmap-" + ProRouterNgCmName(cr),
				MountPath: "/opt/ibm/router/conf",
			},
			v1.VolumeMount{
				Name:      "configmap-" + ProLuaUtilsCmName(cr),
				MountPath: "/opt/ibm/router/nginx/conf/monitoring-util.lua",
				SubPath:   "monitoring-util.lua",
			},
			v1.VolumeMount{
				Name:      "configmap-" + ProLuaCmName(cr),
				MountPath: "/opt/ibm/router/nginx/conf/prom.lua",
				SubPath:   "prom.lua",
			},
		)

	} else {
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "configmap-" + AlertRouterNgCmName(cr),
			MountPath: "/opt/ibm/router/conf",
		})

	}
	return container
}

var (
	routerEntrypointTemplate *template.Template
)

func init() {
	routerEntrypointTemplate = template.Must(template.New("entrypoint.sh").Parse(routerEntrypoint))
	scrapeTargetsTemplate = template.Must(template.New(scrapeTargetsFileName()).Parse(promeScrapeTargets))
	prometheusNgConfTemplate = template.Must(template.New("nginx.conf").Parse(prometheusRouterConfig))
	prometheusLuaTemplate = template.Must(template.New("prom.lua").Parse(luaScripts))
	prometheusLuaUtilsTemplate = template.Must(template.New("monitoring-util.lua").Parse(luaUtilsScripts))
}
