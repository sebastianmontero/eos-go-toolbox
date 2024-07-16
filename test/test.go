package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sebastianmontero/eos-go"
	"github.com/sebastianmontero/eos-go-toolbox/service"
	"gotest.tools/assert"
)

type TestUtil struct {
	t      *testing.T
	client *service.EOS
}

func NewTestUtil(t *testing.T, client *service.EOS) *TestUtil {
	return &TestUtil{t: t, client: client}
}

func (m *TestUtil) AssertAction(contract eos.AccountName, actionName eos.ActionName, actionData map[string]interface{}, actionPos int) map[string]interface{} {
	action, err := m.client.GetActionAt(contract, actionName, actionPos)
	assert.NilError(m.t, err)
	actionsJSON, err := json.Marshal(action)
	assert.NilError(m.t, err)
	m.t.Logf("actions: %v", string(actionsJSON))
	assert.Assert(m.t, action != nil, "No action at pos: %v", actionPos)
	actualData := action.Params
	for key, value := range actionData {
		valueJSON, err := json.Marshal(value)
		assert.NilError(m.t, err)
		var valueInterface interface{}
		err = json.Unmarshal(valueJSON, &valueInterface)
		assert.NilError(m.t, err)
		assert.DeepEqual(m.t, actualData[key], valueInterface)
	}
	return actualData
}

func (m *TestUtil) AssertNotifications(account eos.AccountName, notifications [][]string) {
	m.BaseAssertNotifications(account, "notify", notifications, false)
}

func (m *TestUtil) AssertNotificationsExact(account eos.AccountName, notifications [][]string) {
	m.BaseAssertNotifications(account, "notify", notifications, true)
}

func (m *TestUtil) BaseAssertNotifications(account eos.AccountName, actionName eos.ActionName, notifications [][]string, verifyExactQuantity bool) {
	numActions := len(notifications)
	maxNumActions := numActions
	if verifyExactQuantity {
		maxNumActions = maxNumActions + 1
	}
	actions, err := m.client.GetActions(account, actionName, maxNumActions)
	assert.NilError(m.t, err)
	// actionsJSON, err := json.Marshal(actions)
	// assert.NilError(m.t, err)
	// m.t.Logf("actions: %v", string(actionsJSON))
	assert.Equal(m.t, len(actions), numActions)
	for i, msgs := range notifications {
		action := actions[numActions-i-1]
		actualData := action.Params
		// m.t.Logf("actualData: %v", action)
		assert.DeepEqual(m.t, actualData["user"], string(account))
		for _, msg := range msgs {
			notification := actualData["msg"].(string)
			assert.Assert(m.t, strings.Contains(notification, msg), "Notification: %v does not contain msg: %v", notification, msg)
		}
	}
}
