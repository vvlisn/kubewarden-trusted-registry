module github.com/kubewarden/kubewarden-trusted-registry

go 1.22

toolchain go1.23.4

require (
	github.com/deckarep/golang-set/v2 v2.7.0
	github.com/francoispqt/onelog v0.0.0-20190306043706-8c2bb31b10a4
	github.com/kubewarden/k8s-objects v1.29.0-kw1
	github.com/kubewarden/policy-sdk-go v0.11.0
	github.com/tidwall/gjson v1.18.0
	github.com/wapc/wapc-guest-tinygo v0.3.3
)

require (
	github.com/francoispqt/gojay v0.0.0-20181220093123-f2cc13a668ca // indirect
	github.com/go-openapi/strfmt v0.21.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
)

replace github.com/go-openapi/strfmt => github.com/kubewarden/strfmt v0.1.3
