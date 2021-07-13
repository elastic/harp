module github.com/elastic/harp/tools

go 1.16

replace github.com/elastic/go-licenser => github.com/elastic/go-licenser v0.3.2-0.20200604055331-209fde246f25

require (
	github.com/daixiang0/gci v0.2.8
	github.com/dvyukov/go-fuzz v0.0.0-20210602112143-b1f3d6f4ef4e
	github.com/elastic/go-licenser v0.0.0-00010101000000-000000000000
	github.com/elazarl/go-bindata-assetfs v1.0.1 // indirect
	github.com/fatih/color v1.12.0
	github.com/frapposelli/wwhrd v0.4.0
	github.com/golang/mock v1.6.0
	github.com/golangci/golangci-lint v1.41.1
	github.com/google/wire v0.5.0
	github.com/izumin5210/gex v0.6.1
	github.com/magefile/mage v1.11.0
	github.com/stephens2424/writerset v1.0.2 // indirect
	go.elastic.co/go-licence-detector v0.5.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gotest.tools/gotestsum v1.6.4
	mvdan.cc/gofumpt v0.1.1
)
