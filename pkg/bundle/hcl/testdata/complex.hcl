# Complex sample

# Bundle
annotations = {
    "secrets.elastic.co/leakSeverity" = "moderate"
}

# Packages
package "platform/security/databases/postgresql" {
    description = "Administrative access for database management."

    annotations = {
        "infosec.elastic.co/v1/SecretManagement#rotationPeriod" = "90"
        "infosec.elastic.co/v1/SecretManagement#generationDate" = "{{ now | isodate }}"
    }

    secrets = {
        "USER" = "admin-{{ randAlpha 8 }}"
        "PASSWORD" = "{{ strongPassword | b64enc }}"
    }
}
