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
	"github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/model"
	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconsiler) syncPrometheus() error {
	//scrape targets secret
	if r.CurrentState.PrometheusScrapeTargetsSecret == nil {
		secret, err := model.NewScrapeTargetsSecret(r.CR, r.CurrentState.Exporter)
		if err != nil {
			log.Error(err, "Faild to create secret object for prometheus scrape target")
			return err
		}
		if err = r.createObject(secret); err != nil {
			log.Error(err, "Faild to create secret object in kubernetes for prometheus scrape target")
			return err
		}

	} else {
		secret, err := model.UpdatedScrapeTargetsSecret(r.CR, r.CurrentState.Exporter, r.CurrentState.PrometheusScrapeTargetsSecret)
		if err != nil {
			return err
		}
		if err = r.updateObject(secret); err != nil {
			return err
		}
	}
	log.Info("prometheus scrape targets secret is sync")
	//Prometheus instance
	if r.CurrentState.ManagedPrometheus == nil {
		prometheus, err := model.NewPrometheus(r.CR)
		if err != nil {
			log.Error(err, "Failed to create prometheus object")
			return err
		}
		if err = r.createObject(prometheus); err != nil {
			log.Error(err, "Failed to create prometheus in cluster")
			return err
		}
	} else {
		prometheus, err := model.UpdatedPrometheus(r.CR, r.CurrentState.ManagedPrometheus)
		if err != nil {
			log.Error(err, "Failed to create updated prometheus object")
			return err
		}
		if err := r.updateObject(prometheus); err != nil {
			log.Error(err, "Failed to update prometheus object")
			return nil
		}

	}
	log.Info("prometheus object is sync")
	//service
	if r.CurrentState.PromeSvc == nil {
		svc := model.NewPrometheusSvc(r.CR)
		if err := r.createObject(svc); err != nil {
			return err
		}
	} else {
		svc := model.UpdatedPrometheusSvc(r.CR, r.CurrentState.PromeSvc)
		if err := r.updateObject(svc); err != nil {
			return nil
		}

	}
	log.Info("prometheus service object is sync")
	//ingress
	if r.CurrentState.PromeIngress == nil {
		ingress := model.NewPrometheusIngress(r.CR)
		if err := r.createObject(ingress); err != nil {
			return err
		}
	} else {
		ingress := model.UpdatedPrometheusIngress(r.CR, r.CurrentState.PromeIngress)
		if err := r.updateObject(ingress); err != nil {
			return nil
		}

	}
	log.Info("prometheus ingress object is sync")
	for name, rule := range model.DefaultPromethuesRules {
		remoteRule := &promv1.PrometheusRule{}
		k := client.ObjectKey{Name: name, Namespace: r.CR.Namespace}
		if err := r.Client.Get(r.Context, k, remoteRule); err != nil {
			if errors.IsNotFound(err) {
				remoteRule = rule.DeepCopy()
				remoteRule.ObjectMeta = metav1.ObjectMeta{
					Name:      name,
					Namespace: r.CR.Namespace,
					Labels:    model.PrometheusLabels(r.CR),
				}
				if err = r.createObject(remoteRule); err != nil {
					log.Error(err, "Failed to create PrometheusRule: "+name)
					return err

				}

			} else {
				log.Error(err, "Failed to get PrometheusRule: "+name)
				return err

			}

		}

	}
	log.Info("default prometheus rules are created")
	return nil
}
