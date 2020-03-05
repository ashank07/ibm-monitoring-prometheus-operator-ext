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

package reconsiler

import (
	promev1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	ev1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	exportersv1alpha1 "github.com/IBM/ibm-monitoring-exporters-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/model"
)

//readPrometheus reads Prometheus object and also its service and ingresse
func (r *Reconsiler) readPrometheus() error {
	//read prometheus cr
	prome := promev1.Prometheus{}
	key := client.ObjectKey{Name: model.PromethuesName(r.CR), Namespace: r.CR.Namespace}
	err := r.Client.Get(r.Context, key, &prome)
	if err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.ManagedPrometheus = nil
		} else {
			log.Error(err, "Failed to get prometheuse")
			return err
		}
	} else {

		r.CurrentState.ManagedPrometheus = &prome
	}
	//read ingress and service for the prometheus
	ingress := ev1beta1.Ingress{}
	svc := v1.Service{}
	key = client.ObjectKey{Namespace: r.CR.Namespace, Name: model.PromethuesName(r.CR)}
	if err := r.Client.Get(r.Context, key, &ingress); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.PromeIngress = nil

		} else {
			log.Error(err, "failed to get ingress of Prometheus "+prome.Name)
			return err
		}
	} else {
		r.CurrentState.PromeIngress = &ingress
	}
	if err := r.Client.Get(r.Context, key, &svc); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.PromeSvc = nil
		} else {
			log.Error(err, "failed to get service of Prometheus "+prome.Name)
			return err
		}
	} else {
		r.CurrentState.PromeSvc = &svc

	}
	//read scrape targets secret
	secret := &v1.Secret{}
	key = client.ObjectKey{Namespace: r.CR.Namespace, Name: model.ScrapeTargetsSecretName(r.CR)}
	if err := r.Client.Get(r.Context, key, secret); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.PrometheusScrapeTargetsSecret = nil
		} else {
			log.Error(err, "failed to get prometheus scrape targets secret  "+model.ScrapeTargetsSecretName(r.CR))
			return err
		}
	} else {
		r.CurrentState.PrometheusScrapeTargetsSecret = secret

	}
	// read prometheus nginx configmap
	cm := &v1.ConfigMap{}
	key = client.ObjectKey{Name: model.ProRouterNgCmName(r.CR), Namespace: r.CR.Namespace}
	if err := r.Client.Get(r.Context, key, cm); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.PromeNgCm = nil
		} else {
			log.Error(err, "failed to get prometheus nginx configmap  "+model.ProRouterNgCmName(r.CR))
			return err
		}

	} else {
		r.CurrentState.PromeNgCm = cm
	}
	return nil
}

//readAlertmanager reads Alertmanager and also its service and ingress
func (r *Reconsiler) readAlertmanager() error {
	am := promev1.Alertmanager{}
	key := client.ObjectKey{Name: model.AlertmanagerName(r.CR), Namespace: r.CR.Namespace}
	if err := r.Client.Get(r.Context, key, &am); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.ManagedAlertmanager = nil
		} else {
			log.Error(err, "Failed to get Alertmanagers")
			return err
		}
	} else {

		r.CurrentState.ManagedAlertmanager = &am
	}

	//read ingress and service for the alermanager
	ingress := ev1beta1.Ingress{}
	svc := v1.Service{}
	key = client.ObjectKey{Namespace: r.CR.Namespace, Name: model.AlertmanagerName(r.CR)}
	if err := r.Client.Get(r.Context, key, &ingress); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.AlertManagerIngress = nil
		} else {
			log.Error(err, "Failed to get ingress of Alertmanager "+am.Name)
			return err
		}

	} else {
		r.CurrentState.AlertManagerIngress = &ingress

	}
	if err := r.Client.Get(r.Context, key, &svc); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.AlertmanagerSvc = nil

		} else {
			log.Error(err, "Failed to get service of Alertmanager "+am.Name)
			return err
		}

	} else {
		r.CurrentState.AlertmanagerSvc = &svc

	}

	return nil
}
func (r *Reconsiler) readExporter() error {
	exporters := exportersv1alpha1.ExporterList{}
	if err := r.Client.List(r.Context, &exporters); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.Exporter = nil
			return nil
		}
		log.Error(err, "Failed to list exporters")
		return err
	}
	// We assume there is only or exporter CR
	r.CurrentState.Exporter = &(exporters.Items[0])
	return nil

}

func (r *Reconsiler) readMCMCtlDeployment() error {
	deployment := appsv1.Deployment{}
	key := client.ObjectKey{Namespace: r.CR.Namespace, Name: model.MCMCtlDeploymentName(r.CR)}
	if err := r.Client.Get(r.Context, key, &deployment); err != nil {
		if errors.IsNotFound(err) {
			r.CurrentState.MCMCtrlDeployment = nil
			return nil
		}
		log.Error(err, "Failed to get MCM monitoring controller deployment")
		return err
	}
	r.CurrentState.MCMCtrlDeployment = &deployment
	return nil
}
func (r *Reconsiler) readSecrets() error {
	secret := &v1.Secret{}
	//monitoring cert secret
	key := client.ObjectKey{Name: r.CR.Spec.MonitoringSecret, Namespace: r.CR.Namespace}
	if err := r.Client.Get(r.Context, key, secret); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get monitoring secet "+r.CR.Spec.MonitoringSecret)
			return err
		}
		r.CurrentState.MonitoringSecret = nil

	} else {
		r.CurrentState.MonitoringSecret = secret.DeepCopy()
	}

	//monitoring client cert secret
	key = client.ObjectKey{Name: r.CR.Spec.MonitoringClientSecret, Namespace: r.CR.Namespace}
	if err := r.Client.Get(r.Context, key, secret); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get monitoring client secet "+r.CR.Spec.MonitoringClientSecret)
			return err
		}
		r.CurrentState.MonitoringClientSecret = nil

	} else {
		r.CurrentState.MonitoringClientSecret = secret.DeepCopy()
	}

	return nil

}

func (r *Reconsiler) readRouterCms() error {
	key := client.ObjectKey{
		Name:      model.RouterEntryCmName(r.CR),
		Namespace: r.CR.Namespace,
	}
	cm := &v1.ConfigMap{}
	if err := r.Client.Get(r.Context, key, cm); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get router entrypoint configmap "+key.Name)
			return err
		}
		r.CurrentState.RouterEntryCm = nil

	} else {
		r.CurrentState.RouterEntryCm = cm.DeepCopy()
	}

	key.Name = model.ProRouterNgCmName(r.CR)
	if err := r.Client.Get(r.Context, key, cm); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get prometheus router nginx configmap "+key.Name)
			return err
		}
		r.CurrentState.PromeNgCm = nil

	} else {
		r.CurrentState.PromeNgCm = cm.DeepCopy()
	}

	key.Name = model.ProLuaCmName(r.CR)
	if err := r.Client.Get(r.Context, key, cm); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get prometheus lua scripts configmap "+key.Name)
			return err
		}
		r.CurrentState.ProLuaCm = nil

	} else {
		r.CurrentState.ProLuaCm = cm.DeepCopy()
	}

	key.Name = model.ProLuaUtilsCmName(r.CR)
	if err := r.Client.Get(r.Context, key, cm); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get prometheus lua utils scripts configmap "+key.Name)
			return err
		}
		r.CurrentState.ProLuaUtilsCm = nil

	} else {
		r.CurrentState.ProLuaUtilsCm = cm.DeepCopy()
	}

	key.Name = model.AlertRouterNgCmName(r.CR)
	if err := r.Client.Get(r.Context, key, cm); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "failed to get alertmanager nginx configmap "+key.Name)
			return err
		}
		r.CurrentState.AlertNgCm = nil

	} else {
		r.CurrentState.AlertNgCm = cm.DeepCopy()
	}

	return nil
}
