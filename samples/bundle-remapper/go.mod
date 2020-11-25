module HARP.elastic.co/remapper

go 1.15

replace github.com/elastic/harp => ../../

// Snyk finding
replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

require (
	github.com/elastic/harp v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.16.0
)
