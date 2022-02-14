package "app/credentials/database" {
    description = ""

    secrets = {
        "USER" = "{{ strongPassword }}"
    }
}
