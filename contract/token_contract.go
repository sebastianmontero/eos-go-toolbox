package contract

import (
	"strings"

	eosc "github.com/sebastianmontero/eos-go"
	"github.com/sebastianmontero/eos-go-toolbox/service"
	"github.com/sebastianmontero/eos-go-toolbox/util"
	"github.com/sebastianmontero/eos-go/token"
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

func NewTokenContractWithDefaultContract(eos *service.EOS, contract string) *TokenContract {
	return &TokenContract{
		&Contract{
			EOS:          eos,
			ContractName: contract,
		},
	}
}

func (m *TokenContract) CreateToken(contract, issuerName, maxSupply interface{}, failIfExists bool) (*service.PushTransactionFullResp, error) {

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
	return m.CreateTokenBase(m.getContract(contract), data, failIfExists)
}

func (m *TokenContract) CreateTokenBase(contract, data interface{}, failIfExists bool) (*service.PushTransactionFullResp, error) {

	contract = m.getContract(contract)
	resp, err := m.ExecActionC(contract, contract, "create", data)
	if err != nil {
		if failIfExists || !strings.Contains(err.Error(), "token with symbol already exists") {
			return nil, err
		}
	}
	return resp, nil
}

func (m *TokenContract) Issue(contract, to, quantity interface{}, memo string) (*service.PushTransactionFullResp, error) {
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
	return m.ExecActionC(m.getContract(contract), to, "issue", data)
}

func (m *TokenContract) Transfer(contract, from, to, quantity interface{}, memo string) (*service.PushTransactionFullResp, error) {

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
	return m.ExecActionC(m.getContract(contract), from, "transfer", data)
}

func (m *TokenContract) GetBalance(account, symbol, contract interface{}) (*eosc.Asset, error) {
	return m.EOS.GetBalance(account, symbol, m.getContract(contract))
}

func (m *TokenContract) GetStat(symbol, contract interface{}) (*eosc.GetCurrencyStatsResp, error) {
	return m.EOS.GetCurrencyStat(symbol, m.getContract(contract))
}

func (m *TokenContract) getContract(contract interface{}) interface{} {

	if contract == nil {
		if m.ContractName == "" {
			panic("Token contract must be specified as a defualt contract was not provided")
		}
		return m.ContractName
	}
	return contract
}
