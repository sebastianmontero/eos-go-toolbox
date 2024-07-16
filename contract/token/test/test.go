package test

import (
	"testing"

	"github.com/sebastianmontero/eos-go"
	"github.com/sebastianmontero/eos-go-toolbox/contract"
	"github.com/sebastianmontero/eos-go-toolbox/service"
	"github.com/sebastianmontero/eos-go-toolbox/test"
	"github.com/sebastianmontero/eos-go-toolbox/util"
	"gotest.tools/assert"
)

type TestUtil struct {
	t             *testing.T
	eosTestUtil   *test.TestUtil
	tokenContract *contract.TokenContract
}

func NewTestUtil(t *testing.T, client *service.EOS) *TestUtil {
	return &TestUtil{
		t:             t,
		eosTestUtil:   test.NewTestUtil(t, client),
		tokenContract: contract.NewTokenContract(client),
	}
}

func (m *TestUtil) AssertBalance(tokenContract, account eos.AccountName, balance interface{}) {

	bal, err := util.ToAsset(balance)
	assert.NilError(m.t, err)
	actualBalance, err := m.tokenContract.GetBalance(account, bal.Symbol, tokenContract)
	assert.NilError(m.t, err)
	assert.Assert(m.t, actualBalance != nil)
	assert.Equal(m.t, *actualBalance, bal)
}

func (m *TestUtil) AssertHasNoBalance(tokenContract, account eos.AccountName, symbol interface{}) {

	symb, err := util.ToSymbol(symbol)
	assert.NilError(m.t, err)
	actualBalance, err := m.tokenContract.GetBalance(account, symb, tokenContract)
	assert.NilError(m.t, err)
	assert.Assert(m.t, actualBalance == nil)
}

func (m *TestUtil) AssertStat(tokenContract eos.AccountName, expected *eos.GetCurrencyStatsResp) {

	actual, err := m.tokenContract.GetStat(expected.Supply.Symbol, tokenContract)
	assert.NilError(m.t, err)
	assert.Assert(m.t, actual != nil)
	assert.DeepEqual(m.t, actual.Supply, expected.Supply)
	assert.DeepEqual(m.t, actual.MaxSupply, expected.MaxSupply)
	assert.Equal(m.t, actual.Issuer, expected.Issuer)
}
