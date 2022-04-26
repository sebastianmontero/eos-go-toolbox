package util

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/eoscanada/eos-go"
)

const (
	PercentageAdjustment = float64(10000000)
	charset              = "abcdefghijklmnopqrstuvwxyz" + "12345"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func MultiplyAsset(asset eos.Asset, times int64) eos.Asset {
	total := eos.Int64(int64(asset.Amount) * times)
	return eos.Asset{Amount: total, Symbol: asset.Symbol}
}

func CalculatePercentage(amount interface{}, percentage uint32) eos.Asset {
	amnt, err := ToAsset(amount)
	if err != nil {
		panic(fmt.Sprintf("failed to calculate percentage, could not parse amount: %v", amount))
	}
	perctAmnt := float64(amnt.Amount) * float64((float64(percentage) / PercentageAdjustment))
	return eos.Asset{Amount: eos.Int64(perctAmnt), Symbol: amnt.Symbol}
}

func DivideAssets(dividend, divisor eos.Asset) eos.Asset {
	dividendAdj := AdjustPrecision(big.NewInt(int64(dividend.Amount)), dividend.Precision, dividend.Precision+divisor.Precision)
	result := big.NewInt(0).Div(dividendAdj, big.NewInt(int64(divisor.Amount)))
	result = AdjustPrecision(result, dividend.Precision, divisor.Precision)
	if !result.IsInt64() {
		panic("Division overflow")
	}
	return eos.Asset{Amount: eos.Int64(result.Int64()), Symbol: divisor.Symbol}
}

func DivideAsset(asset eos.Asset, divisor uint64) eos.Asset {
	if divisor == 0 {
		panic("can not divide by zero")
	}
	return eos.Asset{Amount: asset.Amount / eos.Int64(divisor), Symbol: asset.Symbol}
}

func AdjustPrecision(amount *big.Int, precision, newPrecision uint8) *big.Int {
	diff := int(newPrecision) - int(precision)
	multiplier := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(math.Abs(float64(diff)))), nil)
	if diff > 0 {
		return amount.Mul(amount, multiplier)
	} else if diff < 0 {
		return amount.Div(amount, multiplier)
	} else {
		return amount
	}
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

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandAccountName() string {
	return StringWithCharset(12, charset)
}
