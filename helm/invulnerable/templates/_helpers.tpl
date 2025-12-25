{{/*
Expand the name of the chart.
*/}}
{{- define "invulnerable.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "invulnerable.fullname" -}}
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
{{- define "invulnerable.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "invulnerable.labels" -}}
helm.sh/chart: {{ include "invulnerable.chart" . }}
{{ include "invulnerable.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "invulnerable.selectorLabels" -}}
app.kubernetes.io/name: {{ include "invulnerable.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "invulnerable.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "invulnerable.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "invulnerable.frontend.labels" -}}
{{ include "invulnerable.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "invulnerable.frontend.selectorLabels" -}}
{{ include "invulnerable.selectorLabels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Backend labels
*/}}
{{- define "invulnerable.backend.labels" -}}
{{ include "invulnerable.labels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "invulnerable.backend.selectorLabels" -}}
{{ include "invulnerable.selectorLabels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Scanner labels
*/}}
{{- define "invulnerable.scanner.labels" -}}
{{ include "invulnerable.labels" . }}
app.kubernetes.io/component: scanner
{{- end }}

{{/*
Scanner selector labels
*/}}
{{- define "invulnerable.scanner.selectorLabels" -}}
{{ include "invulnerable.selectorLabels" . }}
app.kubernetes.io/component: scanner
{{- end }}

{{/*
Backend API endpoint
*/}}
{{- define "invulnerable.backend.apiEndpoint" -}}
{{- if .Values.scanner.apiEndpoint }}
{{- .Values.scanner.apiEndpoint }}
{{- else }}
{{- printf "http://%s-backend.%s.svc.cluster.local:%d" (include "invulnerable.fullname" .) .Release.Namespace (.Values.backend.service.port | int) }}
{{- end }}
{{- end }}

{{/*
Frontend image
*/}}
{{- define "invulnerable.frontend.image" -}}
{{- $registry := .Values.frontend.image.registry | default .Values.image.registry -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry .Values.frontend.image.repository (.Values.frontend.image.tag | default .Chart.AppVersion) }}
{{- else }}
{{- printf "%s:%s" .Values.frontend.image.repository (.Values.frontend.image.tag | default .Chart.AppVersion) }}
{{- end }}
{{- end }}

{{/*
Backend image
*/}}
{{- define "invulnerable.backend.image" -}}
{{- $registry := .Values.backend.image.registry | default .Values.image.registry -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry .Values.backend.image.repository (.Values.backend.image.tag | default .Chart.AppVersion) }}
{{- else }}
{{- printf "%s:%s" .Values.backend.image.repository (.Values.backend.image.tag | default .Chart.AppVersion) }}
{{- end }}
{{- end }}

{{/*
Scanner image
*/}}
{{- define "invulnerable.scanner.image" -}}
{{- $registry := .Values.scanner.image.registry | default .Values.image.registry -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry .Values.scanner.image.repository (.Values.scanner.image.tag | default .Chart.AppVersion) }}
{{- else }}
{{- printf "%s:%s" .Values.scanner.image.repository (.Values.scanner.image.tag | default .Chart.AppVersion) }}
{{- end }}
{{- end }}
