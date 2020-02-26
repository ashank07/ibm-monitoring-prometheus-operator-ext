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
	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//DefaultPromethuesRules is a dictionary to stor defaulat prometheus rules information
var DefaultPromethuesRules map[string]*promv1.PrometheusRule

func init() {
	DefaultPromethuesRules = make(map[string]*promv1.PrometheusRule)
	DefaultPromethuesRules["pods-terminated"] = &promv1.PrometheusRule{
		Spec: promv1.PrometheusRuleSpec{
			Groups: []promv1.RuleGroup{
				{
					Name: "podsTerminated",
					Rules: []promv1.Rule{
						{
							Alert: "podsTerminated",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum_over_time(kube_pod_container_status_terminated_reason{reason!="Completed"}[1h]) > 0`,
							},
							Annotations: map[string]string{
								"description": `Pod {{ "{{ " }} $labels.pod {{ " }}" }} in namespace {{ "{{ " }} $labels.namespace {{ " }}" }} has a termination status other than completed.`,
								"summary":     `Pod was terminated`,
							},
						},
					},
				},
			},
		},
	}

	DefaultPromethuesRules["pods-restarting"] = &promv1.PrometheusRule{
		Spec: promv1.PrometheusRuleSpec{
			Groups: []promv1.RuleGroup{
				{
					Name: "podsRestarting",
					Rules: []promv1.Rule{
						{
							Alert: "podsRestarting",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `increase(kube_pod_container_status_restarts_total[1h]) > 5`,
							},
							Annotations: map[string]string{
								"description": `Pod {{ "{{ " }} $labels.pod {{ " }}" }} in namespace {{ "{{ " }} $labels.namespace {{ " }}" }} is restarting a lot`,
								"summary":     `Pod restarting a lot`,
							},
						},
					},
				},
			},
		},
	}
	/*

		DefaultPromethuesRules["node-memory-usage"] = &promv1.PrometheusRule{
			Spec: promv1.PrometheusRuleSpec{
				Groups: []promv1.RuleGroup{
					{
						Name: "NodeMemoryUsage",
						Rules: []promv1.Rule{
							{
								Alert: "NodeMemoryUsage",
								Expr: intstr.IntOrString{
									Type:   intstr.String,
									StrVal: `((node_memory_MemTotal_bytes - (node_memory_MemFree_bytes + node_memory_Buffers_bytes + node_memory_Cached_bytes))/ node_memory_MemTotal_bytes) * 100 > {{ .Values.prometheus.alerts.nodeMemoryUsage.nodeMemoryUsageThreshold }}`,
								},
								For: "5m",
								Annotations: map[string]string{
									"description": `{{ "{{ " }} $labels.instance {{ " }}" }}: Memory usage is above the {{ .Values.prometheus.alerts.nodeMemoryUsage.nodeMemoryUsageThreshold }}% threshold.  The current value is: {{ "{{ " }} $value {{ " }}" }}.`,
									"summary":     `{{ "{{ " }}$labels.instance{{ " }}" }}: High memory usage detected`,
								},
							},
						},
					},
				},
			},
		}

		DefaultPromethuesRules["high-cpu-usage"] = &promv1.PrometheusRule{
			Spec: promv1.PrometheusRuleSpec{
				Groups: []promv1.RuleGroup{
					{
						Name: "HighCPUUsage",
						Rules: []promv1.Rule{
							{
								Alert: "HighCPUUsage",
								Expr: intstr.IntOrString{
									Type:   intstr.String,
									StrVal: `(100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)) > {{ .Values.prometheus.alerts.highCPUUsage.highCPUUsageThreshold }}`,
								},
								For: "5m",
								Annotations: map[string]string{
									"description": `{{ "{{ " }} $labels.instance {{ " }}" }}: CPU usage is above the {{ .Values.prometheus.alerts.highCPUUsage.highCPUUsageThreshold }}% threshold.  The current value is: {{ "{{ " }} $value {{ " }}" }}.`,
									"summary":     "High CPU Usage",
								},
							},
						},
					},
				},
			},
		}*/

	DefaultPromethuesRules["failed-jobs"] = &promv1.PrometheusRule{
		Spec: promv1.PrometheusRuleSpec{
			Groups: []promv1.RuleGroup{
				{
					Name: "failedJobs",
					Rules: []promv1.Rule{
						{
							Alert: "failedJobs",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "kube_job_failed != 0",
							},
							Annotations: map[string]string{
								"description": `Job {{ "{{ " }} $labels.exported_job {{ " }}" }} in namespace {{ "{{ " }} $labels.namespace {{ " }}" }} failed for some reason.`,
								"summary":     "Failed job",
							},
						},
					},
				},
			},
		},
	}
}
