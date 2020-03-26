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

import "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/model"

func (r *Reconsiler) syncProOperatorDeployment() error {

	if r.CurrentState.PrometheusOperatorDeployment == nil {
		deployment := model.NewProOperatorDeployment(r.CR)
		if err := r.createObject(deployment); err != nil {
			log.Error(err, "Failed to create deployment for prometheus operator in cluster")
			return err
		}
	} else {
		deployment := model.UpdatedProOperatorDeployment(r.CR, r.CurrentState.PrometheusOperatorDeployment)

		if err := r.updateObject(deployment); err != nil {
			log.Error(err, "Failed to update prometheus operator deployment in cluster")
			return err
		}
	}
	log.Info("prometheus operator deployment is sync")

	return nil
}
