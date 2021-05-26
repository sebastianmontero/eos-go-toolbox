package contract

import (
	"strings"

	eosc "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
	"github.com/sebastianmontero/eos-go-toolbox/service"
	"github.com/sebastianmontero/eos-go-toolbox/util"
)

type TokenContract struct {
	*Contract
}

func NewTokenContract(eos *service.EOS) *TokenContract {
	return &TokenContract{
		&Contract{
			EOS: eos,
		},
	}
}

func (m *TokenContract) CreateToken(contract, issuerName, maxSupply interface{}, failIfExists bool) (*eosc.PushTransactionFullResp, error) {

	issuer, err := util.ToAccountName(issuerName)
	if err != nil {
		return nil, err
	}

	supply, err := util.ToAsset(maxSupply)
	if err != nil {
		return nil, err
	}

	data := token.Create{
		Issuer:        issuer,
		MaximumSupply: supply,
	}
	return m.CreateTokenBase(contract, data, failIfExists)
}

func (m *TokenContract) CreateTokenBase(contract, data interface{}, failIfExists bool) (*eosc.PushTransactionFullResp, error) {

	resp, err := m.ExecActionC(contract, contract, "create", data)
	if err != nil {
		if failIfExists || !strings.Contains(err.Error(), "token with symbol already exists") {
			return nil, err
		}
	}
	return resp, nil
}

func (m *TokenContract) Issue(contract, to, quantity interface{}, memo string) (*eosc.PushTransactionFullResp, error) {
	toAN, err := util.ToAccountName(to)
	if err != nil {
		return nil, err
	}

	qty, err := util.ToAsset(quantity)
	if err != nil {
		return nil, err
	}

	data := token.Issue{
		To:       toAN,
		Quantity: qty,
		Memo:     memo,
	}
	return m.ExecActionC(contract, contract, "issue", data)
}

func (m *TokenContract) Transfer(contract, from, to, quantity interface{}, memo string) (*eosc.PushTransactionFullResp, error) {

	fromAN, err := util.ToAccountName(from)
	if err != nil {
		return nil, err
	}

	toAN, err := util.ToAccountName(to)
	if err != nil {
		return nil, err
	}

	qty, err := util.ToAsset(quantity)
	if err != nil {
		return nil, err
	}

	data := token.Transfer{
		From:     fromAN,
		To:       toAN,
		Quantity: qty,
		Memo:     memo,
	}
	return m.ExecActionC(contract, from, "transfer", data)
}

func (m *TokenContract) GetBalance(account, symbol, contract interface{}) (*eosc.Asset, error) {
	return m.EOS.GetBalance(account, symbol, contract)
}

func (m *TokenContract) GetStat(symbol, contract interface{}) (*eosc.GetCurrencyStatsResp, error) {
	return m.EOS.GetCurrencyStat(symbol, contract)
}
