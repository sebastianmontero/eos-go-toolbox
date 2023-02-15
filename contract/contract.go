// The MIT License (MIT)

// Copyright (c) 2020, Digital Scarcity

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package contract

import (
	"fmt"
	"time"

	eos "github.com/eoscanada/eos-go"
	"github.com/sebastianmontero/eos-go-toolbox/service"
)

type Contract struct {
	ContractName string
	EOS          *service.EOS
}

// ExecActionC Accepts contract name on which to execute the action
func (m *Contract) ExecActionC(contract, permissionLevel, action, data interface{}) (*service.PushTransactionFullResp, error) {
	resp, err := m.EOS.SimpleTrx(contract, action, permissionLevel, data)
	if err != nil {
		return nil, err
	}
	return &service.PushTransactionFullResp{PushTransactionFullResp: resp}, nil
}

func (m *Contract) ExecAction(permissionLevel, action, data interface{}) (*service.PushTransactionFullResp, error) {
	return m.ExecActionC(m.ContractName, permissionLevel, action, data)
}

func (m *Contract) ProposeAction(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, permissionLevel, actionName, data interface{}) (*service.ProposeResponse, error) {
	action, err := m.EOS.BuildAction(m.ContractName, actionName, permissionLevel, data, 5)
	if err != nil {
		return nil, fmt.Errorf("failed proposing multisig action, error building action: %v", err)
	}
	return m.EOS.ProposeMultiSig(proposerName, requested, expireIn, action)
}

func (m *Contract) GetTableRows(request eos.GetTableRowsRequest, rows interface{}) error {

	if request.Code == "" {
		request.Code = string(m.ContractName)
	}
	if request.Scope == "" {
		request.Scope = string(m.ContractName)
	}
	if request.Limit == 0 {
		request.Limit = 100
	}
	return m.EOS.GetTableRows(request, rows)
}

func (m *Contract) GetAllTableRows(request eos.GetTableRowsRequest, keyName string, rows interface{}) error {

	if request.Code == "" {
		request.Code = string(m.ContractName)
	}
	if request.Scope == "" {
		request.Scope = string(m.ContractName)
	}

	return m.EOS.GetAllTableRows(request, keyName, rows)
}

func (m *Contract) GetAllTableRowsAsMap(request eos.GetTableRowsRequest, keyName string) ([]map[string]interface{}, error) {

	if request.Code == "" {
		request.Code = string(m.ContractName)
	}
	if request.Scope == "" {
		request.Scope = string(m.ContractName)
	}

	return m.EOS.GetAllTableRowsAsMap(request, keyName)
}

func (m *Contract) GetTableScopes(request eos.GetTableByScopeRequest) (*service.TableScopesResp, error) {

	if request.Code == "" {
		request.Code = string(m.ContractName)
	}

	if request.Limit == 0 {
		request.Limit = 100
	}
	return m.EOS.GetTableScopes(request)
}

func (m *Contract) GetAllTableScopes(table string) ([]*service.TableScope, error) {
	return m.EOS.GetAllTableScopes(string(m.ContractName), table)
}

func (m *Contract) IsTableScopeEmpty(scope, table string) (bool, error) {
	return m.EOS.IsTableScopeEmpty(string(m.ContractName), scope, table)
}

func (m *Contract) IsTableEmpty(table string) (bool, error) {
	return m.EOS.IsTableEmpty(string(m.ContractName), table)
}

func (m *Contract) AreTablesEmpty(tables []string) (bool, error) {
	return m.EOS.AreTablesEmpty(string(m.ContractName), tables)
}
