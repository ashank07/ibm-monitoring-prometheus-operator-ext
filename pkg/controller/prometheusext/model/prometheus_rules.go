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
	"html/template"

	promv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	promext "github.com/IBM/ibm-monitoring-prometheus-operator-ext/pkg/apis/monitoring/v1alpha1"
)

//defaultPromethuesRules is a dictionary to stor default prometheus rules information
var (
	defaultPromethuesRules map[PrometheusRuleName]*promv1.PrometheusRule
	nodeMemUsageExprTempl  *template.Template
	nodeMemUsageDesTempl   *template.Template
	highCPUExprTempl       *template.Template
	highCPUDesTempl        *template.Template
)

const (
	nodeMemUsageExpr = `((node_memory_MemTotal_bytes - (node_memory_MemFree_bytes + node_memory_Buffers_bytes + node_memory_Cached_bytes))/ node_memory_MemTotal_bytes) * 100 > {{.NodeMemoryUsage}}`
	nodeMemUsageDes  = `{{ "{{ " }} $labels.instance {{ " }}" }}: Memory usage is above the {{.NodeMemoryUsage}}% threshold.  The current value is: {{ "{{ " }} $value {{ " }}" }}.`
	highCPUExpr      = `(100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)) > {{.HighCPUUsage}}`
	highCPUDes       = `{{ "{{ " }} $labels.instance {{ " }}" }}: CPU usage is above the {{ .HighCPUUsage }}% threshold.  The current value is: {{ "{{ " }} $value {{ " }}" }}.`
	// NodeMem is rule name for high node memory usage
	NodeMem = PrometheusRuleName("node-memory-usage")
	//NodeCPU is rule name for high for cpu usage
	NodeCPU = PrometheusRuleName("high-cpu-usage")
	//PodTerm is rule name for pods terminateion
	PodTerm = PrometheusRuleName("pods-terminated")
	//PodRestart is rule name for pods restarted
	PodRestart = PrometheusRuleName("pods-restarting")
	//FailedJob is rule name for failed jobs
	FailedJob   = PrometheusRuleName("failed-jobs")
	summary     = "summary"
	description = "description"
)

type nodeMemRulePara struct {
	NodeMemoryUsage int
}
type nodeCPURulePara struct {
	HighCPUUsage int
}

//PrometheusRuleName defines prometheus rule names
type PrometheusRuleName string

//DefaultPrometheusRules return dictionary for default prometheus rules
func DefaultPrometheusRules(cr *promext.PrometheusExt) (map[PrometheusRuleName]*promv1.PrometheusRule, error) {
	var tplBuffer bytes.Buffer
	memPara := nodeMemRulePara{
		NodeMemoryUsage: cr.Spec.PrometheusConfig.NodeMemoryThreshold,
	}
	cpuPara := nodeCPURulePara{
		HighCPUUsage: cr.Spec.PrometheusConfig.NodeCPUThreshold,
	}
	//node memory usage
	if err := nodeMemUsageExprTempl.Execute(&tplBuffer, memPara); err != nil {
		return nil, err
	}
	defaultPromethuesRules[NodeMem].Spec.Groups[0].Rules[0].Expr.StrVal = tplBuffer.String()
	tplBuffer.Reset()

	if err := nodeMemUsageDesTempl.Execute(&tplBuffer, memPara); err != nil {
		return nil, err
	}
	defaultPromethuesRules[NodeMem].Spec.Groups[0].Rules[0].Annotations[description] = tplBuffer.String()
	tplBuffer.Reset()
	//node cpu usage
	if err := highCPUExprTempl.Execute(&tplBuffer, cpuPara); err != nil {
		return nil, err
	}
	defaultPromethuesRules[NodeCPU].Spec.Groups[0].Rules[0].Expr.StrVal = tplBuffer.String()
	tplBuffer.Reset()

	if err := highCPUDesTempl.Execute(&tplBuffer, cpuPara); err != nil {
		return nil, err
	}
	defaultPromethuesRules[NodeCPU].Spec.Groups[0].Rules[0].Annotations[description] = tplBuffer.String()
	tplBuffer.Reset()

	return defaultPromethuesRules, nil

}

//create immutable rules
func init() {
	nodeMemUsageExprTempl = template.Must(template.New("NodeMemUsageExpr").Parse(nodeMemUsageExpr))
	nodeMemUsageDesTempl = template.Must(template.New("NodeMemUsageDes").Parse(nodeMemUsageDes))
	highCPUExprTempl = template.Must(template.New("HighCPUExpr").Parse(highCPUExpr))
	highCPUDesTempl = template.Must(template.New("HighCPUDes").Parse(highCPUDes))

	defaultPromethuesRules = make(map[PrometheusRuleName]*promv1.PrometheusRule)
	defaultPromethuesRules[NodeMem] = &promv1.PrometheusRule{
		Spec: promv1.PrometheusRuleSpec{
			Groups: []promv1.RuleGroup{
				{
					Name: "NodeMemoryUsage",
					Rules: []promv1.Rule{
						{
							Alert: "NodeMemoryUsage",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "",
							},
							For: "5m",
							Annotations: map[string]string{
								description: "",
								summary:     `{{ "{{ " }}$labels.instance{{ " }}" }}: High memory usage detected`,
							},
						},
					},
				},
			},
		},
	}
	defaultPromethuesRules[NodeCPU] = &promv1.PrometheusRule{
		Spec: promv1.PrometheusRuleSpec{
			Groups: []promv1.RuleGroup{
				{
					Name: "HighCPUUsage",
					Rules: []promv1.Rule{
						{
							Alert: "HighCPUUsage",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "",
							},
							For: "5m",
							Annotations: map[string]string{
								description: "",
								summary:     "High CPU Usage",
							},
						},
					},
				},
			},
		},
	}
	defaultPromethuesRules[PodTerm] = &promv1.PrometheusRule{
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
								description: `Pod {{ "{{ " }} $labels.pod {{ " }}" }} in namespace {{ "{{ " }} $labels.namespace {{ " }}" }} has a termination status other than completed.`,
								summary:     `Pod was terminated`,
							},
						},
					},
				},
			},
		},
	}

	defaultPromethuesRules[PodRestart] = &promv1.PrometheusRule{
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
								description: `Pod {{ "{{ " }} $labels.pod {{ " }}" }} in namespace {{ "{{ " }} $labels.namespace {{ " }}" }} is restarting a lot`,
								summary:     `Pod restarting a lot`,
							},
						},
					},
				},
			},
		},
	}
	defaultPromethuesRules[FailedJob] = &promv1.PrometheusRule{
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
								description: `Job {{ "{{ " }} $labels.exported_job {{ " }}" }} in namespace {{ "{{ " }} $labels.namespace {{ " }}" }} failed for some reason.`,
								summary:     "Failed job",
							},
						},
					},
				},
			},
		},
	}
}
