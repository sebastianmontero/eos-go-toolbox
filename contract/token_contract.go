package contract

import (
	"fmt"
	"strings"

	eosc "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
	"github.com/sebastianmontero/eos-go-toolbox/service"
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

func (m *TokenContract) CreateToken(contract, issuer string, maxSupply eosc.Asset, failIfExists bool) (*eosc.PushTransactionFullResp, error) {
	return m.CreateTokenT(eosc.AN(contract), eosc.AN(issuer), maxSupply, failIfExists)
}

func (m *TokenContract) CreateTokenT(contract, issuer eosc.AccountName, maxSupply eosc.Asset, failIfExists bool) (*eosc.PushTransactionFullResp, error) {

	data := token.Create{
		Issuer:        issuer,
		MaximumSupply: maxSupply,
	}
	return m.CreateTokenBase(string(contract), data, failIfExists)
}

func (m *TokenContract) CreateTokenBase(contract string, data interface{}, failIfExists bool) (*eosc.PushTransactionFullResp, error) {

	resp, err := m.ExecActionC(contract, contract, "create", data)
	if err != nil {
		if failIfExists || !strings.Contains(err.Error(), "token with symbol already exists") {
			return nil, err
		}
	}
	return resp, nil
}

func (m *TokenContract) Issue(contract, to, quantity, memo string) (*eosc.PushTransactionFullResp, error) {
	qty, err := eosc.NewAssetFromString(quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quantity, error: %v", err)
	}
	return m.IssueT(eosc.AN(contract), eosc.AN(to), qty, memo)
}

func (m *TokenContract) IssueT(contract, to eosc.AccountName, quantity eosc.Asset, memo string) (*eosc.PushTransactionFullResp, error) {
	data := token.Issue{
		To:       to,
		Quantity: quantity,
		Memo:     memo,
	}
	return m.ExecActionC(string(contract), string(contract), "issue", data)
}

func (m *TokenContract) Transfer(contract, from, to, quantity, memo string) (*eosc.PushTransactionFullResp, error) {
	qty, err := eosc.NewAssetFromString(quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quantity, error: %v", err)
	}
	return m.TransferT(eosc.AN(contract), eosc.AN(from), eosc.AN(to), qty, memo)
}

func (m *TokenContract) TransferT(contract, from, to eosc.AccountName, quantity eosc.Asset, memo string) (*eosc.PushTransactionFullResp, error) {

	data := token.Transfer{
		From:     from,
		To:       to,
		Quantity: quantity,
		Memo:     memo,
	}
	return m.ExecActionC(string(contract), string(from), "transfer", data)
}

func (m *TokenContract) GetBalance(account eosc.AccountName, symbol eosc.Symbol, contract eosc.AccountName) (*eosc.Asset, error) {
	return m.EOS.GetBalance(account, symbol, contract)
}
