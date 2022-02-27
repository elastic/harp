package harp

default compliant = false

compliant {
    input.annotations["infosec.elastic.co/v1/SecretPolicy#severity"] == "moderate"
    secrets := ["DB_HOST","DB_NAME","DB_USER","DB_PASSWORD"]
    # Has all secrets
    input.secrets.data[_].key == secrets[_]
}
