# Understanding _helpers.tpl

## What is _helpers.tpl?

`_helpers.tpl` is a **special Helm template file** that contains **reusable template functions** (also called "named templates" or "partials"). 

### Key Characteristics:
- **Prefix with underscore** (`_`) - Tells Helm this file doesn't generate a Kubernetes manifest
- **Contains only template definitions** - No actual Kubernetes resources
- **Shared across all templates** - Can be used in any `.yaml` file in the chart
- **DRY principle** - Define once, use everywhere

---

## Template Functions Explained

### 1. `go-base.name` (Lines 4-6)

```go
{{- define "go-base.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}
```

**Purpose**: Returns the chart name (or override)

**Logic**:
1. Use `.Values.nameOverride` if provided, otherwise use `.Chart.Name` (from Chart.yaml)
2. Truncate to 63 characters (Kubernetes label limit)
3. Remove trailing hyphens

**Example Output**: `go-base`

**Usage in templates**:
```yaml
name: {{ include "go-base.name" . }}
```

---

### 2. `go-base.fullname` (Lines 11-22)

```go
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
```

**Purpose**: Creates a unique, fully qualified name for resources

**Logic**:
1. If `fullnameOverride` is set → use it
2. Otherwise, combine release name + chart name
3. If release name already contains chart name → use release name only
4. Truncate to 63 chars and remove trailing hyphens

**Example Outputs**:
- Release: `my-app`, Chart: `go-base` → `my-app-go-base`
- Release: `go-base-prod`, Chart: `go-base` → `go-base-prod` (no duplication)

**Usage**:
```yaml
metadata:
  name: {{ include "go-base.fullname" . }}
```

**Why Important**: Ensures unique resource names when installing multiple releases

---

### 3. `go-base.chart` (Lines 27-29)

```go
{{- define "go-base.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}
```

**Purpose**: Creates chart identifier with version

**Logic**:
1. Combine chart name + version (e.g., `go-base-1.0.0`)
2. Replace `+` with `_` (Kubernetes label compatibility)
3. Truncate and clean

**Example Output**: `go-base-1.0.0`

**Usage**: In the `helm.sh/chart` label

---

### 4. `go-base.labels` (Lines 34-41)

```go
{{- define "go-base.labels" -}}
helm.sh/chart: {{ include "go-base.chart" . }}
{{ include "go-base.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
```

**Purpose**: Standard set of labels for all resources

**Generates**:
```yaml
helm.sh/chart: go-base-1.0.0
app.kubernetes.io/name: go-base
app.kubernetes.io/instance: my-release
app.kubernetes.io/version: "1.0.0"
app.kubernetes.io/managed-by: Helm
```

**Why Important**: 
- **Tracking**: Know which chart created the resource
- **Versioning**: Track app versions
- **Management**: Identify Helm-managed resources

**Usage**:
```yaml
metadata:
  labels:
    {{- include "go-base.labels" . | nindent 4 }}
```

---

### 5. `go-base.selectorLabels` (Lines 46-49)

```go
{{- define "go-base.selectorLabels" -}}
app.kubernetes.io/name: {{ include "go-base.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
```

**Purpose**: Minimal labels for pod selectors

**Generates**:
```yaml
app.kubernetes.io/name: go-base
app.kubernetes.io/instance: my-release
```

**Why Separate from `labels`**: 
- **Immutable selectors**: Service/Deployment selectors can't change after creation
- **Minimal set**: Only include labels that won't change
- **Consistency**: Same selector across Deployment, Service, HPA

**Usage**:
```yaml
# In Deployment
spec:
  selector:
    matchLabels:
      {{- include "go-base.selectorLabels" . | nindent 6 }}

# In Service
spec:
  selector:
    {{- include "go-base.selectorLabels" . | nindent 4 }}
```

---

### 6. `go-base.serviceAccountName` (Lines 54-60)

```go
{{- define "go-base.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "go-base.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
```

**Purpose**: Determines which ServiceAccount to use

**Logic**:
1. If creating ServiceAccount → use custom name or fullname
2. If not creating → use provided name or "default"

**Example Outputs**:
- Create=true, name="" → `my-release-go-base`
- Create=true, name="custom" → `custom`
- Create=false, name="" → `default`
- Create=false, name="existing" → `existing`

**Usage**:
```yaml
spec:
  serviceAccountName: {{ include "go-base.serviceAccountName" . }}
```

---

## How Templates Are Used

### In deployment.yaml:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "go-base.fullname" . }}        # ← Uses fullname
  labels:
    {{- include "go-base.labels" . | nindent 4 }} # ← Uses labels
spec:
  selector:
    matchLabels:
      {{- include "go-base.selectorLabels" . | nindent 6 }} # ← Uses selector labels
  template:
    metadata:
      labels:
        {{- include "go-base.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ include "go-base.serviceAccountName" . }} # ← Uses SA name
```

---

## Helm Template Syntax Explained

### `{{- define "name" -}}`
- Defines a named template
- `-` removes whitespace before/after

### `{{ include "name" . }}`
- Calls a template
- `.` passes the current context (values, chart info, etc.)

### `| nindent 4`
- Pipes output to `nindent` function
- Indents by 4 spaces (for YAML formatting)

### `.Chart.Name`, `.Values.xxx`, `.Release.Name`
- Built-in Helm objects
- Access chart metadata, values, and release info

---

## Benefits of _helpers.tpl

✅ **Consistency** - Same naming/labeling across all resources
✅ **DRY** - Define once, use everywhere
✅ **Maintainability** - Change logic in one place
✅ **Best Practices** - Follows Kubernetes recommended labels
✅ **Flexibility** - Easy to override via values

---

## Common Patterns

### Adding Custom Labels
```go
{{- define "go-base.labels" -}}
helm.sh/chart: {{ include "go-base.chart" . }}
{{ include "go-base.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
environment: {{ .Values.global.environment }}  # ← Custom label
{{- end }}
```

### Adding Custom Helper
```go
{{/*
Get database host
*/}}
{{- define "go-base.databaseHost" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-pgpool" (include "go-base.fullname" .) }}
{{- else }}
{{- .Values.externalDatabase.host }}
{{- end }}
{{- end }}
```

---

## Summary

The `_helpers.tpl` file is the **foundation** of your Helm chart:

1. **Naming** - Consistent resource names
2. **Labeling** - Standard Kubernetes labels
3. **Reusability** - Shared logic across templates
4. **Maintainability** - Single source of truth

Every template in your chart uses these helpers to ensure consistency and follow Kubernetes best practices!
