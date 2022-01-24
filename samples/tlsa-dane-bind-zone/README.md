# Generate Bind DNS zone file with TLSA entries

## Template

> `$cert` will be a `*x509.Certificate` instance - https://pkg.go.dev/crypto/x509#Certificate

```zone
{{ $cert := parsePemCertificate (.Values.cert) -}}
;subject = {{ $cert.Subject.ToRDNSequence }}
;issuer = {{ $cert.Issuer.ToRDNSequence }}
;notBefore = {{ $cert.NotBefore }}
;notAfter = {{ $cert.NotAfter }}
;
_dane.example.com. IN TLSA 2 1 1 {{ toTLSA 1 1 $cert | upper }}
_25._tcp.smtp.example.com. IN CNAME _dane.example.com.
_587._tcp.smtp.example.com. IN CNAME _dane.example.com.
```

## Output

```sh
$ harp template \
    --set-file cert=$HOME/certs/root_ca.crt \
    --in template.yaml
;subject = CN=HomeLab Root CA
;issuer = CN=HomeLab Root CA
;notBefore = 2020-09-30 10:52:42 +0000 UTC
;notAfter = 2030-09-28 10:52:42 +0000 UTC
;
_dane.example.com. IN TLSA 2 1 1 E8636BB93C9ED85A2484C828A08407541D1AD572A633D7B3F935E9548A51180B
_25._tcp.smtp.example.com. IN CNAME _dane.example.com.
_587._tcp.smtp.example.com. IN CNAME _dane.example.com.
```
