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
	cert "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

//NewCertitication new certification object
func NewCertitication(name string, cr *proext.PrometheusExt, dnsNames []string) *cert.Certificate {
	return &cert.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.Namespace,
			Labels:    getCertLabels(cr),
		},
		Spec: cert.CertificateSpec{
			SecretName: name,
			IssuerRef: cert.ObjectReference{
				Name: "cs-ca-issuer",
				Kind: cert.IssuerKind,
			},
			CommonName: AppLabelValue,
			DNSNames:   dnsNames,
		},
	}
}

func getCertLabels(cr *proext.PrometheusExt) map[string]string {
	lables := make(map[string]string)
	lables[AppLabelKey] = AppLabelValue
	for key, v := range cr.Labels {
		lables[key] = v
	}
	return lables
}

//MonitoringDNSNames get dns names for monitoring cert
func MonitoringDNSNames(cr *proext.PrometheusExt) []string {
	dnsNames := []string{}
	dnsNames = append(dnsNames,
		PromethuesName(cr),
		AlertmanagerName(cr),
		cr.Spec.GrafanaSvcName,
		PromethuesName(cr)+"."+cr.Namespace,
		AlertmanagerName(cr)+"."+cr.Namespace,
		cr.Spec.GrafanaSvcName+"."+cr.Namespace,
		"*."+cr.Namespace,
		"*."+cr.Namespace+".svc")

	return dnsNames

}
