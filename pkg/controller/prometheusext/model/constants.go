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

const (
	//AppLabelKey is key of label
	AppLabelKey = "cs/app"
	//AppLabelValue is value of label
	AppLabelValue   = "ibm-monitoring"
	hubPromemetheus = "hub-prometheus"

	//Component is string "component"
	Component = "component"
	//LoopBackIP loopback ip
	LoopBackIP = "127.0.0.1"
	//ExternalPort external port for alertmanager and prometheus
	ExternalPort = "8443"
	//DefaultPVSize is default storage size for alertmanager and prometheus
	DefaultPVSize = "10Gi"
	//Prometheus means object is Prometheus
	Prometheus = ObjectType("prometheus")
	//Alertmanager means object is Alertmanager
	Alertmanager = ObjectType("alertmanager")
	//Grafana means object is grafana. It is only for certification dnsNames here
	Grafana = ObjectType("grafana")

	//HealthCheckKey lable key for metering check
	HealthCheckKey = "app.kubernetes.io/instance"
	//HealthCheckLabelValue label value for metering check
	HealthCheckLabelValue = "common-monitoring"
	//HealthCheckAnnKey annotation key for metering check
	HealthCheckAnnKey = "clusterhealth.ibm.com/dependencies"
	//HealthCheckAnnValue annotation value for metering check
	HealthCheckAnnValue = "cert-manager, icp-management-ingress, auth-idp"

	defaultHelmPort      = int32(3000)
	defaultClusterDomain = "cluster.local"
	defaultClusterName   = "mycluster"

	alertConfigStr = `  
  global:
  receivers:
    - name: default-receiver
  route:
    group_wait: 10s
    group_interval: 5m
    receiver: default-receiver
    repeat_interval: 3h`
	alertRouterConfig = `
	error_log stderr notice;

    events {
        worker_connections 1024;
    }


    http {
        access_log off;

        include mime.types;
        default_type application/octet-stream;
        sendfile on;
        keepalive_timeout 65;
        server_tokens off;
        more_set_headers "Server: ";

        # Without this, cosocket-based code in worker
        # initialization cannot resolve localhost.

        #upstream alertmanager {
        #    server 127.0.0.1:9093;
        #}

        proxy_cache_path /tmp/nginx-mesos-cache levels=1:2 keys_zone=mesos:1m inactive=10m;

        server {
            listen 8443 ssl default_server;
            ssl_certificate server.crt;
            ssl_certificate_key server.key;
            ssl_client_certificate /opt/ibm/router/caCerts/ca.crt;
            ssl_verify_client on;
            ssl_protocols TLSv1.2;
            # Ref: https://github.com/cloudflare/sslconfig/blob/master/conf
            # Modulo ChaCha20 cipher.
            ssl_ciphers EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:!EECDH+3DES:!RSA+3DES:!MD5;
            ssl_prefer_server_ciphers on;

            server_name dcos.*;
            root /opt/ibm/router/nginx/html;

            location / {
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header Host $http_host;
              header_filter_by_lua_block {
                  ngx.header["Cache-control"] = "no-cache, no-store, must-revalidate"
                  ngx.header["Pragma"] = "no-cache"
                  ngx.header["Access-Control-Allow-Credentials"] = "false"
              }
              proxy_pass http://127.0.0.1:9093/alertmanager/;
            }

            location /index.html {
                return 404;
            }

        }
    }
	`
	prometheusRouterConfig = `
	  error_log stderr notice;

    events {
        worker_connections 1024;
    }

    env KUBERNETES_SERVICE_HOST;
    env KUBERNETES_SERVICE_PORT_HTTPS;

    http {
        access_log off;

        include mime.types;
        default_type application/octet-stream;
        sendfile on;
        keepalive_timeout 65;
        server_tokens off;
        more_set_headers "Server: ";

        # Without this, cosocket-based code in worker
        # initialization cannot resolve localhost.

        #upstream prometheus {
        #    server 127.0.0.1:9090/prometheus/;
        #}

        proxy_cache_path /tmp/nginx-mesos-cache levels=1:2 keys_zone=mesos:1m inactive=10m;

    {{- if .Managed -}}
        lua_package_path '$prefix/conf/?.lua;;';
        lua_shared_dict mesos_state_cache 100m;
        lua_shared_dict shmlocks 1m;

        init_by_lua '
            prom = require "prom"
            util = require "monitoring-util"
        ';
      {{if .Openshift}}
        resolver {OPENSHIFT_RESOLVER};
      {{- else -}}
        resolver kube-dns;
      {{- end -}}
    {{- end }}

        server {
            listen 8080;

            server_name dcos.*;
            root /opt/ibm/router/nginx/html;

            location /-/ready {
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header Host $http_host;

              proxy_pass http://127.0.0.1:9090/prometheus/-/ready;
            }

            location /-/healthy {
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header Host $http_host;

              proxy_pass http://127.0.0.1:9090/prometheus/-/healthy;
            }

            location / {
              return 404;
            }
        }

        server {
            listen 8443 ssl default_server;
            ssl_certificate server.crt;
            ssl_certificate_key server.key;
            ssl_client_certificate /opt/ibm/router/caCerts/ca.crt;
            ssl_verify_client on;
            ssl_protocols TLSv1.2;
            # Ref: https://github.com/cloudflare/sslconfig/blob/master/conf
            # Modulo ChaCha20 cipher.
            ssl_ciphers EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:!EECDH+3DES:!RSA+3DES:!MD5;
            ssl_prefer_server_ciphers on;

            server_name dcos.*;
            root /opt/ibm/router/nginx/html;

            location /federate {
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header Host $http_host;

              proxy_pass http://127.0.0.1:9090/prometheus/federate;
            }

            location /api/v1/series {
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header Host $http_host;
              if ($arg_match[] = "helm_release_info") {
                 content_by_lua 'prom.write_release_response()';
              }
              header_filter_by_lua_block {
                  ngx.header["Cache-control"] = "no-cache, no-store, must-revalidate"
                  ngx.header["Pragma"] = "no-cache"
                  ngx.header["Access-Control-Allow-Credentials"] = "false"
              }
            {{- if .Managed -}}
              rewrite_by_lua 'prom.rewrite_query()';
            {{- end }}
              proxy_pass http://127.0.0.1:9090/prometheus/api/v1/series;

            }

            location /status {
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header Host $http_host;

            {{- if .Managed -}}
              header_filter_by_lua_block {
                  ngx.header["Cache-control"] = "no-cache, no-store, must-revalidate"
                  ngx.header["Pragma"] = "no-cache"
                  ngx.header["Access-Control-Allow-Credentials"] = "false"
                  util.remove_content_len_header()
              }
              body_filter_by_lua 'prom.filter_alertmanager_url()';
            {{- end }}

              proxy_pass http://127.0.0.1:9090/prometheus/status;
            }

            location / {
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header Host $http_host;
              header_filter_by_lua_block {
                  ngx.header["Cache-control"] = "no-cache, no-store, must-revalidate"
                  ngx.header["Pragma"] = "no-cache"
                  ngx.header["Access-Control-Allow-Credentials"] = "false"
              }
            {{- if .Standalone }}
              if ($arg_query = "cluster_datasource_info") {
                 content_by_lua 'prom.write_cluster_datasource_response()';
              }
            {{- end}}
            {{- if .Managed -}}
              rewrite_by_lua 'prom.rewrite_query()';
            {{- end }}

              proxy_pass http://127.0.0.1:9090/prometheus/;
            }

            location /index.html {
                return 404;
            }
        }
    }
	`

	//TODO: what standalone here mean? hardcode to true for now
	promeScrapeTargets = `
  - job_name: prometheus
    static_configs:
      - targets:
        - 127.0.0.1:9090
      {{- if not .Standalone }}
        labels:
          metrics_type: system
      {{- end }}
    metrics_path: /prometheus/metrics

  # A scrape configuration for running Prometheus on a Kubernetes cluster.
  # This uses separate scrape configs for cluster components (i.e. API server, node)
  # and services to allow each to use different authentication configs.
  #
  # Kubernetes labels will be added as Prometheus labels on metrics via the
  # "labelmap" relabeling action.

  # Scrape config for API servers.
  #
  # Kubernetes exposes API servers as endpoints to the default/kubernetes
  # service so this uses "endpoints" role and uses relabelling to only keep
  # the endpoints associated with the default/kubernetes service using the
  # default named port "https". This works for single API server deployments as
  # well as HA API server deployments.
  - job_name: 'kubernetes-apiservers'

    kubernetes_sd_configs:
      - role: endpoints

    # Default to scraping over https. If required, just disable this or change to
    # "http".
    scheme: https

    # This TLS & bearer token file config is used to connect to the actual scrape
    # endpoints for cluster components. This is separate to discovery auth
    # configuration because discovery & scraping are two separate concerns in
    # Prometheus. The discovery auth config is automatic if Prometheus runs inside
    # the cluster. Otherwise, more config options have to be provided within the
    # <kubernetes_sd_config>.
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      # If your node certificates are self-signed or use a different CA to the
      # master CA, then disable certificate verification below. Note that
      # certificate verification is an integral part of a secure infrastructure
      # so this should only be disabled in a controlled environment. You can
      # disable certificate verification by uncommenting the line below.
      #
      insecure_skip_verify: true
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

    # Keep only the default/kubernetes service endpoints for the https port. This
    # will add targets for each API server which Kubernetes adds an endpoint to
    # the default/kubernetes service.
    relabel_configs:
      - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: default;kubernetes;https

  {{- if not .Standalone }}
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: (.*)
        replacement: system
        target_label: metrics_type
  {{- end }}

  - job_name: 'kubernetes-nodes'

    # Default to scraping over https. If required, just disable this or change to
    # "http".
    scheme: https

    # This TLS & bearer token file config is used to connect to the actual scrape
    # endpoints for cluster components. This is separate to discovery auth
    # configuration because discovery & scraping are two separate concerns in
    # Prometheus. The discovery auth config is automatic if Prometheus runs inside
    # the cluster. Otherwise, more config options have to be provided within the
    # <kubernetes_sd_config>.
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      # If your node certificates are self-signed or use a different CA to the
      # master CA, then disable certificate verification below. Note that
      # certificate verification is an integral part of a secure infrastructure
      # so this should only be disabled in a controlled environment. You can
      # disable certificate verification by uncommenting the line below.
      #
      insecure_skip_verify: true
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

    kubernetes_sd_configs:
      - role: node

    relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics

  {{- if not .Standalone }}
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: (.*)
        replacement: system
        target_label: metrics_type
  {{- end }}

  # Scrape config for Kubelet cAdvisor.
  #
  # This is required for Kubernetes 1.7.3 and later, where cAdvisor metrics
  # (those whose names begin with 'container_') have been removed from the
  # Kubelet metrics endpoint.  This job scrapes the cAdvisor endpoint to
  # retrieve those metrics.
  #
  # In Kubernetes 1.7.0-1.7.2, these metrics are only exposed on the cAdvisor
  # HTTP endpoint; use "replacement: /api/v1/nodes/${1}:4194/proxy/metrics"
  # in that case (and ensure cAdvisor's HTTP server hasn't been disabled with
  # the --cadvisor-port=0 Kubelet flag).
  #
  # This job is not necessary and should be removed in Kubernetes 1.6 and
  # earlier versions, or it will cause the metrics to be scraped twice.
  - job_name: 'kubernetes-cadvisor'

    # Default to scraping over https. If required, just disable this or change to
    # "http".
    scheme: https

    # This TLS & bearer token file config is used to connect to the actual scrape
    # endpoints for cluster components. This is separate to discovery auth
    # configuration because discovery & scraping are two separate concerns in
    # Prometheus. The discovery auth config is automatic if Prometheus runs inside
    # the cluster. Otherwise, more config options have to be provided within the
    # <kubernetes_sd_config>.
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

    kubernetes_sd_configs:
      - role: node

    relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor

    metric_relabel_configs:
      - source_labels: ['namespace']
        regex: (.+)
        target_label: kubernetes_namespace
    {{- if not .Standalone }}
      - source_labels: ['kubernetes_namespace']
        regex: (.*)
        target_label: hub_kubernetes_namespace
      - source_labels: ['kubernetes_namespace']
        regex: ""
        replacement: system
        target_label: metrics_type
    {{- end }}

  # Scrape config for service endpoints.
  #
  # The relabeling allows the actual service scrape endpoint to be configured
  # via the following annotations:
  #
  # * "prometheus.io/scrape": Only scrape services that have a value of "true"
  # * "prometheus.io/scheme": If the metrics endpoint is secured then you will need
  # to set this to "https" & most likely set the "tls_config" of the scrape config.
  # * "prometheus.io/path": If the metrics path is not "/metrics" override this.
  # * "prometheus.io/por": If the metrics are exposed on a different port to the
  # service then set this appropriately.
  - job_name: 'kubernetes-service-endpoints'

    kubernetes_sd_configs:
      - role: endpoints

    relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: drop
        regex: https
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      - source_labels: [__meta_kubernetes_endpoint_port_name, __meta_kubernetes_service_annotation_filter_by_port_name]
        action: drop
        regex: ^([^m].+|m[^e].+|me[^t].+|met[^r].+|metr[^i].+|metri[^c].+|metric[^s]).*;true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name

    metric_relabel_configs:
      - source_labels: ['namespace']
        regex: (.+)
        target_label: kubernetes_namespace
    {{- if not .Standalone }}
      - source_labels: ['kubernetes_namespace']
        regex: (.*)
        target_label: hub_kubernetes_namespace
      - source_labels: ['kubernetes_namespace']
        regex: ""
        replacement: system
        target_label: metrics_type
    {{- end }}

  # Scrape config for service endpoints with tls enabled.
  #
  # The relabeling allows the actual service scrape endpoint to be configured
  # via the following annotations:
  #
  # * "prometheus.io/scrape": Only scrape services that have a value of "true"
  # * "prometheus.io/scheme": If the metrics endpoint is secured then you will need
  # to set this to "https" & most likely set the "tls_config" of the scrape config.
  # * "prometheus.io/path": If the metrics path is not "/metrics" override this.
  # * "prometheus.io/port": If the metrics are exposed on a different port to the
  # service then set this appropriately.
  - job_name: 'kubernetes-service-endpoints-with-tls'

    kubernetes_sd_configs:
      - role: endpoints

    relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_skip_verify]
        action: drop
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: keep
        regex: https
      - source_labels: [__meta_kubernetes_endpoint_port_name, __meta_kubernetes_service_annotation_filter_by_port_name]
        action: drop
        regex: ^([^m].+|m[^e].+|me[^t].+|met[^r].+|metr[^i].+|metri[^c].+|metric[^s]).*;true
      - source_labels: [__meta_kubernetes_namespace]
        action: drop
        regex: openshift-(.+)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_namespace]
        action: replace
        target_label: __address__
        regex: (\d+).(\d+).(\d+).(\d+):(\d+);(.+)
        replacement: $1-$2-$3-$4.$6.pod.{{ .ClusterDomain }}:$5
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name

    metric_relabel_configs:
      - source_labels: ['namespace']
        regex: (.+)
        target_label: kubernetes_namespace
    {{- if not .Standalone }}
      - source_labels: ['kubernetes_namespace']
        regex: (.*)
        target_label: hub_kubernetes_namespace
      - source_labels: ['kubernetes_namespace']
        regex: ""
        replacement: system
        target_label: metrics_type
    {{- end }}

    scheme: https

    tls_config:
      ca_file: /etc/prometheus/secrets/{{ .CASecretName }}/ca.crt
      cert_file: /etc/prometheus/secrets/{{ .ClientSecretName }}/tls.crt
      key_file: /etc/prometheus/secrets/{{ .ClientSecretName }}/tls.key
      insecure_skip_verify: true

{{- if .NodeExporter }}
  - job_name: 'node-exporter-endpoints-with-tls'

    kubernetes_sd_configs:
      - role: endpoints

    relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_skip_verify]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: keep
        regex: https
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name

  {{- if not .Standalone }}
    metric_relabel_configs:
      - source_labels: ['kubernetes_namespace']
        regex: (.*)
        target_label: hub_kubernetes_namespace
      - source_labels: ['kubernetes_namespace']
        regex: ""
        replacement: system
        target_label: metrics_type
  {{- end }}

    scheme: https

    tls_config:
      ca_file: /etc/prometheus/secrets/{{ .CASecretName }}/ca.crt
      cert_file: /etc/prometheus/secrets/{{ .ClientSecretName }}/tls.crt
      key_file: /etc/prometheus/secrets/{{ .ClientSecretName }}/tls.key
      insecure_skip_verify: true
{{- end }}

  # Example scrape config for probing services via the Blackbox Exporter.
  #
  # The relabeling allows the actual service scrape endpoint to be configured
  # via the following annotations:
  #
  # * "prometheus.io/prob": Only probe services that have a value of "true"
  - job_name: 'kubernetes-services'

    metrics_path: /probe
    params:
      module: [http_2xx]

    kubernetes_sd_configs:
      - role: service

    relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_probe]
        action: keep
        regex: true
      - source_labels: [__address__]
        target_label: __param_target
      - target_label: __address__
        replacement: blackbox
      - source_labels: [__param_target]
        target_label: instance
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        target_label: kubernetes_name

  {{- if not .Standalone }}
    metric_relabel_configs:
      - source_labels: ['kubernetes_namespace']
        regex: (.*)
        target_label: hub_kubernetes_namespace
      - source_labels: ['kubernetes_namespace']
        regex: ""
        replacement: system
        target_label: metrics_type
  {{- end }}

  # Example scrape config for pods
  #
  # The relabeling allows the actual pod scrape endpoint to be configured via the
  # following annotations:
  #
  # * "prometheus.io/scrape": Only scrape pods that have a value of "true"
  # * "prometheus.io/path": If the metrics path is not "/metrics" override this.
  # * "prometheus.io/port": Scrape the pod on the indicated port instead of the default of "9102".
  - job_name: 'kubernetes-pods'

    kubernetes_sd_configs:
      - role: pod

    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: ${1}:${2}
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: kubernetes_pod_name
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_target]
        action: replace
        target_label: __param_target
        regex: (.*)
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_module]
        action: replace
        target_label: __param_module
        regex: (.*)

  {{- if not .Standalone }}
    metric_relabel_configs:
      - source_labels: ['kubernetes_namespace']
        regex: (.*)
        target_label: hub_kubernetes_namespace
      - source_labels: ['kubernetes_namespace']
        regex: ""
        replacement: system
        target_label: metrics_type
  {{- end }}
    `
	routerEntrypoint = `#!/bin/sh
    if [ -e /opt/ibm/router/certs/tls.crt ]; then
      cp -f /opt/ibm/router/certs/tls.crt /opt/ibm/router/nginx/conf/server.crt
      cp -f /opt/ibm/router/certs/tls.key /opt/ibm/router/nginx/conf/server.key
    fi

    cp -f /opt/ibm/router/conf/nginx.conf /opt/ibm/router/nginx/conf/nginx.conf.monitoring

    sed -i "s/{NODE_NAME}/${NODE_NAME}/g" /opt/ibm/router/nginx/conf/nginx.conf.monitoring

  {{- if .Openshift }}
    export OPENSHIFT_RESOLVER=$(cat /etc/resolv.conf |grep nameserver|awk '{split($0, a, " "); print a[2]}')
    sed -i "s/{OPENSHIFT_RESOLVER}/${OPENSHIFT_RESOLVER}/g" /opt/ibm/router/nginx/conf/nginx.conf.monitoring
  {{- end }}

  {{- if .Managed }}
    if [ -d /opt/ibm/router/lua-scripts ]; then
      cp -f /opt/ibm/router/lua-scripts/*.lua /opt/ibm/router/nginx/conf/
    fi
  {{- end }}
    exec nginx -c /opt/ibm/router/nginx/conf/nginx.conf.monitoring -g 'daemon off;'`

	luaScripts = `    local cjson = require "cjson"
  local util = require "monitoring-util"
  local http = require "lib.resty.http"
  local operators = {count_values=true, group_left=true, group_right=true}

  local function query_label_key()
  {{- if .Standalone }}
      return "kubernetes_namespace"
  {{- else }}
      return "hub_kubernetes_namespace"
  {{- end }}
  end

  local function inject_query(namespaces, query)
      local query_string = query_label_key()..'=~"'
      for i, entry in ipairs(namespaces) do
          query_string = query_string .. entry.namespaceId
          query_string = query_string .. "|"
      end
      if query_string == query_label_key()..'=~"' then
          return nil, util.exit_401()
      end
      --- remove the last |
      query_string = string.sub(query_string, 1, -2)
  {{- if .Standalone }}
      query_string = query_string .. '"'
  {{- else }}
      local cluster_query_string = 'cluster_name=~"'
      for i, entry in ipairs(namespaces) do
          local cluster_name = util.get_cluster(entry.namespaceId)
          if cluster_name ~= nil then
              cluster_query_string = cluster_query_string .. cluster_name
              cluster_query_string = cluster_query_string .. "|"
          end
      end
      if cluster_query_string ~= 'cluster_name=~"' then
          query_string = query_string .. '|",' .. cluster_query_string .. '",metrics_type!="system"'
      else
          query_string = query_string .. '", metrics_type!="system"'
      end
  {{- end }}
      ngx.log(ngx.DEBUG, "query_string ", query_string)

      --- assume metric's name format is A_B_C
      --- first step is to retrieve all metric names in query string
      --- remove 4 exceptions: 1. by (A_B_C) 2. {A_B_C=""} 3. [A_B_C] 4. A_B_C() 5. "A_B_C"
      local metrics_str = query:gsub("by %([^%(%)]+%)", "---")
      metrics_str = metrics_str:gsub("by%([^%(%)]+%)", "---")
      metrics_str = metrics_str:gsub("%b{}", "---")
      metrics_str = metrics_str:gsub("%[[^%]]+%]", "---")
      metrics_str = metrics_str:gsub("([_%w]+_[_%w]+)%(", "---")
      metrics_str = metrics_str:gsub('%"([_%w]+_[_%w]+)%"', "---")

      --- To inject query string
      --- if there is already label filter append the query string to existing ones
      --- if no label filter add it
      for metric in string.gmatch(metrics_str, "([_:%w]+_[_:%w]+)") do
          if not operators[metric] then
              if string.find(query, metric .."{") then
                  query = query:gsub(metric .."{([^%}]*)}", metric .. "{%1," .. query_string .. "}")
              else
                  query = query:gsub(metric, metric .. "{" .. query_string .. "}")
              end
              --- handle the original query string with empty curly like A_B_C{}
              if string.find(query, metric .."{,") then
                  query = query:gsub(metric .."{,", metric .."{")
              end
          end
      end
      --- handle the query string w/o metrics but only labels like {label=""}
      if string.match(query, "^{[^{}]*}$") ~= nil then
          if query == "{}" then
              query = "{" .. query_string .. "}"
          else
              query = string.sub(query, 1, -2) .. "," .. query_string .. "}"
          end
      end
      ngx.log(ngx.DEBUG, "updated query " .. query)
      return query
  end

  local function get_releases(token, time)
      local httpc = http.new()
        local res, err = httpc:request_uri("http://helm-api.{{ .HelmNamespace }}.svc.{{ .ClusterDomain }}:{{.HelmPort}}/api/v2/releases", {
            method = "GET",
            headers = {
              ["Content-Type"] = "application/json",
              ["Authorization"] = "Bearer ".. token,
              ["cookie"] = "cfc-access-token-cookie="..token
            }
        })
        if not res then
            ngx.log(ngx.ERR, "Failed to get helm releases",err)
            return nil, util.exit_500()
        end
        if (res.body == "" or res.body == nil) then
            ngx.log(ngx.ERR, "Empty response body")
            return nil, util.exit_500()
        end
        local x = tostring(res.body)
        ngx.log(ngx.DEBUG, "response is ",x)
        local releases_result = cjson.decode(x).data
        local release_list = {}
        for index, release in ipairs(releases_result) do
            local release_attrs = {}
            release_attrs.__name__ = "helm_release_info"
            release_attrs.release_name = release.attributes.name
            release_attrs.chart_name = release.attributes.chartName
            release_attrs.chart_version = release.attributes.chartVersion
            release_attrs.status = release.attributes.status
            release_attrs.namespace = release.attributes.namespace
            table.insert(release_list, release_attrs)
            local release_str = cjson.encode(release_list)
        end
        return release_list, nil
  end

  local function get_release_pods(token, release_name)
      ngx.log(ngx.DEBUG, "Check pod of release ",release_name)
        local no_pods_str = "NONE"
        if release_name == "" then
            return no_pods_str
        end
        local httpc = http.new()
        local res, err = httpc:request_uri("http://helm-api.{{ .HelmNamespace }}.svc.{{ .ClusterDomain }}:{{.HelmPort}}/api/v2/releases/"..release_name, {
            method = "GET",
            headers = {
              ["Content-Type"] = "application/json",
              ["Authorization"] = "Bearer ".. token,
              ["cookie"] = "cfc-access-token-cookie="..token
            }
        })
        if not res then
            ngx.log(ngx.ERR, "Failed to get pods of release ",err)
            return no_pods_str
        end
        if res.status == 404 then
            ngx.log(ngx.ERR, "The release does not exist: ", release_name)
            return no_pods_str
        end
        if (res.body == "" or res.body == nil) then
            ngx.log(ngx.ERR, "Empty response body")
            return no_pods_str
        end
        local x = tostring(res.body)
        ngx.log(ngx.DEBUG, "response is ",x)
        local resources_str = cjson.decode(x).data.attributes.resources
        local s_index = string.find(resources_str, "==> v1/Pod")
        if s_index == nil then
            return no_pods_str
        end
        local e_index = string.find(resources_str, "==>", s_index + 1)
        local pod_str
        if e_index ~= nil then
            pod_str = string.sub(resources_str, s_index, e_index)
        else
            pod_str = string.sub(resources_str, s_index)
        end
        local i=1
        local pods=""
        for pod_line in string.gmatch(pod_str, "([^\n]+)") do
            if string.find(pod_line, " ") ~= nil then
                if i > 2 then
                    pod_name = string.sub(pod_line, 1, string.find(pod_line, " ") - 1)
                    if i ~= 3 then
                        pod_name = "|"..pod_name
                    end
                    pods=pods..pod_name
                end
            end
            i = i + 1
        end
        ngx.log(ngx.DEBUG, "pods string is ",pods)
        return pods
  end

  local function write_release_response()
      local token, err = util.get_auth_token()
      if err ~= nil then
          err()
      else
          local release_list, err = get_releases(token, nil)
          if err ~= nil then
              err()
          else
              local response = {}
              response.status = "success"
              response.data = release_list
              local response_str = cjson.encode(response)
              ngx.log(ngx.DEBUG, "resp is ", response_str)
              ngx.header["Content-type"] = "application/json"
              ngx.say(response_str)
              ngx.exit(200)
          end
      end
  end

  local function write_cluster_datasource_response()
      local token, err = util.get_auth_token()
      if err ~= nil then
          err()
      else
          local servicemonitor_items, err = util.get_servicemonitor()
          if err ~= nil then
              err()
          else
              local response = {}
              response.status = "success"
              response.data = {}
              response.data.resultType = "matrix"
              if not next(servicemonitor_items) then
                  response.data.result = {}
              else
                  local token, err = util.get_auth_token()
                  if err ~= nil then
                      err()
                  end
                  if token ~= nil then
                      local uid, err = util.get_user_id(token)
                      if err ~= nil then
                          err()
                      end
                      local role_id, err = util.get_user_role(token, uid)
                      if err ~= nil then
                          err()
                      end
                      local unauth_clusters = {}
                      if (role_id ~= '"ClusterAdministrator"' ) then
                          local clusters, err = util.get_clusters()
                          if err ~= nil then
                              err()
                          end
                          local namespaces, err = util.get_user_namespaces(token, uid)
                          if err ~= nil then
                              err()
                          end
                          namespaces_map = {}
                          for i, ns in ipairs(namespaces) do
                              namespaces_map[ns.namespaceId] = ns
                          end
                          for i, cluster in ipairs(clusters) do
                              if namespaces_map[cluster.metadata.namespace] == nil then
                                  unauth_clusters[cluster.metadata.name] = true
                              end
                          end
                      end

                      local result = {}
                      for i, sm in ipairs(servicemonitor_items) do
                          local cluster = sm.spec.endpoints[1].relabelings[1].replacement
                          if not unauth_clusters[cluster] then
                              local sm_attrs = {}
                              sm_attrs.cluster = cluster
                              sm_attrs.datasource = sm.metadata.labels.prometheus
                              if sm_attrs.datasource == nil then
                                  sm_attrs.datasource = "prometheus"
                              end
                              local metric = {}
                              local values = {}
                              table.insert(values, {ngx.var.arg_time, "1.0"})
                              metric["metric"] = sm_attrs
                              metric["values"] = values
                              table.insert(result, metric)
                          end
                      end
                      response.data.result = result
                  end
              end

              local response_str = cjson.encode(response)
              ngx.log(ngx.DEBUG, "resp is ", response_str)
              ngx.header["Content-type"] = "application/json"
              ngx.say(response_str)
              ngx.exit(200)
          end
      end
  end

  local function rewrite_query()
      local args = ngx.req.get_uri_args()
      local query_key = nil
      if args["query"] ~= nil then
          query_key = "query"
      else
          if args["match[]"] ~= nil then
              query_key = "match[]"
          end
      end
      if query_key ~= nil then
          local query = args[query_key]
          local token, err = util.get_auth_token()
          if err ~= nil then
              return err
          end
          if token ~= nil then
              local uid, err = util.get_user_id(token)
              if err ~= nil then
                  return err
              end
              local role_id, err = util.get_user_role(token, uid)
              if err ~= nil then
                  return err
              end
              if (role_id ~= '"ClusterAdministrator"' ) then
                  local namespaces, err = util.get_user_namespaces(token, uid)
                  if err ~= nil then
                      return err
                  end
                  local updated_query, err = inject_query(namespaces, query)
                  if err ~= nil then
                      return err
                  end
                  args[query_key] = updated_query
                  ngx.req.set_uri_args(args)
              end

              --- replace release_name="release1" to pod_name=~"pod1|pod2|pod3"
              if (string.find(query, "release_name=") ~= nil) then
                  local start_index,end_index = string.find(query, "release_name=[^},]+")
                  local release_name = string.sub(query, start_index + 14, end_index - 1)
                  local pod_list = get_release_pods(token, release_name)
                  local updated_query = string.gsub(query, "release_name=[^},]+", "pod_name=~\""..pod_list.."\"")
                  ngx.log(ngx.DEBUG, 'updated_query is ', updated_query)
                  args[query_key] = updated_query
                  ngx.req.set_uri_args(args)
              end
          end
      end
  end

  local function filter_alertmanager_url()
      targetstr = '<a href="https://{{ .AlertmanagerSvcName}}:{{.AlertmanagerSvcPort}}">{{ .AlertmanagerSvcName}}:{{.AlertmanagerSvcPort}}</a>'
      replacestr = '{{ .AlertmanagerSvcName}}:{{.AlertmanagerSvcPort}}'
      ngx.arg[1] = ngx.re.sub(ngx.arg[1], targetstr, replacestr)
  end

  -- Expose interface.
  local _M = {}
  _M.rewrite_query = rewrite_query
  _M.write_release_response = write_release_response
  _M.filter_alertmanager_url = filter_alertmanager_url
  _M.write_cluster_datasource_response = write_cluster_datasource_response

  return _M`

	luaUtilsScripts = `
  local cjson = require "cjson"
  local cookiejar = require "resty.cookie"
  local http = require "lib.resty.http"

  local function exit_401()
      ngx.status = ngx.HTTP_UNAUTHORIZED
      ngx.header["Content-Type"] = "text/html; charset=UTF-8"
      ngx.header["WWW-Authenticate"] = "oauthjwt"
      ngx.say('401 Unauthorized')
      return ngx.exit(ngx.HTTP_UNAUTHORIZED)
  end

  local function exit_500()
      ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
      ngx.header["Content-Type"] = "text/html; charset=UTF-8"
      ngx.header["WWW-Authenticate"] = "oauthjwt"
      ngx.say('Internal Error')
      return ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  end

  local function get_auth_token()
      local auth_header = ngx.var.http_Authorization

      local token = nil
      if auth_header ~= nil then
          ngx.log(ngx.DEBUG, "Authorization header found. Attempt to extract token.")
          _, _, token = string.find(auth_header, "Bearer%s+(.+)")
      end

      if (auth_header == nil or token == nil) then
          ngx.log(ngx.DEBUG, "Authorization header not found.")
          -- Presence of Authorization header overrides cookie method entirely.
          -- Read cookie. Note: ngx.var.cookie_* cannot access a cookie with a
          -- dash in its name.
          local cookie, err = cookiejar:new()
          token = cookie:get("cfc-access-token-cookie")
          if token == nil then
              ngx.log(ngx.ERR, "cfc-access-token-cookie not found.")
          else
              ngx.log(
                  ngx.NOTICE, "Use token from cfc-access-token-cookie, " ..
                  "set corresponding Authorization header for upstream."
                  )
          end
      end

      if token == nil then
          ngx.log(ngx.DEBUG, "to check host")
          local host_header = ngx.req.get_headers()["host"]
          --- if request host is "monitoring-prometheus:9090" or "monitoring-grafana:3000" skip the rbac check
          ngx.log(ngx.DEBUG, "host header is ",host_header)
          if host_header == "{{ .PrometheusSvcName }}:{{ .PrometheusSvcPort }}" or host_header == "{{ .GrafanaSvcName }}:{{ .GrafanaSvcPort }}" then
              ngx.log(ngx.NOTICE, "skip rbac check for request from monitoring stack")
          else
              ngx.log(ngx.ERR, "No auth token in request.")
              return nil, exit_401()
          end
      end

      return token
  end

  local function get_user_id(token)
      local user_id = ""
      local httpc = http.new()
      ngx.req.set_header('Authorization', 'Bearer '.. token)
      local res, err = httpc:request_uri("https://{{.IAMProviderSvcName}}.{{.IAMNamespace}}.svc.{{.ClusterDomain}}:{{.IAMProviderSvcPort}}/v1/auth/userInfo", {
          method = "POST",
          body = "access_token=" .. token,
          headers = {
            ["Content-Type"] = "application/x-www-form-urlencoded"
          },
          ssl_verify = false
      })

      if not res then
          ngx.log(ngx.ERR, "Failed to request userinfo due to ",err)
          return nil, exit_401()
      end
      if (res.body == "" or res.body == nil) then
          ngx.log(ngx.ERR, "Empty response body")
          return nil, exit_401()
      end
      local x = tostring(res.body)
      local uid = cjson.decode(x).sub
      ngx.log(ngx.DEBUG, "UID is ",uid)
      return uid
  end

  local function get_user_role(token, uid)
      local httpc = http.new()
      local res, err = httpc:request_uri("https://{{.IAMManagementSvcName}}.{{ .IAMNamespace}}.svc.{{.ClusterDomain}}:{{.IAMManagementSvcPort}}/identity/api/v1/users/" .. uid .. "/getHighestRoleForCRN", {
          method = "GET",
          headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer ".. token
          },
          query = {
              ["crn"] = "crn:v1:icp:private:k8:{{ .ClusterName }}:n/{{ .Namespace }}:::"
          },
          ssl_verify = false
      })
      if not res then
          ngx.log(ngx.ERR, "Failed to request user role due to ",err)
          return nil, exit_401()
      end
      if (res.body == "" or res.body == nil) then
          ngx.log(ngx.ERR, "Empty response body")
          return nil, exit_401()
      end
      local role_id = tostring(res.body)
      ngx.log(ngx.DEBUG, "user role ", role_id)
      return role_id
  end

  local function get_user_namespaces(token, uid)
      local httpc = http.new()
      res, err = httpc:request_uri("https://{{.IAMManagementSvcName}}.{{ .IAMNamespace}}.svc.{{.ClusterDomain}}:{{.IAMManagementSvcPort}}/identity/api/v1/users/" .. uid .. "/getTeamResources", {
          method = "GET",
          headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer ".. token
          },
          query = {
              ["resourceType"] = "namespace"
          },
          ssl_verify = false
      })
      if not res then
          ngx.log(ngx.ERR, "Failed to request user's authorized namespaces due to ",err)
          return nil, exit_401()
      end
      if (res.body == "" or res.body == nil) then
          ngx.log(ngx.ERR, "Empty response body")
          return nil, exit_401()
      end
      local x = tostring(res.body)
      ngx.log(ngx.DEBUG, "namespaces ",x)
      local namespaces = cjson.decode(x)
      return namespaces
  end

  function readFile(file)
      local f = io.open(file, "rb")
      local content = f:read("*all")
      f:close()
      return content
  end

  local function get_cluster(namespace)
      local httpc = http.new()
      res, err = httpc:request_uri("https://" .. os.getenv("KUBERNETES_SERVICE_HOST") .. ":" .. os.getenv("KUBERNETES_SERVICE_PORT_HTTPS") .. "/apis/clusterregistry.k8s.io/v1alpha1/namespaces/" .. namespace .. "/clusters", {
          method = "GET",
          headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer ".. readFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
          },
          ssl_verify = false
      })
      if not res then
          ngx.log(ngx.ERR, "Failed to request namespace's clusters due to ",err)
          return nil
      end
      if (res.body == "" or res.body == nil or res.status ~= ngx.HTTP_OK) then
          ngx.log(ngx.ERR, "Invalid response ", res.status)
          return nil
      end
      local x = tostring(res.body)
      ngx.log(ngx.DEBUG, "clusters ",x)
      local clusters = cjson.decode(x)
      if clusters.items[1] == nil then
          return nil
      else
          return clusters.items[1].metadata.name
      end
  end

  local function remove_content_len_header()
      ngx.header.content_length = nil
  end

  local function get_all_users(token)
      local httpc = http.new()
      res, err = httpc:request_uri("https://{{.IAMManagementSvcName}}.{{ .IAMNamespace}}.svc.{{.ClusterDomain}}:{{.IAMManagementSvcPort}}/identity/api/v1/users", {
          method = "GET",
          headers = {
            ["Accept"] = "application/json",
            ["Authorization"] = "Bearer ".. token
          },
          ssl_verify = false
      })
      if not res then
          ngx.log(ngx.ERR, "Failed to get all users due to ",err)
          return nil, exit_500()
      end
      if (res.body == "" or res.body == nil) then
          ngx.log(ngx.ERR, "Empty response body")
          return nil, exit_500()
      end
      local x = tostring(res.body)
      ngx.log(ngx.DEBUG, "users: ",x)
      return cjson.decode(x)
  end

  local function get_clusters()
      local httpc = http.new()
      res, err = httpc:request_uri("https://" .. os.getenv("KUBERNETES_SERVICE_HOST") .. ":" .. os.getenv("KUBERNETES_SERVICE_PORT_HTTPS") .. "/apis/clusterregistry.k8s.io/v1alpha1/clusters", {
          method = "GET",
          headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer ".. readFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
          },
          ssl_verify = false
      })
      if (err ~= nil or not res) then
          ngx.log(ngx.ERR, "Failed to request clusters due to ",err)
          return nil, util.exit_500()
      end
      if (res.body == "" or res.body == nil or res.status ~= ngx.HTTP_OK) then
          ngx.log(ngx.ERR, "Invalid response ", res.status)
          return nil, util.exit_500()
      end
      local x = tostring(res.body)
      ngx.log(ngx.DEBUG, "clusters ",x)
      local clusters = cjson.decode(x)
      return clusters.items, nil
  end

  local function get_servicemonitor()
      local httpc = http.new()
      local res, err = httpc:request_uri("https://" .. os.getenv("KUBERNETES_SERVICE_HOST") .. ":" .. os.getenv("KUBERNETES_SERVICE_PORT_HTTPS") .. "/apis/monitoring.coreos.com/v1/namespaces/{{ .Namespace }}/servicemonitors", {
          method = "GET",
          headers = {
            ["Content-Type"] = "application/json",
            ["Authorization"] = "Bearer ".. readFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
          },
          query = {
              ["labelSelector"] = "owner=mcm-cluster"
          },
          ssl_verify = false
      })
      if (err ~= nil or not res) then
          ngx.log(ngx.ERR, "Failed to list servicemonitor ",err)
          return nil, util.exit_500()
      end
      if (res.body == "" or res.body == nil) then
          ngx.log(ngx.ERR, "Empty response body")
          return nil, util.exit_500()
      end
      local x = tostring(res.body)
      ngx.log(ngx.DEBUG, "response is ",x)
      return cjson.decode(x).items, nil
  end

  -- Expose interface.
  local _M = {}
  _M.exit_401 = exit_401
  _M.exit_500 = exit_500
  _M.get_auth_token = get_auth_token
  _M.get_user_id = get_user_id
  _M.get_user_role = get_user_role
  _M.get_user_namespaces = get_user_namespaces
  _M.remove_content_len_header = remove_content_len_header
  _M.get_all_users = get_all_users
  _M.get_cluster = get_cluster
  _M.get_clusters = get_clusters
  _M.get_servicemonitor = get_servicemonitor

  return _M`
)
