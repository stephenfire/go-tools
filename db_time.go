package tools

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	ErrNilSource = errors.New("tools: nil source value")
	ErrNilValue  = errors.New("tools: nil value")
)

// Time mysql timestamp precision to seconds
// using Now(), NewTime(), NewUnixTime() or ZeroTime() to create Time which is without monotonic
// clock reading (by truncating millisecond)
type Time time.Time

const (
	TimeTruncater      = time.Millisecond
	DefaultTimeFormat  = "2006-01-02 15:04:05 MST"
	TimeFormatWithMS   = "2006-01-02 15:04:05.000 MST"
	DefaultParseLayout = time.DateTime

	DateTruncater     = 24 * time.Hour
	DefaultDateFormat = time.DateOnly
)

func NewFullDate(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Time {
	return Time(time.Date(year, month, day, hour, min, sec, nsec, loc))
}

func NewDate(year int, month time.Month, day, hour, min, sec int) Time {
	return NewFullDate(year, month, day, hour, min, sec, 0, time.Local)
}

func NewTime(t time.Time) Time {
	return Time(t.Truncate(TimeTruncater))
}

func NewUnixTime(ts int64) Time {
	if ts > 99999999999 || ts < (-99999999999) {
		return Time(time.UnixMilli(ts).Truncate(TimeTruncater))
	}
	return Time(time.Unix(ts, 0).Truncate(TimeTruncater))
}

func ParseTime(layout, value string) (Time, error) {
	t, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		return Time{}, err
	}
	return Time(t.Truncate(TimeTruncater)), nil
}

// ZeroTime according to time.Time document, which is Jan 1 year 1.
func ZeroTime() Time {
	return Time(time.Time{})
}

func Now() Time {
	return Time(time.Now().Truncate(TimeTruncater))
}

func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t Time) After(d time.Duration) Time {
	return Time(time.Time(t).Add(d))
}

func (t Time) Before(d time.Duration) Time {
	return Time(time.Time(t).Add(-d))
}

func (t Time) NextHour() Time {
	return Time(time.Time(t).Truncate(time.Hour).Add(time.Hour))
}

func (t Time) CurrentHour() Time {
	return Time(time.Time(t).Truncate(time.Hour))
}

func (t Time) CurrentRound(interval time.Duration) Time {
	return Time(time.Time(t).Truncate(interval))
}

func (t Time) NextRound(interval time.Duration) Time {
	return Time(time.Time(t).Add(interval)).CurrentRound(interval)
}

func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

func (t Time) UnixMilli() int64 {
	return time.Time(t).UnixMilli()
}

// NNUnixMilli not negetive unix milli
func (t Time) NNUnixMilli() int64 {
	un := t.UnixMilli()
	if un < 0 {
		return 0
	}
	return un
}

func (t Time) UnixString() string {
	return strconv.FormatInt(t.Unix(), 10)
}

func (t Time) UnixMilliString() string {
	return strconv.FormatInt(time.Time(t).UnixMilli(), 10)
}

func (t Time) HourAfter(d time.Duration) Time {
	return Time(time.Time(t).Add(d)).CurrentHour()
}

func (t Time) FromUnix(seconds int64) Time {
	return NewTime(time.Unix(seconds, 0))
}

func (t Time) Compare(o Time) int {
	return time.Time(t).Compare(time.Time(o))
}

func (t Time) String() string {
	return time.Time(t).Format(DefaultTimeFormat)
}

func (t Time) UTCString() string {
	return time.Time(t).UTC().Format(TimeFormatWithMS)
}

func (t Time) DetailString() string {
	if t.IsZero() {
		return "<nil>"
	}
	return fmt.Sprintf("%d(%s)", t.UnixMilli(), t.String())
}

func (t Time) Time() time.Time {
	return time.Time(t)
}

func (t Time) Sub(o Time) time.Duration {
	return time.Time(t).Sub(time.Time(o))
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal((time.Time)(t).UnixMilli())
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var i *int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	if i == nil {
		return ErrNilSource
	} else {
		*t = NewUnixTime(*i)
	}
	return nil
}

func (t Time) Equal(o Time) bool {
	return t.UnixMilli() == o.UnixMilli()
}

func (t *Time) Scan(value any) error {
	if value == nil {
		return ErrNilSource
	} else {
		if v, ok := value.(time.Time); ok {
			if t == nil {
				return ErrNilValue
			}
			*t = Time(v)
			return nil
		} else {
			return errors.New("tools: Time scan source was not time.Time")
		}
	}
}

func (t Time) Value() (driver.Value, error) {
	return time.Time(t).Truncate(TimeTruncater), nil
}

func (t Time) ToNullTime() NullTime {
	return NullTime{Time: t.Time(), Valid: true}
}

func (t Time) Year() int         { return time.Time(t).Year() }
func (t Time) Month() time.Month { return time.Time(t).Month() }
func (t Time) Day() int          { return time.Time(t).Day() }

func (t Time) ToDate() Date {
	// 不使用Time.Truncate方法，避免因时区问题导致日期变化
	return Date(NewDate(t.Year(), t.Month(), t.Day(), 0, 0, 0))
}

type NullTime sql.NullTime

var _nulltime = NullTime{
	Time:  NewUnixTime(0).Time(),
	Valid: false,
}

func NewNullTime(t ...time.Time) NullTime {
	if len(t) <= 0 || t[0].IsZero() {
		return _nulltime
	} else {
		return NullTime{
			Time:  NewTime(t[0]).Time(),
			Valid: true,
		}
	}
}

func NewNullUnixTime(ut int64) NullTime {
	if ut <= 0 {
		return _nulltime
	} else {
		return NullTime{
			Time:  NewUnixTime(ut).Time(),
			Valid: true,
		}
	}
}

func ParseNullTime(layout, value string) NullTime {
	t, err := ParseTime(layout, value)
	if err != nil {
		return _nulltime
	}
	return t.ToNullTime()
}

func (n NullTime) IsNull() bool {
	return !n.Valid
}

func (n *NullTime) Scan(value any) error {
	return (*sql.NullTime)(n).Scan(value)
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time.Truncate(time.Second), nil
}

func (n NullTime) ToTime() Time {
	if !n.Valid {
		return ZeroTime()
	}
	return NewTime(n.Time)
}

func (n NullTime) Obj() time.Time {
	return n.Time
}

func (n NullTime) Compare(ti time.Time) int {
	if !n.Valid {
		return -1
	}
	return n.Time.Compare(ti)
}

func (n NullTime) Format(layout string) string {
	if !n.Valid {
		return "<nil>"
	}
	return n.Time.Format(layout)
}

func (n NullTime) String() string {
	if n.Valid {
		return Time(n.Time).String()
	}
	return ""
}

func (n NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time.UnixMilli())
	}
	return []byte("null"), nil
}

func (n *NullTime) UnmarshalJSON(data []byte) error {
	var i *int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	if i == nil {
		*n = _nulltime
	} else {
		*n = NewNullUnixTime(*i)
	}
	return nil
}

func (n NullTime) Equal(o NullTime) bool {
	if n.Valid != o.Valid {
		return false
	}
	return n.Time.UnixMilli() == o.Time.UnixMilli()
}

type Date time.Time

func NewADate(year int, month time.Month, day int) Date {
	t := NewDate(year, month, day, 0, 0, 0)
	d := t.ToDate()
	return d
}

func (d Date) String() string {
	return time.Time(d).Format(DefaultDateFormat)
}

func (d Date) Formalize() Date { return Time(d).ToDate() }

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := ParseTime(DefaultDateFormat, s)
	if err != nil {
		return err
	}
	*d = t.ToDate()
	return nil
}

func (d *Date) Scan(value any) error {
	if value == nil {
		return ErrNilSource
	} else {
		if v, ok := value.(time.Time); ok {
			if d == nil {
				return ErrNilValue
			}
			*d = Date(v)
			return nil
		} else {
			return errors.New("tools: Date scan source was not time.Time")
		}
	}
}

func (d Date) Value() (driver.Value, error) {
	return time.Time(d).Truncate(DateTruncater), nil
}
