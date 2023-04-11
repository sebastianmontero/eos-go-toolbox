module github.com/sebastianmontero/eos-go-toolbox

go 1.16

require (
	github.com/digital-scarcity/eos-go-test v0.0.0-20221012234131-071de01a715c
	github.com/eoscanada/eos-go v0.10.2
	github.com/ethereum/go-ethereum v1.9.9 // indirect
	github.com/tidwall/sjson v1.0.4 // indirect
	gotest.tools v2.2.0+incompatible
)

// replace github.com/digital-scarcity/eos-go-test => ../eos-go-test
replace github.com/eoscanada/eos-go => ../eos-go
