package dto

import (
	"encoding/json"
	"fmt"

	eosc "github.com/eoscanada/eos-go"
)

type Block struct {
	ID           eosc.Checksum256    `json:"id"`
	PreviousId   eosc.Checksum256    `json:"previous_id"`
	Number       uint32              `json:"number"`
	Timestamp    eosc.BlockTimestamp `json:"timestamp"`
	Producer     eosc.AccountName    `json:"producer"`
	Status       string              `json:"status"`
	Transactions []*Transaction      `json:"transactions"`
}

func (m *Block) GetActions(account eosc.AccountName, action eosc.ActionName, quantity int) []*Action {
	actions := make([]*Action, 0)
	for i := len(m.Transactions) - 1; i >= 0; i-- {
		trx := m.Transactions[i]
		for j := len(trx.Actions) - 1; j >= 0; j-- {

			if len(actions) >= quantity {
				return actions
			}
			act := trx.Actions[j]
			if act.IsAction(account, action) {
				actions = append(actions, act)
			}
		}
	}
	return actions
}

func (m *Block) String() string {
	str, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("error marshalling block to json, err: %v", err))
	}
	return string(str)
}

type Transaction struct {
	ID        eosc.Checksum256    `json:"id"`
	BlockNum  uint32              `json:"block_num"`
	BlockTime eosc.BlockTimestamp `json:"block_time"`
	Status    string              `json:"status"`
	Actions   []*Action           `json:"actions"`
}

type Action struct {
	GlobalSequence uint64           `json:"global_sequence"`
	Receiver       eosc.AccountName `json:"receiver"`
	Account        eosc.AccountName `json:"account"`
	Action         eosc.ActionName  `json:"action"`
	// Authorization  []PermissionLevel `json:"authorization,omitempty"`
	Params map[string]interface{} `json:"params,omitempty" eos:"-"`
}

func (m *Action) IsAction(account eosc.AccountName, action eosc.ActionName) bool {
	return m.Receiver == account && m.Action == action
}

// type PermissionLevel struct {
// 	Account    AccountName    `json:"account"`
// 	Permission PermissionName `json:"permission"`
// }
