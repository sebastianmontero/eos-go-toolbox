package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	eosc "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"github.com/sebastianmontero/eos-go-toolbox/util"
)

var EOSIOKey = "5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3"

const retries = 5
const retrySleep = 2

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

func GetEOSIOKey() *ecc.PrivateKey {
	key, _ := GetKey(EOSIOKey)
	return key
}

func GetEOSIOPublicKey() *ecc.PublicKey {
	key, _ := GetKey(EOSIOKey)
	pubKey := key.PublicKey()
	return &pubKey
}

func GetKey(privateKey string) (*ecc.PrivateKey, error) {
	key, err := ecc.NewPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed creating key for string key: %v, error: %v", privateKey, err)
	}
	return key, nil
}

func (m *EOS) AddKey(privateKey string) (*ecc.PublicKey, error) {
	// logger.Infof("PKey: %v", pkey)
	if m.API.Signer == nil {
		m.API.SetSigner(&eosc.KeyBag{})
	}
	key, err := GetKey(privateKey)
	if err != nil {
		return nil, err
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
				time.Sleep(time.Duration(retrySleep) * time.Second)
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
		strings.Contains(errMsg, "Transaction took too long") ||
		strings.Contains(errMsg, "exceeded the current CPU usage limit") ||
		strings.Contains(errMsg, "ABI serialization time has exceeded")

}

func (m *EOS) SimpleTrx(contract, actionName, permissionLevel, data interface{}) (*eosc.PushTransactionFullResp, error) {
	action, err := m.buildAction(contract, actionName, permissionLevel, data, retries)
	if err != nil {
		return nil, err
	}
	return m.Trx(retries, action)
}

func (m *EOS) DebugTrx(contract, actionName, permissionLevel, data interface{}) (*eosc.PushTransactionFullResp, error) {
	txOpts := &eosc.TxOptions{}
	if err := txOpts.FillFromChain(context.Background(), m.API); err != nil {
		return nil, err
	}
	action, err := m.buildAction(contract, actionName, permissionLevel, data, retries)
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

func (m *EOS) CreateAccount(accountName interface{}, publicKey *ecc.PublicKey, failIfExists bool) (eosc.AccountName, error) {
	account, err := util.ToAccountName(accountName)
	if err != nil {
		return "", err
	}
	if !failIfExists {
		accountData, err := m.GetAccount(accountName)
		if err != nil {
			return "", err
		}
		if accountData != nil {
			return account, nil
		}
	}
	_, err = m.Trx(retries, system.NewNewAccount("eosio", account, *publicKey))
	if err != nil {
		if failIfExists || !strings.Contains(err.Error(), "name is already taken") {
			return "", err
		}
	}
	return account, nil
}

func (m *EOS) CreateRandomAccount(publicKey *ecc.PublicKey) (eosc.AccountName, error) {

	if publicKey == nil {
		publicKey = GetEOSIOPublicKey()
	}
	return m.CreateAccount(util.RandAccountName(), publicKey, true)
}

func (m *EOS) GetAccount(accountName interface{}) (*eosc.AccountResp, error) {
	account, err := util.ToAccountName(accountName)
	if err != nil {
		return nil, err
	}
	accountData, err := m.API.GetAccount(context.Background(), account)
	if err != nil {
		if strings.Contains(err.Error(), "resource not found") {
			return nil, nil
		}
		return nil, err
	}
	return accountData, nil
}

//SetContract sets contract, if publicKey is not nil it creates the account
func (m *EOS) SetContract(accountName interface{}, wasmFile, abiFile string, publicKey *ecc.PublicKey) (*eosc.PushTransactionFullResp, error) {

	if publicKey != nil {
		_, err := m.CreateAccount(accountName, publicKey, false)
		if err != nil {
			return nil, err
		}
	}
	account, err := util.ToAccountName(accountName)
	if err != nil {
		return nil, err
	}
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

func (m *EOS) SetEOSIOCode(accountName interface{}, publicKey *ecc.PublicKey) error {
	acct, err := util.ToAccountName(accountName)
	if err != nil {
		return err
	}
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

	_, err = m.Trx(retries, codePermissionAction)
	if err != nil {
		return fmt.Errorf("error setting eosio.code permission for account: %v, error: %v", accountName, err)
	}
	return nil
}

func (m *EOS) CreateSimplePermission(accountName, newPermissionName interface{}, publicKey *ecc.PublicKey) error {
	acct, err := util.ToAccountName(accountName)
	if err != nil {
		return err
	}
	newPermission, err := util.ToPermissionName(newPermissionName)
	if err != nil {
		return err
	}
	permission := eosc.PermissionName("active")
	codePermissionAction := system.NewUpdateAuth(
		acct,
		newPermission,
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

	_, err = m.Trx(retries, codePermissionAction)
	if err != nil {
		return fmt.Errorf("error creating permission: %v, for account: %v, error: %v", newPermission, acct, err)
	}
	return nil
}

func (m *EOS) LinkPermission(accountName, actionName, permissionName interface{}) error {
	acct, err := util.ToAccountName(accountName)
	if err != nil {
		return err
	}

	action, err := util.ToActionName(actionName)
	if err != nil {
		return err
	}
	permission, err := util.ToPermissionName(permissionName)
	if err != nil {
		return err
	}
	linkAction := system.NewLinkAuth(acct, acct, action, eosc.PermissionName(permission))
	_, err = m.Trx(retries, linkAction)
	if err != nil {
		return fmt.Errorf("error linking permission: %v, to action %v:%v, error: %v", permission, acct, action, err)
	}
	return nil
}

func (m *EOS) GetTableRows(request eosc.GetTableRowsRequest, rows interface{}) error {
	return m.GetTableRowsRetries(request, rows, retries)
}

func (m *EOS) GetTableRowsRetries(request eosc.GetTableRowsRequest, rows interface{}, retries int) error {

	request.JSON = true
	response, err := m.API.GetTableRows(context.Background(), request)
	if err != nil {
		if retries > 0 {
			if isRetryableError(err) {
				time.Sleep(time.Duration(retrySleep) * time.Second)
				return m.GetTableRowsRetries(request, rows, retries-1)
			}
		}
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
	case string, eosc.Name:
		vUint64, err := eosc.StringToName(fmt.Sprintf("%v", v))
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

func (m *EOS) GetBalance(accountName, symbol, contractName interface{}) (*eosc.Asset, error) {
	account, err := util.ToAccountName(accountName)
	if err != nil {
		return nil, err
	}
	contract, err := util.ToAccountName(contractName)
	if err != nil {
		return nil, err
	}
	sym, err := util.ToSymbol(symbol)
	if err != nil {
		return nil, err
	}
	assets, err := m.API.GetCurrencyBalance(context.Background(), account, sym.MustSymbolCode().String(), contract)
	if err != nil {
		return nil, fmt.Errorf("failed to get currency balance, error: %v", err)
	}
	if len(assets) > 0 {
		return &assets[0], nil
	}
	return nil, nil
}

func (m *EOS) GetCurrencyStat(symbol, contractName interface{}) (*eosc.GetCurrencyStatsResp, error) {

	contract, err := util.ToAccountName(contractName)
	if err != nil {
		return nil, err
	}
	sym, err := util.ToSymbol(symbol)
	if err != nil {
		return nil, err
	}
	stats, err := m.API.GetCurrencyStats(context.Background(), contract, sym.MustSymbolCode().String())
	if err != nil {
		return nil, fmt.Errorf("failed to get currency stat, error: %v", err)
	}
	return stats, nil
}

func (m *EOS) buildAction(contractName, actionName, permissionLevel, data interface{}, retries int) (*eosc.Action, error) {
	var actionData eosc.ActionData

	contract, err := util.ToAccountName(contractName)
	if err != nil {
		return nil, err
	}

	action, err := util.ToActionName(actionName)
	if err != nil {
		return nil, err
	}

	pl, err := util.ToPermissionLevel(permissionLevel)
	if err != nil {
		return nil, err
	}

	switch v := data.(type) {
	case nil:
		actionData = eosc.NewActionDataFromHexData([]byte("{}"))
	case map[string]interface{}:
		actionBinary, err := m.API.ABIJSONToBin(context.Background(), contract, eosc.Name(action), v)
		if err != nil {
			if retries > 0 {
				if isRetryableError(err) {
					time.Sleep(time.Duration(retrySleep) * time.Second)
					return m.buildAction(contractName, actionName, permissionLevel, data, retries-1)
				}
			}
			return nil, fmt.Errorf("cannot pack action data for action: %v", err)
		}
		actionData = eosc.NewActionDataFromHexData([]byte(actionBinary))

	default:
		actionData = eosc.NewActionData(data)
	}

	return &eosc.Action{
		Account:       contract,
		Name:          action,
		Authorization: []eosc.PermissionLevel{pl},
		ActionData:    actionData,
	}, nil
}
