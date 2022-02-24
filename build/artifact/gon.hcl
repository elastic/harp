source = [
    "harp-darwin-amd64",
    "harp-darwin-amd64-fips",
    "harp-darwin-arm64",
    "harp-darwin-arm64-fips"
]

bundle_id = "co.elastic.harp"

apple_id {
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "9470D0A7B70090A8EF31C3B33AB3868B38B27A3D"
}
