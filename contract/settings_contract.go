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
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	eos "github.com/sebastianmontero/eos-go"
	"github.com/sebastianmontero/eos-go-toolbox/dto"
	"github.com/sebastianmontero/eos-go-toolbox/service"
)

type ModifySettingArgs struct {
	Setter eos.AccountName `json:"setter"`
	Key    string          `json:"key"`
	Value  *dto.FlexValue  `json:"value"`
}

func (m *ModifySettingArgs) String() string {
	result, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Failed marshalling round: %v", err))
	}
	return string(result)
}

type EraseSettingArgs struct {
	Setter eos.AccountName `json:"setter"`
	Key    string          `json:"key"`
}

func (m *EraseSettingArgs) String() string {
	result, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Failed marshalling round: %v", err))
	}
	return string(result)
}

type Setting struct {
	ID          uint64             `json:"id"`
	Key         string             `json:"key"`
	Values      []*dto.FlexValue   `json:"values"`
	CreatedDate eos.BlockTimestamp `json:"created_date"`
	UpdatedDate eos.BlockTimestamp `json:"updated_date"`
}

func (m *Setting) String() string {
	result, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Failed marshalling round: %v", err))
	}
	return string(result)
}

func (m *Setting) ValueCount() int {
	return len(m.Values)
}

func (m *Setting) Get(pos int) (*dto.FlexValue, error) {
	if pos < m.ValueCount() {
		return m.Values[0], nil
	}
	return nil, fmt.Errorf("setting has no value at pos: %v, value count: %v", pos, m.ValueCount())
}

func (m *Setting) GetAsName(pos int) (eos.Name, error) {
	v, err := m.Get(pos)
	if err != nil {
		return eos.Name(""), err
	}
	return v.Name(), nil
}

func (m *Setting) GetAsString(pos int) (string, error) {
	v, err := m.Get(pos)
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

func (m *Setting) GetAsTimePoint(pos int) (eos.TimePoint, error) {
	v, err := m.Get(pos)
	if err != nil {
		return eos.TimePoint(0), err
	}
	return v.TimePoint(), nil
}

func (m *Setting) GetAsAsset(pos int) (eos.Asset, error) {
	v, err := m.Get(pos)
	if err != nil {
		return eos.Asset{}, err
	}
	return v.Asset(), nil
}

func (m *Setting) GetAsInt64(pos int) (int64, error) {
	v, err := m.Get(pos)
	if err != nil {
		return 0, err
	}
	return v.Int64(), nil
}

func (m *Setting) GetAsUint32(pos int) (uint32, error) {
	v, err := m.Get(pos)
	if err != nil {
		return 0, err
	}
	return v.Uint32(), nil
}

type SettingsContract struct {
	*Contract
}

func NewSettingsContract(eos *service.EOS, contractName string) *SettingsContract {
	return &SettingsContract{
		&Contract{
			EOS:          eos,
			ContractName: contractName,
		},
	}
}

func (m *SettingsContract) SettingAsUint32(key string) (uint32, error) {
	setting, err := m.GetSetting(key)
	if err != nil {
		return 0, err
	}
	value, err := setting.GetAsUint32(0)
	if err != nil {
		return 0, fmt.Errorf("failed getting setting: %v as uint32, error: %v", value, err)
	}
	return value, nil
}

func (m *SettingsContract) SettingAsAsset(key string) (eos.Asset, error) {
	setting, err := m.GetSetting(key)
	if err != nil {
		return eos.Asset{}, err
	}
	value, err := setting.GetAsAsset(0)
	if err != nil {
		return eos.Asset{}, fmt.Errorf("failed getting setting: %v as asset, error: %v", value, err)
	}
	return value, nil
}

func (m *SettingsContract) SettingAsInt64(key string) (int64, error) {
	setting, err := m.GetSetting(key)
	if err != nil {
		return 0, err
	}
	value, err := setting.GetAsInt64(0)
	if err != nil {
		return 0, fmt.Errorf("failed getting setting: %v as int64, error: %v", value, err)
	}
	return value, nil
}

func (m *SettingsContract) SettingAsName(key string) (eos.Name, error) {
	setting, err := m.GetSetting(key)
	if err != nil {
		return "", err
	}
	value, err := setting.GetAsName(0)
	if err != nil {
		return "", fmt.Errorf("failed getting setting: %v as name, error: %v", value, err)
	}
	return value, nil
}

func (m *SettingsContract) SettingAsString(key string) (string, error) {
	setting, err := m.GetSetting(key)
	if err != nil {
		return "", err
	}
	value, err := setting.GetAsString(0)
	if err != nil {
		return "", fmt.Errorf("failed getting setting: %v as string, error: %v", value, err)
	}
	return value, nil
}

func (m *SettingsContract) setter(owner eos.AccountName,
	key string, flexValue *dto.FlexValue, action eos.ActionName) (*service.PushTransactionFullResp, error) {
	actionData := m.getSetterData(owner, key, flexValue)
	return m.ExecAction(string(owner), string(action), actionData)
}

func (m *SettingsContract) proposeSetter(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, owner eos.AccountName,
	key string, flexValue *dto.FlexValue, action eos.ActionName) (*service.ProposeResponse, error) {
	actionData := m.getSetterData(owner, key, flexValue)
	return m.ProposeAction(proposerName, requested, expireIn, string(owner), string(action), actionData)
}

func (m *SettingsContract) getSetterData(owner eos.AccountName,
	key string, flexValue *dto.FlexValue) *ModifySettingArgs {
	return &ModifySettingArgs{
		Setter: owner,
		Key:    key,
		Value:  flexValue,
	}
}

func (m *SettingsContract) SetSetting(owner eos.AccountName,
	key string, flexValue *dto.FlexValue) (*service.PushTransactionFullResp, error) {

	return m.setter(owner, key, flexValue, eos.ActN("setsetting"))
}

func (m *SettingsContract) ProposeSetSetting(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, owner eos.AccountName,
	key string, flexValue *dto.FlexValue) (*service.ProposeResponse, error) {

	return m.proposeSetter(proposerName, requested, expireIn, owner, key, flexValue, eos.ActN("setsetting"))
}

func (m *SettingsContract) AppendSetting(owner eos.AccountName,
	key string, flexValue *dto.FlexValue) (*service.PushTransactionFullResp, error) {

	return m.setter(owner, key, flexValue, eos.ActN("appndsetting"))
}

func (m *SettingsContract) ProposeAppendSetting(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, owner eos.AccountName,
	key string, flexValue *dto.FlexValue) (*service.ProposeResponse, error) {

	return m.proposeSetter(proposerName, requested, expireIn, owner, key, flexValue, eos.ActN("appndsetting"))
}

func (m *SettingsContract) ClipSetting(owner eos.AccountName,
	key string, flexValue *dto.FlexValue) (*service.PushTransactionFullResp, error) {

	return m.setter(owner, key, flexValue, eos.ActN("clipsetting"))
}

func (m *SettingsContract) ProposeClipSetting(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, owner eos.AccountName,
	key string, flexValue *dto.FlexValue) (*service.ProposeResponse, error) {

	return m.proposeSetter(proposerName, requested, expireIn, owner, key, flexValue, eos.ActN("clipsetting"))
}

func (m *SettingsContract) EraseSetting(owner eos.AccountName, key string) (*service.PushTransactionFullResp, error) {

	eraseArgs := &EraseSettingArgs{
		Setter: owner,
		Key:    key,
	}

	return m.ExecAction(owner, "erasesetting", eraseArgs)
}

func (m *SettingsContract) ProposeEraseSetting(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, owner eos.AccountName,
	key string) (*service.ProposeResponse, error) {

	eraseArgs := &EraseSettingArgs{
		Setter: owner,
		Key:    key,
	}
	return m.ProposeAction(proposerName, requested, expireIn, string(owner), "erasesetting", eraseArgs)
}

func (m *SettingsContract) GetSettings() ([]Setting, error) {
	return m.GetSettingsReq(nil)
}

func (m *SettingsContract) GetAllSettingsAsMap() ([]map[string]interface{}, error) {
	req := eos.GetTableRowsRequest{
		Table: "settings",
	}
	return m.GetAllTableRowsFromAsMap(req, "id", "0", func(keyValue string) (string, error) {
		f, _, err := new(big.Float).Parse(keyValue, 10)
		if err != nil {
			return "", err
		}
		num, _ := f.Uint64()
		return strconv.FormatUint(num, 10), nil
	})
}

func (m *SettingsContract) GetSetting(key string) (*Setting, error) {
	setting, err := m.FindSetting(key)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, fmt.Errorf("setting: %v does not exist", key)
	}
	return setting, nil
}

func (m *SettingsContract) FindSetting(key string) (*Setting, error) {
	settings, err := m.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed getting settings: %v, error: %v", key, err)
	}
	for _, setting := range settings {
		if setting.Key == key {
			return &setting, nil
		}
	}
	return nil, nil
}

func (m *SettingsContract) GetSettingsReq(req *eos.GetTableRowsRequest) ([]Setting, error) {

	var settings []Setting
	if req == nil {
		req = &eos.GetTableRowsRequest{}
	}
	req.Table = "settings"
	err := m.GetTableRows(*req, &settings)
	if err != nil {
		return nil, fmt.Errorf("get table rows %v", err)
	}
	return settings, nil
}

func (m *SettingsContract) BuildSetSettingActions(owner eos.AccountName, settings interface{}) ([]*eos.Action, error) {
	actions := make([]*eos.Action, 0)
	for _, value := range settings.([]interface{}) {
		setting, err := GetConfigSetting(value.(map[interface{}]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed getting config setting: %v, error: %v", value, err)
		}
		action, err := m.BuildAction("setsetting", owner, m.getSetterData(owner, setting.Key, setting.Values[0]))
		if err != nil {
			return nil, fmt.Errorf("failed building action: %v, error: %v", setting, err)
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func (m *SettingsContract) SetupConfigSettings(owner eos.AccountName, settings interface{}) (*service.PushTransactionFullResp, error) {
	actions, err := m.BuildSetSettingActions(owner, settings)
	if err != nil {
		return nil, fmt.Errorf("failed building set setting actions: %v, error: %v", settings, err)
	}
	return m.ExecActions(actions...)
}

func (m *SettingsContract) SetupConfigSetting(owner eos.AccountName, configSetting map[interface{}]interface{}) error {

	setting, err := GetConfigSetting(configSetting)
	if err != nil {
		return err
	}
	_, err = m.SetSetting(owner, setting.Key, setting.Values[0])
	if err != nil {
		fmt.Println("Error in set setting: ", setting.Key, setting.Values[0])
		return err
	}

	return nil
}

func (m *SettingsContract) ProposeConfigSettings(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, owner eos.AccountName, settings interface{}) (*service.ProposeResponse, error) {
	actions, err := m.BuildSetSettingActions(owner, settings)
	if err != nil {
		return nil, fmt.Errorf("failed building set setting actions: %v, error: %v", settings, err)
	}
	return m.EOS.ProposeMultiSig(proposerName, requested, expireIn, actions...)
}

func (m *SettingsContract) ProposeConfigSetting(proposerName interface{}, requested []eos.PermissionLevel, expireIn time.Duration, owner eos.AccountName, configSetting map[interface{}]interface{}) error {

	setting, err := GetConfigSetting(configSetting)
	if err != nil {
		return err
	}
	_, err = m.ProposeSetSetting(proposerName, requested, expireIn, owner, setting.Key, setting.Values[0])
	if err != nil {
		return err
	}

	return nil
}

func GetConfigSetting(setting map[interface{}]interface{}) (*Setting, error) {

	fv, err := dto.ParseToFlexValue(setting["type"].(string), setting["value"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed parsing setting to flex value, error: %v", err)
	}
	return &Setting{
		Key: setting["key"].(string),
		Values: []*dto.FlexValue{
			fv,
		},
	}, nil
}
