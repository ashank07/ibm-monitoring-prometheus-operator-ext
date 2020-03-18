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

func (r *Reconsiler) syncMCMCtl() error {

	if r.CurrentState.MCMCtrlDeployment == nil {
		deployment, err := model.NewMCMCtlDeployment(r.CR)
		if err != nil {
			log.Error(err, "Failed to create deployment object for MCM controller")
			return err
		}
		if err = r.createObject(deployment); err != nil {
			log.Error(err, "Failed to create deployment for MCM controller in cluster")
			return err
		}
	} else {
		deployment, err := model.UpdatedMCMCtlDeployment(r.CR, r.CurrentState.MCMCtrlDeployment)
		if err != nil {
			log.Error(err, "Failed to create deployment object for MCM controller")
			return err

		}
		if err = r.updateObject(deployment); err != nil {
			log.Error(err, "Failed to update mcm controller deployment in cluster")
			return err
		}
	}
	log.Info("mcm controller deployment is sync")

	return nil
}
