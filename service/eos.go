package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/eoscanada/eos-go"
	eosc "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
)

var EOSIOKey = "5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3"

var retries = 3

type EOS struct {
	API         *eosc.API
	SetSignerFn func(*eosc.API)
}

func NewEOSFromUrl(url string) *EOS {
	api := eosc.New(url)
	return NewEOS(api)
}

func NewEOS(api *eosc.API) *EOS {
	return &EOS{
		API: api,
	}
}

func (m *EOS) AddKey(privateKey string) (*ecc.PublicKey, error) {
	// logger.Infof("PKey: %v", pkey)
	if m.API.Signer == nil {
		m.API.SetSigner(&eosc.KeyBag{})
	}
	key, err := ecc.NewPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("privateKey parameter is not a valid format: %s", err)
	}
	err = m.API.Signer.ImportPrivateKey(context.Background(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed importing key, error: %v", err)
	}
	publicKey := key.PublicKey()
	return &publicKey, nil
}

func (m *EOS) AddEOSIOKey() (*ecc.PublicKey, error) {
	return m.AddKey(EOSIOKey)
}

func (m *EOS) Trx(retries int, actions ...*eosc.Action) (*eosc.PushTransactionFullResp, error) {
	// for _, action := range actions {
	// 	logger.Infof("Trx Account: %v Name: %v, Authorization: %v, Data: %v", action.Account, action.Name, action.Authorization, action.ActionData)

	// }
	if m.API.Signer == nil && m.SetSignerFn != nil {
		m.SetSignerFn(m.API)
	}
	resp, err := m.API.SignPushActions(context.Background(), actions...)
	if err != nil {
		if retries > 0 {
			if isRetryableError(err) {
				time.Sleep(3 * time.Second)
				return m.Trx(retries-1, actions...)
			}
		}
		return nil, fmt.Errorf("failed to push trx: %v, error: %v", actions, err)
	}
	return resp, nil
}

func isRetryableError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "connection reset by peer") ||
		strings.Contains(errMsg, "Transaction took too long")
}

func (m *EOS) SimpleTrx(contract, actionName string, permissionLevel interface{}, data interface{}) (*eosc.PushTransactionFullResp, error) {
	action, err := m.buildAction(contract, actionName, permissionLevel, data)
	if err != nil {
		return nil, err
	}
	return m.Trx(retries, action)
}

func (m *EOS) DebugTrx(contract, actionName string, permissionLevel interface{}, data interface{}) (*eosc.PushTransactionFullResp, error) {
	txOpts := &eosc.TxOptions{}
	if err := txOpts.FillFromChain(context.Background(), m.API); err != nil {
		return nil, err
	}
	action, err := m.buildAction(contract, actionName, permissionLevel, data)
	if err != nil {
		return nil, err
	}

	tx := eosc.NewTransaction([]*eosc.Action{action}, txOpts)
	signedTx, packedTx, err := m.API.SignTransaction(context.Background(), tx, txOpts.ChainID, eosc.CompressionNone)
	if err != nil {
		return nil, err
	}

	content, err := json.MarshalIndent(signedTx, "", "  ")
	if err != nil {
		return nil, err
	}

	fmt.Println(string(content))
	return m.API.PushTransaction(context.Background(), packedTx)
}

func (m *EOS) CreateAccount(accountName string, publicKey *ecc.PublicKey, failIfExists bool) (eosc.AccountName, error) {
	account := eosc.AccountName(accountName)
	_, err := m.Trx(retries, system.NewNewAccount("eosio", account, *publicKey))
	if err != nil {
		if failIfExists || !strings.Contains(err.Error(), "name is already taken") {
			return "", err
		}
	}
	return account, nil
}

//SetContract sets contract, if publicKey is not nil it creates the account
func (m *EOS) SetContract(accountName, wasmFile, abiFile string, publicKey *ecc.PublicKey) (*eosc.PushTransactionFullResp, error) {

	if publicKey != nil {
		_, err := m.CreateAccount(accountName, publicKey, false)
		if err != nil {
			return nil, err
		}
	}
	account := eosc.AccountName(accountName)
	setCodeAction, err := system.NewSetCode(account, wasmFile)
	if err != nil {
		return nil, fmt.Errorf("unable construct set_code action: %v", err)
	}

	setAbiAction, err := system.NewSetABI(account, abiFile)
	if err != nil {
		return nil, fmt.Errorf("unable construct set_abi action: %v", err)
	}
	resp, err := m.Trx(retries, setCodeAction, setAbiAction)
	if err != nil {
		if !strings.Contains(err.Error(), "contract is already running this version of code") {
			return nil, err
		}
	}
	return resp, nil
}

func (m *EOS) SetEOSIOCode(accountName string, publicKey *ecc.PublicKey) error {
	acct := eosc.AN(accountName)
	codePermissionAction := system.NewUpdateAuth(acct,
		"active",
		"owner",
		eosc.Authority{
			Threshold: 1,
			Keys: []eosc.KeyWeight{{
				PublicKey: *publicKey,
				Weight:    1,
			}},
			Accounts: []eosc.PermissionLevelWeight{{
				Permission: eosc.PermissionLevel{
					Actor:      acct,
					Permission: "eosio.code",
				},
				Weight: 1,
			}},
			Waits: []eosc.WaitWeight{},
		}, "owner")

	_, err := m.Trx(retries, codePermissionAction)
	if err != nil {
		return fmt.Errorf("error setting eosio.code permission for account: %v, error: %v", accountName, err)
	}
	return nil
}

func (m *EOS) CreateSimplePermission(account, newPermission string, publicKey *ecc.PublicKey) error {
	acct := eosc.AN(account)
	permission := eosc.PermissionName("active")
	codePermissionAction := system.NewUpdateAuth(acct,
		eosc.PermissionName(newPermission),
		permission,
		eosc.Authority{
			Threshold: 1,
			Keys: []eosc.KeyWeight{{
				PublicKey: *publicKey,
				Weight:    1,
			}},
			Accounts: []eosc.PermissionLevelWeight{{
				Permission: eosc.PermissionLevel{
					Actor:      acct,
					Permission: permission,
				},
				Weight: 1,
			}},
			Waits: []eosc.WaitWeight{},
		}, permission)

	_, err := m.Trx(retries, codePermissionAction)
	if err != nil {
		return fmt.Errorf("error creating permission: %v, for account: %v, error: %v", newPermission, account, err)
	}
	return nil
}

func (m *EOS) LinkPermission(account, action, permission string) error {
	acct := eosc.AN(account)
	linkAction := system.NewLinkAuth(acct, acct, eos.ActN(action), eosc.PermissionName(permission))
	_, err := m.Trx(retries, linkAction)
	if err != nil {
		return fmt.Errorf("error linking permission: %v, to action %v:%v, error: %v", permission, account, action, err)
	}
	return nil
}

func (m *EOS) GetTableRows(request eosc.GetTableRowsRequest, rows interface{}) error {

	request.JSON = true
	response, err := m.API.GetTableRows(context.Background(), request)
	if err != nil {
		return fmt.Errorf("get table rows %v", err)
	}

	err = response.JSONToStructs(rows)
	if err != nil {
		return fmt.Errorf("json to structs %v", err)
	}
	return nil
}

func (m *EOS) GetComposedIndexValue(firstValue interface{}, secondValue interface{}) (string, error) {

	firstInt64, err := m.getUInt64Value(firstValue)
	if err != nil {
		return "", err
	}
	secondInt64, err := m.getUInt64Value(secondValue)
	if err != nil {
		return "", err
	}
	firstBigInt := big.NewInt(0)
	secondBigInt := big.NewInt(0)
	firstBigInt.SetUint64(firstInt64)
	secondBigInt.SetUint64(secondInt64)
	firstBigInt = firstBigInt.Lsh(firstBigInt, 64)
	firstBigInt = firstBigInt.Add(firstBigInt, secondBigInt)
	return firstBigInt.String(), nil
}

func (m *EOS) getUInt64Value(value interface{}) (uint64, error) {

	switch v := value.(type) {
	case string, eos.Name:
		vUint64, err := eos.StringToName(fmt.Sprintf("%v", v))
		if err != nil {
			return 0, fmt.Errorf("failed to convert status to uint64, err: %v", err)
		}
		return vUint64, nil
	case int, int32, int64, uint64:
		vUint64, err := strconv.ParseUint(fmt.Sprintf("%v", v), 10, 64)
		if err != nil {
			return 0, err
		}
		return vUint64, nil
	default:
		return 0, fmt.Errorf("unable to get uint64 value from: %v", value)
	}
}

func (m *EOS) GetBalance(account eosc.AccountName, symbol eosc.Symbol, contract eosc.AccountName) (*eosc.Asset, error) {

	assets, err := m.API.GetCurrencyBalance(context.Background(), account, symbol.MustSymbolCode().String(), contract)
	if err != nil {
		return nil, fmt.Errorf("failed to get currency balance, error: %v", err)
	}
	if len(assets) > 0 {
		return &assets[0], nil
	}
	return nil, nil
}

func (m *EOS) buildAction(contract, action string, permissionLevel interface{}, data interface{}) (*eosc.Action, error) {
	var actionData eosc.ActionData
	switch v := data.(type) {
	case nil:
		actionData = eosc.NewActionDataFromHexData([]byte("{}"))
	case map[string]interface{}:
		actionBinary, err := m.API.ABIJSONToBin(context.Background(), eosc.AN(contract), eosc.Name(action), v)
		if err != nil {
			return nil, fmt.Errorf("cannot pack action data for action: %v", err)
		}
		actionData = eosc.NewActionDataFromHexData([]byte(actionBinary))

	default:
		actionData = eosc.NewActionData(data)
	}

	pl, err := m.ParsePermissionLevel(permissionLevel)

	if err != nil {
		return nil, err
	}

	return &eosc.Action{
		Account:       eosc.AN(contract),
		Name:          eosc.ActN(action),
		Authorization: []eosc.PermissionLevel{pl},
		ActionData:    actionData,
	}, nil
}

func (m *EOS) ParsePermissionLevel(permissionLevel interface{}) (eosc.PermissionLevel, error) {
	switch v := permissionLevel.(type) {
	case eosc.PermissionLevel:
		return v, nil
	case eosc.AccountName, eosc.Name, string:
		pl, err := eosc.NewPermissionLevel(fmt.Sprintf("%v", v))
		if err != nil {
			return pl, fmt.Errorf("unable to parse permission level: %v, error: %v", v, err)
		}
		return pl, nil
	default:
		return eosc.PermissionLevel{}, fmt.Errorf("unable to parse permission level: %v, of type: %t", v, v)
	}
}
