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
	"k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/model"
)

func (r *Reconsiler) syncAlertmanager() error {
	//config secret
	secret := model.AlertmanagerConfigSecret(r.CR)
	if err := r.Client.Get(r.Context, client.ObjectKey{Name: secret.Name, Namespace: r.CR.Namespace}, secret); err != nil {
		if errors.IsNotFound(err) {
			if err = r.createObject(secret); err != nil {
				log.Error(err, "Failed to create secret for alertmanager configration")
				return err
			}
		} else {
			return err
		}

	}
	log.Info("alertmanager configration secret is sync")
	//alertmanager
	if r.CurrentState.ManagedAlertmanager == nil {
		am, err := model.NewAlertmanager(r.CR)
		if err != nil {
			log.Error(err, "Failed to create alertmanager object")
			return err
		}
		if err = r.createObject(am); err != nil {
			log.Error(err, "Failed to create alertmanager in cluster")
			return err
		}

	} else {
		am, err := model.UpdatedAlertmanager(r.CR, r.CurrentState.ManagedAlertmanager)
		if err != nil {
			log.Error(err, "failed to create updated alertmanager object")
			return err
		}
		if err = r.updateObject(am); err != nil {
			log.Error(err, "Failed to update alertmanager object in cluster")
			return err
		}

	}
	log.Info("alertmanager object is sync")
	//service
	if r.CurrentState.AlertmanagerSvc == nil {
		svc := model.NewAlertmanagetSvc(r.CR)
		if err := r.createObject(svc); err != nil {
			log.Error(err, "Failed to create service for alertmanager in cluster")
			return err
		}

	} else {
		svc := model.UpdatedAlertmanagetSvc(r.CR, r.CurrentState.AlertmanagerSvc)
		if err := r.updateObject(svc); err != nil {
			log.Error(err, "Failed to update alertmanager service in cluster")
			return err
		}

	}
	log.Info("alertmanager service object is sync")
	//ingress
	if r.CurrentState.AlertManagerIngress == nil {
		ingress := model.NewAlertmanagerIngress(r.CR)
		if err := r.createObject(ingress); err != nil {
			log.Error(err, "Failed to create ingress for alertmanager in cluster")
			return err
		}

	} else {
		ingress := model.UpdatedAlertmanagetIngress(r.CR, r.CurrentState.AlertManagerIngress)
		if err := r.updateObject(ingress); err != nil {
			log.Error(err, "Failed to update alertmanager ingress in cluster")
			return err
		}

	}
	log.Info("alertmanager ingress object is sync")

	return nil
}
