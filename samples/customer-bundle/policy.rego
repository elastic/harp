package main

# ------------------------------------------------------------------------------

is_template {
    input.kind == "BundleTemplate"
}

# ------------------------------------------------------------------------------

deny[msg] {
    not is_template
    msg = "HRP-BT-SCHEMA-00001 - The given BundleTemplate source doesn't declare a 'BundleTemplate' resource."
}

deny[msg] {
    is_template
    not input.meta
    msg = "HRP-BT-SCHEMA-00002 - Missing 'meta' section."
}

# ------------------------------------------------------------------------------

violation[msg] {
    not input.meta
    msg = "HRP-BT-CORE-00001 - The 'meta.name' must be defined."
}

violation[msg] {
    secrets := input.spec.namespaces.infrastructure[_].regions[_].services[_].secrets[_]
    not secrets.description
    msg = "HRP-BT-INFRA-00001 - Infrastructure secrets must have a description."
}

# ------------------------------------------------------------------------------

