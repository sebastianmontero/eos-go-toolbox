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
	eos "github.com/eoscanada/eos-go"
	"github.com/sebastianmontero/eos-go-toolbox/service"
)

type Contract struct {
	ContractName string
	EOS          *service.EOS
}

//ExecActionC Accepts contract name on which to execute the action
func (m *Contract) ExecActionC(contract, actor, action string, data interface{}) (*eos.PushTransactionFullResp, error) {
	return m.EOS.SimpleTrx(contract, action, actor, data)
}

func (m *Contract) ExecAction(actor, action string, data interface{}) (*eos.PushTransactionFullResp, error) {
	return m.ExecActionC(m.ContractName, actor, action, data)
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
