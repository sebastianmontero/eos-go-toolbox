package service_test

import (
	"testing"

	eosc "github.com/eoscanada/eos-go"
	"github.com/sebastianmontero/eos-go-toolbox/service"
	"gotest.tools/assert"
)

func TestGetAccount(t *testing.T) {
	E.Setup(t)
	eos := service.NewEOS(E.A)
	accountName := "usera"
	accountData, err := eos.GetAccount(accountName)
	assert.NilError(t, err)
	assert.Equal(t, accountData.AccountName, eosc.AN(accountName))
}

func TestGetAccountNonExistantAccount(t *testing.T) {
	E.Setup(t)
	eos := service.NewEOS(E.A)
	accountName := "nonexistant"
	accountData, err := eos.GetAccount(accountName)
	assert.NilError(t, err)
	assert.Assert(t, accountData == nil)
}
