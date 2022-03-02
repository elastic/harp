{{- $environments := .Values.environments -}}
{{- $services := .Values.services -}}
{{- $regions := .Values.regions -}}

{{ range $region := $regions -}}
package "platform/production/security/{{ $region }}/database/postgres/billing/admin_acccount" {
    description = "Administrative access for database management."

    annotations = {
        "infosec.elastic.co/v1/SecretManagement#rotationPeriod" = "90"
        "infosec.elastic.co/v1/SecretManagement#generationDate" = "{{ now | isodate }}"
    }

    secrets = {
        "PASSWORD" = "{{ strongPassword | b64enc }}"
        "USER" = "admin-{{ randAlpha 8 }}"
    }
}

{{- range $quality := $environments -}}
{{- range $srv := $services }}
package "app/{{ $quality }}/security/{{ $region }}/database/postgres/{{ $srv }}/usage_account" {
    description = "{{ $srv }} usage account."

    annotations = {
        "infosec.elastic.co/v1/SecretManagement#rotationPeriod" = "30"
        "infosec.elastic.co/v1/SecretManagement#generationDate" = "{{ now | isodate }}"
    }

    secrets = {
        "PASSWORD" = "{{ strongPassword | b64enc }}"
        "USER" = "{{ $srv }}-{{ randAlpha 8 }}"
    }
}
{{ end }}
{{ end }}
{{ end }}
