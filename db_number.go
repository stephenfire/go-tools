package tools

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type NullInt32 sql.NullInt32

// LE0NullInt32 Less or Equal 0 is Null, or not Null
func LE0NullInt32(i int32) NullInt32 {
	if i <= 0 {
		return NullInt32(sql.NullInt32{
			Int32: 0,
			Valid: false,
		})
	}
	return NullInt32(sql.NullInt32{
		Int32: i,
		Valid: true,
	})
}

func (i NullInt32) IsNull() bool                 { return !i.Valid }
func (i NullInt32) IsAvailable() bool            { return i.Valid && i.Int32 > 0 }
func (i *NullInt32) Scan(value any) error        { return (*sql.NullInt32)(i).Scan(value) }
func (i NullInt32) Value() (driver.Value, error) { return sql.NullInt32(i).Value() }

func (i NullInt32) Int() int32 {
	if i.Valid {
		return i.Int32
	}
	return 0
}

func (i NullInt32) MarshalJSON() ([]byte, error) {
	if i.Valid {
		return json.Marshal(i.Int())
	}
	return []byte("null"), nil
}

func (i *NullInt32) UnmarshalJSON(input []byte) error {
	var v *int32
	iv := &v
	if err := json.Unmarshal(input, iv); err != nil {
		return err
	}
	if v == nil {
		i.Int32 = 0
		i.Valid = false
	} else {
		i.Int32 = *v
		i.Valid = true
	}
	return nil
}

func (i NullInt32) String() string {
	if i.Valid {
		return ID(i.Int32).String()
	}
	return ""
}

type NullInt64 sql.NullInt64

func NewNullInt64(i int64, valid bool) NullInt64 {
	return NullInt64{
		Int64: i,
		Valid: valid,
	}
}

// LE0NullInt64 Less or Equal to 0 is Null, or not Null
func LE0NullInt64(i int64) NullInt64 {
	if i <= 0 {
		return NullInt64(sql.NullInt64{
			Int64: 0,
			Valid: false,
		})
	}
	return NullInt64(sql.NullInt64{
		Int64: i,
		Valid: true,
	})
}

func NewIDValuer(ids ...int64) []driver.Valuer {
	return TsToSs(func(t int64) (driver.Valuer, bool) {
		if ID(t).IsValid() {
			return LE0NullInt64(t), true
		}
		return nil, false
	}, ids...)
}

func (i NullInt64) IsNull() bool                 { return !i.Valid }
func (i NullInt64) IsAvailable() bool            { return i.Valid && i.Int64 > 0 }
func (i *NullInt64) Scan(value any) error        { return (*sql.NullInt64)(i).Scan(value) }
func (i NullInt64) Value() (driver.Value, error) { return sql.NullInt64(i).Value() }

func (i *NullInt64) SetValue(val int64) {
	i.Int64 = val
	i.Valid = true
}

func (i *NullInt64) Clear() {
	i.Valid = false
	i.Int64 = 0
}

func (i NullInt64) Int() int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}

func (i NullInt64) MarshalJSON() ([]byte, error) {
	if i.Valid {
		return json.Marshal(i.Int())
	}
	return []byte("null"), nil
}

func (i *NullInt64) UnmarshalJSON(input []byte) error {
	var v *int64
	iv := &v
	if err := json.Unmarshal(input, iv); err != nil {
		return err
	}
	if v == nil {
		i.Int64 = 0
		i.Valid = false
	} else {
		i.Int64 = *v
		i.Valid = true
	}
	return nil
}

func (i NullInt64) String() string {
	if i.Valid {
		return ID(i.Int64).String()
	}
	return ""
}
