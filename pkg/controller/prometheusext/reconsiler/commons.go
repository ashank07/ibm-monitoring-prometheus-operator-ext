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
	"fmt"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/controller/prometheusext/model"
)

func (r *Reconsiler) syncSecrets() error {
	if err := r.syncSecret(r.CurrentState.MonitoringSecret, r.CR.Spec.Certs.MonitoringSecret, model.MonitoringDNSNames(r.CR)); err != nil {
		return err
	}
	log.Info("monitoring certificate is sync")
	if err := r.syncSecret(r.CurrentState.MonitoringClientSecret, r.CR.Spec.Certs.MonitoringClientSecret, []string{}); err != nil {
		return err
	}
	log.Info("monitoring client certificate is sync")
	return nil
}
func (r *Reconsiler) syncSecret(currentSecret *v1.Secret, secretName string, dnsNames []string) error {
	cert := model.NewCertitication(secretName, r.CR, dnsNames)
	if currentSecret != nil {
		if r.CR.Spec.Certs.AutoClean {
			key := client.ObjectKey{Name: cert.Name, Namespace: cert.Namespace}
			if err := r.Client.Get(r.Context, key, cert); err != nil {
				if kerrors.IsNotFound(err) {
					//secret exists but no certicate
					//delete the secret and create new one
					log.Info("deleting tls secret" + secretName + "which is old and out of control")
					if err = r.Client.Delete(r.Context, currentSecret); err != nil {
						log.Error(err, "failed to delete old tls secret: "+secretName)
						return err
					}

				} else {
					//failed to get certificate because other errors
					log.Error(err, "failed to get certification object: "+secretName)
					return err

				}
			} else {
				return nil
			}

		} else {
			// when it is not autoclean keep secret no matter who created it
			log.Info("exporter cert secret exists")
			return nil
		}
	}

	if err := r.createObject(cert); err != nil {
		if kerrors.IsAlreadyExists(err) {
			log.Info("certificate object already exists.")
			return model.NewRequeueError("syncCertSecret", "wait for cert secret to be created after creating certificate object")
		}
		log.Error(err, "failed to create certificate")
		return err
	}
	// We can not verify if secret is created or not for now so return to next loop
	return model.NewRequeueError("syncCertSecret", "wait for cert secret to be created after creating certification object")

}
func (r *Reconsiler) syncRouterCms() error {
	if err := r.syncRouterEntryCm(); err != nil {
		return err
	}
	log.Info("router entrypoint configmap is sync")
	if err := r.syncProRouterNgCm(); err != nil {
		return err
	}
	log.Info("prometheus router's nginx configmap is sync")
	if err := r.syncProLuaUtilsCm(); err != nil {
		return err
	}
	log.Info("prometheus lua utils configmap is sync")
	if err := r.syncProLuaCm(); err != nil {
		return err
	}
	log.Info("prometheus lua script configmap is sync")
	if err := r.syncAlertRouterNgCm(); err != nil {
		return err
	}
	log.Info("alertmanager router's nginx confgimap is sync")
	return nil
}

func (r *Reconsiler) syncProRouterNgCm() error {
	if r.CurrentState.PromeNgCm == nil {
		cm, err := model.NewProRouterNgCm(r.CR)
		if err != nil {
			return err
		}
		if err = r.createObject(cm); err != nil {
			log.Error(err, "failed to create prometheus router nginx configmap")
			return err
		}
	} else {
		cm, err := model.UpdatedProRouterNgCm(r.CR, r.CurrentState.PromeNgCm)
		if err != nil {
			return err
		}
		if err = r.updateObject(cm); err != nil {
			return err
		}
	}
	return nil
}
func (r *Reconsiler) syncAlertRouterNgCm() error {
	if r.CurrentState.AlertNgCm == nil {
		cm := model.NewAlertmanagerRouterNgCm(r.CR)
		if err := r.createObject(cm); err != nil {
			log.Error(err, "failed to create configmap for alertmanager router nginx config in cluster")
			return err
		}
	} else {
		cm := model.UpdatedAlertRouterNgcm(r.CR, r.CurrentState.AlertNgCm)
		if err := r.updateObject(cm); err != nil {
			log.Error(err, "failed to update configmap for alertmanager router nginx config in cluster")
			return err
		}

	}
	return nil
}
func (r *Reconsiler) syncProLuaCm() error {
	if r.CurrentState.ProLuaCm == nil {
		cm, err := model.NewProLuaCm(r.CR)

		if err != nil {
			log.Error(err, "failed to create configmpa object for prometheus lua script")
			return err
		}
		if err = r.createObject(cm); err != nil {
			log.Error(err, "failed to create configmap in kubernetes for prometheus lua script")
			return err
		}

	} else {
		cm, err := model.UpdatedProLuaCm(r.CR, r.CurrentState.ProLuaCm)
		if err != nil {
			log.Error(err, "failed to update onfigmpa object for prometheus lua script")
			return err
		}
		if err = r.updateObject(cm); err != nil {
			log.Error(err, "failed to update configmap in kubernetes for prometheus lua script")
			return err
		}

	}
	return nil
}
func (r *Reconsiler) syncProLuaUtilsCm() error {
	if r.CurrentState.ProLuaUtilsCm == nil {
		cm, err := model.NewProLuaUtilsCm(r.CR)
		if err != nil {
			log.Error(err, "failed to create configmap for prometheus lua utils")
			return err
		}
		if err = r.createObject(cm); err != nil {
			log.Error(err, "failed to create prometheus lua script configmap in kubernets")
			return err
		}
	} else {
		cm, err := model.UpdatedProLuaUtilsCm(r.CR, r.CurrentState.ProLuaUtilsCm)
		if err != nil {
			log.Error(err, "failed to create updated configmap for prometheus lua utils script")
			return err
		}
		if err = r.updateObject(cm); err != nil {
			log.Error(err, "failed to update configmap in kubernetes for prometheus lua utils script")
			return err
		}
	}

	return nil
}
func (r *Reconsiler) syncRouterEntryCm() error {

	if r.CurrentState.RouterEntryCm == nil {
		cm, err := model.NewRouterEntryCm(r.CR)
		if err != nil {
			log.Error(err, "failed to create configmap for router entrypoint")
			return err
		}

		if err = r.createObject(cm); err != nil {
			log.Error(err, "failed to create configmap for router entrypoint in kubernestes")
			return err
		}

	} else {
		cm, err := model.UpdatedRouterEntryCm(r.CR, r.CurrentState.RouterEntryCm)
		if err != nil {
			return err
		}
		if err = r.updateObject(cm); err != nil {
			log.Error(err, "failed to update configmap for router entrypoint in kubernestes")
			return err
		}

	}
	return nil
}
func (r *Reconsiler) syncStorageClass() error {
	scName, err := r.getStorageClass()
	if err != nil {
		return err
	}
	if r.CR.Annotations == nil {
		r.CR.Annotations = make(map[string]string)

	}
	r.CR.Annotations[model.StorageClassAnn] = scName
	if err := r.Client.Update(r.Context, r.CR); err != nil {
		log.Error(err, "failed to update storage calss annotation")
		return err
	}
	return nil
}
func (r *Reconsiler) syncClusterHostInfo() error {
	host, port, err := r.getClusterInfo()
	if err != nil {
		log.Error(err, "failed to get cluster host and port")
		return err
	}
	if r.CR.Annotations == nil {
		r.CR.Annotations = make(map[string]string)
	}
	r.CR.Annotations[model.ClusterHostAnn] = host
	r.CR.Annotations[model.ClusterPortAnn] = port
	if err := r.Client.Update(r.Context, r.CR); err != nil {
		log.Error(err, "failed to update cluster host info annotations")
		return err
	}
	return nil
}

func (r *Reconsiler) getClusterInfo() (string, string, error) {
	var host string
	var port string
	//check from cr spec
	if r.CR.Spec.ClusterAddress != "" {
		host = r.CR.Spec.ClusterAddress
	}
	if r.CR.Spec.ClusterPort != 0 {
		port = fmt.Sprintf("%d", r.CR.Spec.ClusterPort)
	} else {
		port = model.ExternalPort
	}
	if host != "" {
		return host, port, nil
	}

	//check from cr annotation
	if host == "" {
		ann, ok := r.CR.Annotations[model.ClusterHostAnn]
		if ok && ann != "" {
			host = ann
		}
	}
	if host != "" {
		return host, port, nil
	}

	//check from ingress configmap
	key := client.ObjectKey{Namespace: r.CR.ObjectMeta.Namespace, Name: model.ManagedIngressCm}
	cm := v1.ConfigMap{}
	if err := r.Client.Get(r.Context, key, &cm); err != nil {
		return host, port, err
	}
	host = cm.Data["MANAGEMENT_INGRESS_ROUTE_HOST"]

	return host, port, nil
}

func (r *Reconsiler) getStorageClass() (string, error) {
	if r.CR.Spec.StorageClassName != "" {
		return r.CR.Spec.StorageClassName, nil

	}

	ann, ok := r.CR.Annotations[model.StorageClassAnn]
	if ok && ann != "" {
		return ann, nil
	}

	scList := &storagev1.StorageClassList{}
	err := r.Client.List(r.Context, scList)
	if err != nil {
		log.Error(err, "failed to get storage class list")
		return "", err
	}
	if len(scList.Items) == 0 {
		err := fmt.Errorf("could not find storage class in the cluster")
		log.Error(err, "")
		return "", err
	}

	var defaultSC []string
	var nonDefaultSC []string

	for _, sc := range scList.Items {
		if sc.Provisioner == "kubernetes.io/no-provisioner" {
			continue
		}
		if sc.ObjectMeta.GetAnnotations()["storageclass.kubernetes.io/is-default-class"] == "true" {
			defaultSC = append(defaultSC, sc.GetName())
			continue
		}
		nonDefaultSC = append(nonDefaultSC, sc.GetName())
	}

	if len(defaultSC) != 0 {
		return defaultSC[0], nil

	}

	if len(nonDefaultSC) != 0 {
		return nonDefaultSC[0], nil
	}

	err = fmt.Errorf("could not find dynamic provisioner storage class in the cluster")
	log.Error(err, "")

	return "", err
}
