module github.com/elastic/harp

go 1.15

// Snyk finding
replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	cloud.google.com/go/storage v1.12.0
	github.com/Azure/azure-sdk-for-go v48.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.10 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/awnumar/memguard v0.22.2
	github.com/aws/aws-sdk-go v1.35.20
	github.com/basgys/goxml2json v1.1.0
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/blang/semver/v4 v4.0.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/cloudflare/tableflip v1.2.0
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/dnaeon/go-vcr v1.1.0 // indirect
	github.com/fatih/color v1.10.0
	github.com/fatih/structs v1.1.0
	github.com/fernet/fernet-go v0.0.0-20191111064656-eff2850e6001
	github.com/go-akka/configuration v0.0.0-20200606091224-a002c0330665
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.2
	github.com/google/gofuzz v1.2.0
	github.com/google/gops v0.3.13
	github.com/gosimple/slug v1.9.0
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hcl/v2 v2.7.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/iancoleman/strcase v0.1.2
	github.com/imdario/mergo v0.3.11
	github.com/jmespath/go-jmespath v0.4.0
	github.com/magefile/mage v1.10.0
	github.com/mcuadros/go-defaults v1.2.0
	github.com/oklog/run v1.1.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/pelletier/go-toml v1.8.1
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v0.0.0-00010101000000-000000000000 // indirect
	github.com/sethvargo/go-diceware v0.2.0
	github.com/sethvargo/go-password v0.2.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/spf13/afero v1.4.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/ugorji/go/codec v1.1.13
	github.com/zclconf/go-cty v1.7.0
	gitlab.com/NebulousLabs/merkletree v0.0.0-20200118113624-07fbf710afc4
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/square/go-jose.v2 v2.5.1
	sigs.k8s.io/yaml v1.2.0
)
