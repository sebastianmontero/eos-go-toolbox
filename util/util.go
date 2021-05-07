package util

import (
	"fmt"
	"time"

	"github.com/eoscanada/eos-go"
)

func MultiplyAsset(asset eos.Asset, times int64) eos.Asset {
	total := eos.Int64(int64(asset.Amount) * times)
	return eos.Asset{Amount: total, Symbol: asset.Symbol}
}

//ToTime Converts string time to time.Time
func ToTime(strTime string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05", strTime)
	if err != nil {
		return time.Parse("2006-01-02T15:04:05.000", strTime)
	}
	return t, nil
}

func ToAccountName(value interface{}) (eos.AccountName, error) {
	switch v := value.(type) {
	case string:
		return eos.AN(v), nil
	case eos.AccountName:
		return v, nil
	default:
		return "", fmt.Errorf("failed to convert: %v of type: %T to AccountName", value, value)
	}
}

func ToActionName(value interface{}) (eos.ActionName, error) {
	switch v := value.(type) {
	case string:
		return eos.ActN(v), nil
	case eos.ActionName:
		return v, nil
	default:
		return "", fmt.Errorf("failed to convert: %v of type: %T to ActionName", value, value)
	}
}

func ToName(value interface{}) (eos.Name, error) {
	switch v := value.(type) {
	case string:
		return eos.Name(v), nil
	case eos.Name:
		return v, nil
	default:
		return "", fmt.Errorf("failed to convert: %v of type: %T to Name", value, value)
	}
}

func ToPermissionName(value interface{}) (eos.PermissionName, error) {
	switch v := value.(type) {
	case string:
		return eos.PN(v), nil
	case eos.PermissionName:
		return v, nil
	default:
		return "", fmt.Errorf("failed to convert: %v of type: %T to PermissionName", value, value)
	}
}

func ToAsset(value interface{}) (eos.Asset, error) {
	switch v := value.(type) {
	case string:
		asset, err := eos.NewAssetFromString(v)
		if err != nil {
			return eos.Asset{}, fmt.Errorf("failed to parse value: %v to Asset, error: %v", v, err)
		}
		return asset, nil
	case eos.Asset:
		return v, nil
	default:
		return eos.Asset{}, fmt.Errorf("failed to convert: %v of type: %T to Asset", value, value)
	}
}

func ToSymbol(value interface{}) (eos.Symbol, error) {
	switch v := value.(type) {
	case string:
		symbol, err := eos.StringToSymbol(v)
		if err != nil {
			return eos.Symbol{}, fmt.Errorf("failed to parse value: %v to Symbol, error: %v", v, err)
		}
		return symbol, nil
	case eos.Symbol:
		return v, nil
	default:
		return eos.Symbol{}, fmt.Errorf("failed to convert: %v of type: %T to Symbol", value, value)
	}
}

func ToPermissionLevel(permissionLevel interface{}) (eos.PermissionLevel, error) {
	switch v := permissionLevel.(type) {
	case eos.PermissionLevel:
		return v, nil
	case eos.AccountName, eos.Name, string:
		pl, err := eos.NewPermissionLevel(fmt.Sprintf("%v", v))
		if err != nil {
			return pl, fmt.Errorf("unable to parse permission level: %v, error: %v", v, err)
		}
		return pl, nil
	default:
		return eos.PermissionLevel{}, fmt.Errorf("unable to parse permission level: %v, of type: %t", v, v)
	}
}
