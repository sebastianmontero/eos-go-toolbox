package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

var MicrosecondsPerHr int64 = 60 * 60 * 1000000
var NanosecondsPerMicroSecond int64 = 1000

type Microseconds struct {
	Microseconds string `json:"_count"`
}

func NewMicroseconds(hrs int64) *Microseconds {
	return &Microseconds{
		Microseconds: strconv.FormatInt(hrs*MicrosecondsPerHr, 10),
	}
}

func (m *Microseconds) Hrs() int64 {
	ms, _ := strconv.ParseInt(m.Microseconds, 10, 64)
	return ms / MicrosecondsPerHr
}

func (m *Microseconds) UnmarshalJSON(b []byte) error {
	ms := make(map[string]interface{})
	if err := json.Unmarshal(b, &ms); err != nil {
		return err
	}
	if countI, ok := ms["_count"]; ok {
		var microseconds string
		switch count := countI.(type) {
		case float64:
			microseconds = strconv.FormatFloat(count, 'f', 0, 64)
		case string:
			microseconds = count
		default:
			return fmt.Errorf("Microseconds count of unknown type: %T", count)
		}
		*m = Microseconds{
			Microseconds: microseconds,
		}
	} else {
		return fmt.Errorf("error unmarshalling microseconds no '_count' property found: %v", ms)
	}
	return nil
}
func (m *Microseconds) NumMicroseconds() int64 {
	v, err := strconv.ParseInt(m.Microseconds, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed parsing microseconds: %v, error: %v", m.Microseconds, err))
	}
	return v
}

func (m *Microseconds) AsTimeDuration() time.Duration {
	v, err := strconv.ParseInt(m.Microseconds, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed parsing microseconds: %v, error: %v", m.Microseconds, err))
	}
	return time.Duration(v * NanosecondsPerMicroSecond)
}

func (m *Microseconds) String() string {
	return m.Microseconds
}

func (m *Microseconds) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"_count": m.Microseconds,
	}
}
