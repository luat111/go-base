{{/*
Expand the name of the chart.
*/}}
{{- define "go-base.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "go-base.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "go-base.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "go-base.labels" -}}
helm.sh/chart: {{ include "go-base.chart" . }}
{{ include "go-base.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "go-base.selectorLabels" -}}
app.kubernetes.io/name: {{ include "go-base.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "go-base.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "go-base.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Linkerd annotations for pod injection
*/}}
{{- define "go-base.linkerd.annotations" -}}
{{- if .Values.serviceMesh.enabled }}
{{- if eq .Values.serviceMesh.provider "linkerd" }}
linkerd.io/inject: {{ .Values.serviceMesh.linkerd.inject }}
{{- if .Values.serviceMesh.linkerd.proxy.resources }}
config.linkerd.io/proxy-cpu-request: {{ .Values.serviceMesh.linkerd.proxy.resources.cpu.request | quote }}
config.linkerd.io/proxy-memory-request: {{ .Values.serviceMesh.linkerd.proxy.resources.memory.request | quote }}
config.linkerd.io/proxy-cpu-limit: {{ .Values.serviceMesh.linkerd.proxy.resources.cpu.limit | quote }}
config.linkerd.io/proxy-memory-limit: {{ .Values.serviceMesh.linkerd.proxy.resources.memory.limit | quote }}
{{- end }}
{{- if .Values.serviceMesh.linkerd.proxy.nativeSidecar }}
config.alpha.linkerd.io/proxy-enable-native-sidecar: "true"
{{- end }}
{{- if .Values.serviceMesh.linkerd.http2.enabled }}
{{- if .Values.serviceMesh.linkerd.http2.autoUpgrade }}
config.linkerd.io/enable-http2-upgrade: "true"
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Search engine host
*/}}
{{- define "go-base.search.host" -}}
{{- if eq .Values.search.provider "manticore" }}
{{- printf "%s-manticore" (include "go-base.fullname" .) }}
{{- else if eq .Values.search.provider "elasticsearch" }}
{{- printf "%s-elasticsearch" (include "go-base.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Search engine HTTP port
*/}}
{{- define "go-base.search.port" -}}
{{- if eq .Values.search.provider "manticore" }}
{{- .Values.search.manticore.service.httpPort }}
{{- else if eq .Values.search.provider "elasticsearch" }}
{{- .Values.search.elasticsearch.service.httpPort }}
{{- end }}
{{- end }}

{{/*
Search engine URL
*/}}
{{- define "go-base.search.url" -}}
{{- printf "http://%s:%v" (include "go-base.search.host" .) (include "go-base.search.port" .) }}
{{- end }}
{{- end }}
