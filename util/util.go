package util

import (
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
