package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// EpochTime is a time.Time that implements json Marshaler and Unmarshaler interfaces.
// It is marshaled to integer type (UNIX epoch seconds) and unmarshaled from it.
//
// EpochTime does not implement the sql.Scanner interface so you have to cast it first to time.Time.
//
// 0 second is considered special for EpochTime. It is considered the zero value of time, so creating empty
// EpochTime will output 0 second. You can basically treat 0 or negative seconds as invalid, as they are a very old
// point in time.
type EpochTime time.Time

func (e *EpochTime) UnmarshalJSON(bytes []byte) error {
	if len(bytes) == 0 {
		*e = EpochTime{}
		return nil
	}

	var t int64 = 0
	err := json.Unmarshal(bytes, &t)
	if err != nil {
		return err
	}

	if t == 0 {
		*e = EpochTime{}
		return nil
	}

	*e = EpochTime(time.Unix(t, 0))
	return nil
}

func (e EpochTime) MarshalJSON() ([]byte, error) {
	var emptyEpochTime EpochTime
	if e == emptyEpochTime {
		return json.Marshal(0)
	}
	t := time.Time(e).Unix()
	return json.Marshal(t)
}

// Iso8601Date is a string representing a date in ISO 8601 format (YYYY-MM-DD).
type Iso8601Date string

func (i *Iso8601Date) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		t, err := time.ParseInLocation("2006-01-02", v, time.UTC)
		if err != nil {
			return err
		}
		*i = Iso8601Date(t.Format("2006-01-02"))
	case time.Time:
		*i = Iso8601Date(v.Format("2006-01-02"))
	default:
		return errors.New("unknown type, can only be a string or time.Time")
	}
	return nil
}

func (i Iso8601Date) Value() (driver.Value, error) {
	return string(i), nil
}

func (i *Iso8601Date) UnmarshalJSON(bytes []byte) error {
	s := ""
	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	*i = Iso8601Date(s)
	return nil
}

func (i Iso8601Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(i))
}

// Time parses the Iso8601Date to time.Time.
// It will only return error if Iso8601Date is not constructed through available methods correctly.
func (i Iso8601Date) Time() (time.Time, error) {
	return time.Parse("2006-01-02", string(i))
}

func ParseIso8601Date(s string) (Iso8601Date, error) {
	_, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		return "", err
	}
	return Iso8601Date(s), nil
}
