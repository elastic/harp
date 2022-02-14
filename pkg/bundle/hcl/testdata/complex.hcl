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
        "PASSWORD" = "{{ strongPassword | b64enc }}"
        "USER" = "admin-{{ randAlpha 8 }}"
    }
}
