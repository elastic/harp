package harp

default keep = false

keep {
    input.annotations["infosec.elastic.co/v1/SecretPolicy#severity"] == "moderate"
    input.secrets.data[_].key == "cookieEncryptionKey"
}
