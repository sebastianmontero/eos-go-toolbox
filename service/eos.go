package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/eoscanada/eos-go"
	eosc "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/msig"
	"github.com/eoscanada/eos-go/system"
	"github.com/sebastianmontero/eos-go-toolbox/dto"
	"github.com/sebastianmontero/eos-go-toolbox/util"
)

var EOSIOKey = "5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3"

const retries = 10
const retrySleep = 2

type TableScope struct {
	Code  string `json:"code"`
	Scope string `json:"scope"`
	Table string `json:"table"`
	Payer string `json:"payer"`
	Count uint64 `json:"count"`
}

type TableScopesResp struct {
	More   string
	Scopes []*TableScope
}

type PushTransactionFullResp struct {
	*eosc.PushTransactionFullResp
}

func (m *PushTransactionFullResp) String() string {
	return fmt.Sprintf("TransactionID: %v\n", m.TransactionID)
}

type ProposeResponse struct {
	*eosc.PushTransactionFullResp
	ProposalName eosc.Name
}

func (m *ProposeResponse) String() string {
	return fmt.Sprintf("ProposalName: %v %v", m.ProposalName, m.PushTransactionFullResp)
}

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

func NameFromString(s string) (eos.Name, error) {
	n, err := eos.StringToName(s)
	if err != nil {
		return "", err
	}
	if s != eos.NameToString(n) {
		return "", fmt.Errorf("invalid eos name: %v", s)
	}
	return eos.Name(s), nil
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
	// fmt.Println("Error: ", errMsg)
	return strings.Contains(errMsg, "deadline") ||
		strings.Contains(errMsg, "connection reset by peer") ||
		strings.Contains(errMsg, "Transaction took too long") ||
		strings.Contains(errMsg, "exceeded the current CPU usage limit") ||
		strings.Contains(errMsg, "ABI serialization time has exceeded") ||
		strings.Contains(errMsg, "Invalid Reference Block: Transaction's reference block did not match. Is this transaction from a different fork")

}

func (m *EOS) SimpleTrx(contract, actionName, permissionLevel, data interface{}) (*eosc.PushTransactionFullResp, error) {
	action, err := m.BuildAction(contract, actionName, permissionLevel, data, retries)
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
	action, err := m.BuildAction(contract, actionName, permissionLevel, data, retries)
	if err != nil {
		return nil, err
	}
	fmt.Println("Action Data: ", action.ActionData.Data)
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

func (m *EOS) BuildTrx(expireIn time.Duration, actions ...*eosc.Action) (*eosc.Transaction, error) {
	txOpts := &eosc.TxOptions{}
	if err := txOpts.FillFromChain(context.Background(), m.API); err != nil {
		return nil, fmt.Errorf("failed getting txOptions to build trx, error: %v", err)
	}
	tx := eosc.NewTransaction(actions, txOpts)
	if expireIn > 0 {
		tx.SetExpiration(expireIn)
	}
	return tx, nil
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

// SetContract sets contract, if publicKey is not nil it creates the account
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

func (m *EOS) GetSetContractActions(accountName interface{}, wasmFile, abiFile string) ([]*eosc.Action, error) {

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
	return []*eosc.Action{
		setCodeAction,
		setAbiAction,
	}, nil
}

func (m *EOS) GetAccountPermission(accountName interface{}, permissionName string) (*eosc.Permission, error) {
	account, err := m.GetAccount(accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get account object for name: %v, error: %v", accountName, err)
	}
	for _, permission := range account.Permissions {
		if permission.PermName == permissionName {
			return &permission, nil
		}
	}
	return nil, nil
}

func (m *EOS) SetEOSIOCode(accountName interface{}) (bool, error) {

	codePermissionAction, err := m.GetSetEOSIOCodeAction(accountName)
	if err != nil {
		return false, fmt.Errorf("failed setting eosio.code permission for account: %v, error: %v", accountName, err)
	}
	if codePermissionAction != nil {
		_, err = m.Trx(retries, codePermissionAction)
		if err != nil {
			return false, fmt.Errorf("error setting eosio.code permission for account: %v, error: %v", accountName, err)
		}
		return true, nil
	} else {
		return false, nil
	}
}

func (m *EOS) ProposeSetEOSIOCode(proposerName interface{}, requested []eosc.PermissionLevel, expireIn time.Duration, accountName interface{}) (*ProposeResponse, error) {

	codePermissionAction, err := m.GetSetEOSIOCodeAction(accountName)
	if err != nil {
		return nil, fmt.Errorf("failed proposing set eosio.code permission for account: %v, error: %v", accountName, err)
	}
	if codePermissionAction != nil {
		response, err := m.ProposeMultiSig(proposerName, requested, expireIn, codePermissionAction)
		if err != nil {
			return nil, fmt.Errorf("error proposing set eosio.code permission for account: %v, error: %v", accountName, err)
		}
		return response, nil
	} else {
		return nil, nil
	}
}

func (m *EOS) GetSetEOSIOCodeAction(accountName interface{}) (*eosc.Action, error) {

	acct, err := util.ToAccountName(accountName)
	if err != nil {
		return nil, err
	}

	permission, err := m.GetAccountPermission(accountName, "active")
	if err != nil {
		return nil, fmt.Errorf("failed to get active permission for account: %v, error: %v", accountName, err)
	}
	accountAuthorityPos, err := FindAccountAuthority(permission, acct, "eosio.code")
	if err != nil {
		return nil, fmt.Errorf("failed to check if eosio.code is already set, error: %v", err)
	}
	authority := permission.RequiredAuth
	if accountAuthorityPos != -1 {
		accountAuthority := &authority.Accounts[accountAuthorityPos]
		if accountAuthority.Weight == uint16(permission.RequiredAuth.Threshold) {
			return nil, nil
		}
		accountAuthority.Weight = uint16(permission.RequiredAuth.Threshold)
	} else {
		authority.Accounts = append(authority.Accounts, eosc.PermissionLevelWeight{
			Permission: eosc.PermissionLevel{
				Actor:      acct,
				Permission: "eosio.code",
			},
			Weight: uint16(permission.RequiredAuth.Threshold),
		})
	}
	codePermissionAction := system.NewUpdateAuth(acct,
		"active",
		"owner",
		authority,
		"owner")
	return codePermissionAction, nil
}

func FindAccountAuthority(permission *eosc.Permission, accountName, permissionName interface{}) (int, error) {
	acct, err := util.ToAccountName(accountName)
	if err != nil {
		return -1, err
	}
	permName, err := util.ToPermissionName(permissionName)
	if err != nil {
		return -1, err
	}
	for i, account := range permission.RequiredAuth.Accounts {
		if account.Permission.Actor == acct && account.Permission.Permission == permName {
			return i, nil
		}
	}
	return -1, nil
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

func (m *EOS) LinkPermission(accountName, actionName, permissionName interface{}, failIfExists bool) error {
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
		if failIfExists || !strings.Contains(err.Error(), "new requirement is same as old") {
			return fmt.Errorf("error linking permission: %v, to action %v:%v, error: %v", permission, acct, action, err)
		}
	}
	return nil
}

func (m *EOS) AreTablesEmpty(code string, tables []string) (bool, error) {
	for _, table := range tables {
		empty, err := m.IsTableEmpty(code, table)
		if err != nil {
			return false, err
		}
		if !empty {
			return false, nil
		}
	}
	return true, nil
}

func (m *EOS) IsTableEmpty(code, table string) (bool, error) {
	scopes, err := m.GetAllTableScopes(code, table)
	if err != nil {
		return false, fmt.Errorf("error getting table: %v scopes, error: %v", table, err)
	}
	for _, scope := range scopes {
		empty, err := m.IsTableScopeEmpty(code, scope.Scope, table)
		if err != nil {
			return false, err
		}
		if !empty {
			return false, nil
		}
	}
	return true, nil
}

func (m *EOS) IsTableScopeEmpty(code, scope, table string) (bool, error) {
	req := &eosc.GetTableRowsRequest{
		Code:  code,
		Scope: scope,
		Table: table,
		Limit: 1,
	}
	var rows []interface{}
	err := m.GetTableRows(*req, &rows)
	if err != nil {
		return false, fmt.Errorf("error getting table: %v, scope: %v rows, error: %v", table, scope, err)
	}
	return len(rows) == 0, nil
}

func (m *EOS) GetAllTableRows(req eosc.GetTableRowsRequest, keyName string, structuredRows interface{}) error {
	allRows, err := m.GetAllTableRowsAsMap(req, keyName)
	if err != nil {
		return err
	}
	serialized, err := json.Marshal(allRows)
	if err != nil {
		return fmt.Errorf("failed marshalling rows to json %v", err)
	}
	err = json.Unmarshal(serialized, structuredRows)
	if err != nil {
		return fmt.Errorf("failed unmarshalling json to structured rows: %v", err)
	}
	return nil

}

func (m *EOS) GetAllTableRowsAsMap(req eosc.GetTableRowsRequest, keyName string) ([]map[string]interface{}, error) {
	return m.GetAllTableRowsFromAsMap(req, keyName, "")
}

func (m *EOS) GetAllTableRowsFromAsMap(req eosc.GetTableRowsRequest, keyName, startFrom string) ([]map[string]interface{}, error) {
	allRows := make([]map[string]interface{}, 0)
	lowerBound := startFrom
	for {
		req.LowerBound = lowerBound
		req.Limit = 1000
		var rows []map[string]interface{}
		err := m.GetTableRows(req, &rows)
		if err != nil {
			return nil, fmt.Errorf("failed getting table rows %v", err)
		}
		if lowerBound != startFrom {
			rows = rows[1:]
		}
		if len(rows) == 0 {
			return allRows, nil
		}
		lowerBound = fmt.Sprintf("%v", (rows[len(rows)-1][keyName]))
		allRows = append(allRows, rows...)
	}

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

func (m *EOS) GetAllTableScopes(code, table string) ([]*TableScope, error) {
	scopes := make([]*TableScope, 0)
	req := eosc.GetTableByScopeRequest{
		Code:  code,
		Table: table,
		Limit: 10000,
	}
	for {
		resp, err := m.GetTableScopes(req)
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, resp.Scopes...)
		if resp.More == "" {
			return scopes, nil
		}
		req.LowerBound = resp.More
	}
}

func (m *EOS) GetTableScopes(request eosc.GetTableByScopeRequest) (*TableScopesResp, error) {
	return m.GetTableScopesRetries(request, retries)
}

func (m *EOS) GetTableScopesRetries(request eosc.GetTableByScopeRequest, retries int) (*TableScopesResp, error) {

	response, err := m.API.GetTableByScope(context.Background(), request)
	if err != nil {
		if retries > 0 {
			if isRetryableError(err) {
				time.Sleep(time.Duration(retrySleep) * time.Second)
				return m.GetTableScopesRetries(request, retries-1)
			}
		}
		return nil, fmt.Errorf("get table scopes %v", err)
	}
	var scopes []*TableScope

	err = json.Unmarshal(response.Rows, &scopes)
	if err != nil {
		return nil, fmt.Errorf("json to structs %v", err)
	}
	return &TableScopesResp{
		More:   response.More,
		Scopes: scopes,
	}, nil
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
	case string, eosc.Name, eosc.AccountName:
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

func (m *EOS) BuildAction(contractName, actionName, permissionLevel, data interface{}, retries int) (*eosc.Action, error) {
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

	switch data.(type) {
	case nil:
		actionData = eosc.NewActionDataFromHexData([]byte("{}"))
	// case map[string]interface{}:
	// 	actionBinary, err := m.API.ABIJSONToBin(context.Background(), contract, eosc.Name(action), v)
	// 	if err != nil {
	// 		if retries > 0 {
	// 			if isRetryableError(err) {
	// 				time.Sleep(time.Duration(retrySleep) * time.Second)
	// 				return m.BuildAction(contractName, actionName, permissionLevel, data, retries-1)
	// 			}
	// 		}
	// 		return nil, fmt.Errorf("cannot pack action data for action: %v", err)
	// 	}
	// 	actionData = eosc.NewActionDataFromHexData([]byte(actionBinary))

	default:
		// fmt.Println("Encoding data: ", data)
		actionData = eosc.NewActionData(data)
	}

	return &eosc.Action{
		Account:       contract,
		Name:          action,
		Authorization: []eosc.PermissionLevel{pl},
		ActionData:    actionData,
	}, nil
}

func (m *EOS) ProposeMultiSig(proposerName interface{}, requested []eosc.PermissionLevel, expireIn time.Duration, actions ...*eosc.Action) (*ProposeResponse, error) {
	proposer, err := util.ToAccountName(proposerName)
	if err != nil {
		return nil, err
	}
	proposalName := eosc.Name(util.RandAccountName())
	transaction, err := m.BuildTrx(expireIn, actions...)
	if err != nil {
		return nil, fmt.Errorf("failed to propose multi sig, unable to build transaction, err: %v", err)
	}
	proposeAction := msig.NewPropose(proposer, proposalName, requested, transaction)
	resp, err := m.Trx(retries, proposeAction)
	if err != nil {
		return nil, fmt.Errorf("failed pushing propose transaction, error: %v", err)
	}
	return &ProposeResponse{
		ProposalName:            proposalName,
		PushTransactionFullResp: resp,
	}, nil
}

func (m *EOS) ApproveMultiSig(proposerName interface{}, proposalName interface{}, permissionLevel eosc.PermissionLevel) (*eosc.PushTransactionFullResp, error) {
	proposer, err := util.ToAccountName(proposerName)
	if err != nil {
		return nil, err
	}
	proposal, err := util.ToName(proposalName)
	if err != nil {
		return nil, fmt.Errorf("failed to approve multi sig, invalid proposal name, err: %v", err)
	}

	approveAction := msig.NewApprove(proposer, proposal, permissionLevel)
	resp, err := m.Trx(retries, approveAction)
	if err != nil {
		return nil, fmt.Errorf("failed pushing approve transaction, error: %v", err)
	}
	return resp, nil
}

func (m *EOS) ExecuteMultiSig(proposerName interface{}, proposalName interface{}, executerName interface{}) (*eosc.PushTransactionFullResp, error) {
	proposer, err := util.ToAccountName(proposerName)
	if err != nil {
		return nil, err
	}
	proposal, err := util.ToName(proposalName)
	if err != nil {
		return nil, fmt.Errorf("failed to approve multi sig, invalid proposal name, err: %v", err)
	}

	executer, err := util.ToAccountName(executerName)
	if err != nil {
		return nil, err
	}

	execAction := msig.NewExec(proposer, proposal, executer)
	resp, err := m.Trx(retries, execAction)
	if err != nil {
		return nil, fmt.Errorf("failed pushing exec transaction, error: %v", err)
	}
	return resp, nil
}

func (m *EOS) ProposeSetContractMultiSig(proposerName interface{}, requested []eosc.PermissionLevel, expireIn time.Duration, accountName interface{}, wasmFile, abiFile string) (*ProposeResponse, error) {

	actions, err := m.GetSetContractActions(accountName, wasmFile, abiFile)
	if err != nil {
		return nil, fmt.Errorf("failed building set contract actions, error: %v", err)
	}
	return m.ProposeMultiSig(proposerName, requested, expireIn, actions...)

}

func (m *EOS) GetInfo() (*eosc.InfoResp, error) {
	info, err := m.API.GetInfo(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed getting chain info, error: %v", err)
	}
	return info, nil
}

func (m *EOS) GetActions(account eosc.AccountName, action eosc.ActionName, quantity int) ([]*dto.Action, error) {
	time.Sleep(time.Second)
	info, err := m.GetInfo()
	if err != nil {
		return nil, err
	}
	// infostr, _ := json.Marshal(info)
	// fmt.Println("Info: \n", string(infostr))
	actions := make([]*dto.Action, 0)
	for blockNum := int64(info.HeadBlockNum + 10); blockNum >= 0; blockNum-- {
		if len(actions) >= quantity {
			// actstr, _ := json.Marshal(actions)
			// fmt.Printf("Actions account: %v action:%v actions:%v \n", account, action, string(actstr))
			return actions, nil
		}
		block, err := m.GetBlock(uint32(blockNum))
		if err != nil {
			if !strings.Contains(err.Error(), "block trace missing") {
				return nil, err
			}
		}
		if block != nil {
			// fmt.Println("Block: \n", block)
			actions = append(actions, block.GetActions(account, action, quantity-len(actions))...)
		}
	}
	// actstr, _ := json.Marshal(actions)
	// fmt.Printf("Actions account: %v action:%v actions:%v \n", account, action, string(actstr))
	return actions, nil
}

func (m *EOS) GetActionAt(account eosc.AccountName, action eosc.ActionName, pos int) (*dto.Action, error) {
	actions, err := m.GetActions(account, action, pos+1)
	if err != nil {
		return nil, err
	}
	if len(actions) > pos {
		return actions[pos], nil
	}
	return nil, nil
}

func (m *EOS) GetBlock(blockNum uint32) (out *dto.Block, err error) {
	err = m.call(context.Background(), "trace_api", "get_block", M{"block_num": blockNum}, &out)
	if err != nil {
		err = fmt.Errorf("failed getting block: %v, error: %v", blockNum, err)
	}
	return
}

func (m *EOS) call(ctx context.Context, baseAPI string, endpoint string, body interface{}, out interface{}) error {
	jsonBody, err := enc(body)
	if err != nil {
		return err
	}

	targetURL := fmt.Sprintf("%s/v1/%s/%s", m.API.BaseURL, baseAPI, endpoint)
	req, err := http.NewRequest("POST", targetURL, jsonBody)
	if err != nil {
		return fmt.Errorf("NewRequest: %w", err)
	}

	for k, v := range m.API.Header {
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header[k] = append(req.Header[k], v...)
	}

	if m.API.Debug {
		// Useful when debugging API calls
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(string(requestDump))
		fmt.Println("")
	}

	resp, err := m.API.HttpClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("%s: %w", req.URL.String(), err)
	}
	defer resp.Body.Close()

	var cnt bytes.Buffer
	_, err = io.Copy(&cnt, resp.Body)
	if err != nil {
		return fmt.Errorf("Copy: %w", err)
	}

	if resp.StatusCode == 404 {
		var apiErr eosc.APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return ErrNotFound
		}
		return apiErr
	}

	if resp.StatusCode > 299 {
		var apiErr eosc.APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return fmt.Errorf("%s: status code=%d, body=%s", req.URL.String(), resp.StatusCode, cnt.String())
		}

		// Handle cases where some API calls (/v1/chain/get_account for example) returns a 500
		// error when retrieving data that does not exist.
		if apiErr.IsUnknownKeyError() {
			return ErrNotFound
		}

		return apiErr
	}

	if m.API.Debug {
		fmt.Println("RESPONSE:")
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(cnt.String())
		fmt.Println("-------------------------------")
		fmt.Printf("%q\n", responseDump)
		fmt.Println("")
	}

	if err := json.Unmarshal(cnt.Bytes(), &out); err != nil {
		return fmt.Errorf("Unmarshal: %w", err)
	}

	return nil
}

var ErrNotFound = errors.New("resource not found")

type M map[string]interface{}

func enc(v interface{}) (io.Reader, error) {
	if v == nil {
		return nil, nil
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}
