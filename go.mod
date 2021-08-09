module github.com/elastic/harp

go 1.16

// Snyk finding
replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

// Nancy findings
replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/alessio/shellescape v1.4.1
	github.com/awnumar/memguard v0.22.2
	github.com/basgys/goxml2json v1.1.0
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/blang/semver/v4 v4.0.0
	github.com/cloudflare/tableflip v1.2.2
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/fatih/color v1.12.0
	github.com/fatih/structs v1.1.0
	github.com/fernet/fernet-go v0.0.0-20191111064656-eff2850e6001
	github.com/go-akka/configuration v0.0.0-20200606091224-a002c0330665
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4
	github.com/google/cel-go v0.7.3
	github.com/google/go-cmp v0.5.6
	github.com/google/gofuzz v1.2.0
	github.com/google/gops v0.3.19
	github.com/gosimple/slug v1.10.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hcl/v2 v2.10.1
	github.com/hashicorp/vault/api v1.1.1
	github.com/iancoleman/strcase v0.2.0
	github.com/imdario/mergo v0.3.12
	github.com/jmespath/go-jmespath v0.4.0
	github.com/magefile/mage v1.11.0
	github.com/mcuadros/go-defaults v1.2.0
	github.com/miscreant/miscreant.go v0.0.0-20200214223636-26d376326b75
	github.com/oklog/run v1.1.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/pelletier/go-toml v1.9.3
	github.com/pkg/errors v0.9.1
	github.com/sethvargo/go-diceware v0.2.1
	github.com/sethvargo/go-password v0.2.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go/codec v1.2.6
	github.com/zclconf/go-cty v1.9.0
	gitlab.com/NebulousLabs/merkletree v0.0.0-20200118113624-07fbf710afc4
	go.uber.org/zap v1.18.1
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210806184541-e5e7981a1069
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	google.golang.org/genproto v0.0.0-20210805201207-89edb61ffb67
	google.golang.org/grpc v1.39.1
	google.golang.org/protobuf v1.27.1
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	sigs.k8s.io/yaml v1.2.0
)
