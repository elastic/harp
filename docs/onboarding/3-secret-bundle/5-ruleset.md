# RuleSet

- [RuleSet](#ruleset)
  - [Query language](#query-language)
    - [CEL Expressions](#cel-expressions)
      - [Package matchers](#package-matchers)
      - [Secret context](#secret-context)

## Query language
### CEL Expressions

#### Package matchers

* `p.match_path(globstring) bool` - Returns true if the package name match the given Glob pattern.
* `p.match_label(globstring) bool` - Returns true if one of the package labels match the given Glob pattern.
* `p.match_annotation(globstring) bool` - Returns true if one of the package annotations match the given Glob pattern.
* `p.match_secret(globstring) bool` - Returns true if the package has a secret with given pattern.
* `p.has_secret(string) bool` - Returns true if the package has a secret with given key.
* `p.has_all_secrets(...string) bool` - Returns true if the package has all secrets with given keys.
* `p.is_cso_compliant() bool` - Returns true is the package name is CSO compliant.

#### Secret context

* `p.secret(string) Secret` - Returns the secret context matching the secret key of the package.
* `p.secret(string).is_required()` - Flag the given secret key as required.
* `p.secret(string).is_base64()` - Flag the given secret value has a valid base64 encoded string.
* `p.secret(string).is_url()` - Flag the given secret value as a valid URL.
* `p.secret(string).is_uuid()` - Flag the given secret value as a valid UUID.
* `p.secret(string).is_email()` - Flag the given secret value as a valid email.
* `p.secret(string).is_json()` - Flag the given secret value as a valid JSON.

---

* [Previous topic](4-patch.md)
* [Index](../)
