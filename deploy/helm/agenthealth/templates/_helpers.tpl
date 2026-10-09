{{- define "agenthealth.fullname" -}}
{{- printf "%s-agenthealth" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "agenthealth.pod" -}}
template:
  metadata:
    labels:
      app.kubernetes.io/instance: {{ .Release.Name | quote }}
    annotations:
      checksum/config: {{ .Values.config | sha256sum | quote }}
  spec:
    automountServiceAccountToken: false
    terminationGracePeriodSeconds: 10
    securityContext:
      runAsNonRoot: true
      runAsUser: 65532
      runAsGroup: 65532
      seccompProfile:
        type: RuntimeDefault
    restartPolicy: {{ if eq .Values.mode "serve" }}Always{{ else }}Never{{ end }}
    containers:
      - name: agenthealth
        image: {{ if .Values.image.digest }}{{ printf "%s@%s" .Values.image.repository .Values.image.digest | quote }}{{ else }}{{ printf "%s:%s" .Values.image.repository .Values.image.tag | quote }}{{ end }}
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        args:
{{- if eq .Values.mode "serve" }}
          - serve
          - /etc/agenthealth/config.yaml
          - --listen
          - 0.0.0.0:8081
{{- if .Values.ahpTokenEnv }}
          - --token-env
          - {{ .Values.ahpTokenEnv | quote }}
{{- end }}
        ports:
          - name: ahp
            containerPort: 8081
        startupProbe:
          httpGet: {path: /live, port: ahp}
          periodSeconds: 5
          timeoutSeconds: 2
          failureThreshold: 12
        livenessProbe:
          httpGet: {path: /live, port: ahp}
          periodSeconds: 10
          timeoutSeconds: 2
        readinessProbe:
          httpGet: {path: /ready, port: ahp}
          periodSeconds: 5
          timeoutSeconds: 2
{{- else }}
          - check
          - /etc/agenthealth/config.yaml
          - --format
          - json
{{- end }}
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop: [ALL]
        resources:
{{ toYaml .Values.resources | indent 10 }}
{{- if .Values.secretEnv }}
        env:
{{- range .Values.secretEnv }}
          - name: {{ .name | quote }}
            valueFrom:
              secretKeyRef:
                name: {{ .secretName | quote }}
                key: {{ .key | quote }}
{{- end }}
{{- end }}
        volumeMounts:
          - name: config
            mountPath: /etc/agenthealth
            readOnly: true
    volumes:
      - name: config
        configMap:
          name: {{ default (include "agenthealth.fullname" .) .Values.existingConfigMap | quote }}
{{- end -}}
