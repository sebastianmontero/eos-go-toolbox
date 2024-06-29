package util

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/sebastianmontero/eos-go"
)

const (
	PercentageAdjustment = float64(10000000)
	HundredPercent       = uint32(10000000)
	charset              = "abcdefghijklmnopqrstuvwxyz" + "12345"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func MultiplyAsset(asset eos.Asset, times int64) eos.Asset {
	total := eos.Int64(int64(asset.Amount) * times)
	return eos.Asset{Amount: total, Symbol: asset.Symbol}
}

func MultiplyAssets(amount, value eos.Asset) eos.Asset {
	result := big.NewInt(0).Mul(big.NewInt(int64(amount.Amount)), big.NewInt(int64(value.Amount)))
	result = AdjustPrecision(result, amount.Precision+value.Precision, value.Precision)
	if !result.IsInt64() {
		panic("Division overflow")
	}
	return eos.Asset{Amount: eos.Int64(result.Int64()), Symbol: value.Symbol}
}

// func CalculatePercentage(amount interface{}, percentage uint32) eos.Asset {
// 	amnt, err := ToAsset(amount)
// 	if err != nil {
// 		panic(fmt.Sprintf("failed to calculate percentage, could not parse amount: %v", amount))
// 	}
// 	perctAmnt := float64(amnt.Amount) * float64((float64(percentage) / PercentageAdjustment))
// 	return eos.Asset{Amount: eos.Int64(perctAmnt), Symbol: amnt.Symbol}
// }

func CalculateAssetPercentage(amount interface{}, percentage uint32) eos.Asset {
	amnt, err := ToAsset(amount)
	if err != nil {
		panic(fmt.Sprintf("failed to calculate percentage, could not parse amount: %v", amount))
	}
	return eos.Asset{Amount: eos.Int64(CalculatePercentage(int64(amnt.Amount), percentage)), Symbol: amnt.Symbol}
}

func CalculatePercentage(amount int64, percentage uint32) int64 {
	percAdj := big.NewInt(int64(PercentageAdjustment))
	r := big.NewInt(0)
	r.Mul(big.NewInt(int64(amount)), big.NewInt(int64(percentage)))
	r.Div(r, percAdj)
	if !r.IsInt64() {
		panic("amount percentage overflows int64")
	}
	return r.Int64()
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

func IsNullTime(t time.Time) bool {
	return t.Unix() == 0
}

func IsNullTimePoint(t eos.TimePoint) bool {
	return t == eos.TimePoint(0)
}

func BlockTimestampToString(t eos.BlockTimestamp) string {
	return TimeToString(t.Time)
}

func TimePointToString(t eos.TimePoint) string {
	if IsNullTimePoint(t) {
		return ""
	}
	return t.String()
}

func TimeToString(t time.Time) string {
	if IsNullTime(t) {
		return ""
	}
	return t.Format("2006-01-02T15:04:05.000")
}

func UTCTimeToString(t time.Time) string {
	if IsNullTime(t) {
		return ""
	}
	return t.UTC().Format("2006-01-02T15:04:05.000")
}

func NowToString() string {
	return TimeToString(time.Now())
}

func UTCNowToString() string {
	return UTCTimeToString(time.Now())
}

func UTCShiftedTimeToString(d time.Duration) string {
	return UTCTimeToString(time.Now().Add(d))
}

func ShiftedTime(d time.Duration) time.Time {
	return time.Now().Add(d)
}

func ShiftedBlockTimestamp(d time.Duration) eos.BlockTimestamp {
	return eos.BlockTimestamp{Time: time.Now().Add(d)}
}

func ShiftedTimePoint(d time.Duration) eos.TimePoint {
	return eos.TimePoint(time.Now().Add(d).UnixMicro())
}

// ToTime Converts string time to time.Time
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
	case eos.Name:
		return eos.AN(v.String()), nil
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
	case eos.AccountName:
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

func ToChecksum256(value interface{}) (eos.Checksum256, error) {
	switch v := value.(type) {
	case string:
		b, err := hex.DecodeString(v)
		if err != nil {
			return eos.Checksum256{}, fmt.Errorf("failed to parse value: %v to Checksum256, error: %v", v, err)
		}
		return eos.Checksum256(b), nil
	case []byte:
		return eos.Checksum256(v), nil
	case eos.Checksum256:
		return v, nil
	default:
		return eos.Checksum256{}, fmt.Errorf("failed to convert: %v of type: %T to Checksum256", value, value)
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
	letters := "abcdefghijklmnopqrstuvwxyz"
	all := letters + "12345."
	return StringWithCharset(1, letters) + StringWithCharset(11, all)
}
