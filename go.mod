module github.com/sebastianmontero/eos-go-toolbox

go 1.16

require (
	github.com/digital-scarcity/eos-go-test v0.0.0-20230415144134-50e76c085618
	github.com/sebastianmontero/eos-go v0.10.5-0.20251014033848-1f05f693154c
	gotest.tools v2.2.0+incompatible
)

// replace github.com/digital-scarcity/eos-go-test => ../eos-go-test
// replace github.com/sebastianmontero/eos-go => ../eos-go
