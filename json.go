package tools

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type JSONBuilder struct{}

func (b JSONBuilder) FromString(jsonStr string) (JSON, error) {
	if strings.TrimSpace(jsonStr) == "" {
		return nil, nil
	}
	j := new(JSON)
	if err := j.Scan([]byte(jsonStr)); err != nil {
		return JSON(jsonStr), err
	}
	return *j, nil
}

func (b JSONBuilder) NoErrString(jsonStr string) JSON {
	j, _ := b.FromString(jsonStr)
	return j
}

func (b JSONBuilder) FromObj(obj any) (JSON, error) {
	jsonstr, err := JsonString(obj)
	if err != nil {
		return nil, err
	}
	return JSON(jsonstr), nil
}

func (b JSONBuilder) NoErrObj(obj any) JSON {
	j, _ := b.FromObj(obj)
	return j
}

type JSON json.RawMessage

func (j JSON) IsNull() bool {
	return len(j) == 0
}

func (j *JSON) Scan(value interface{}) error {
	bs, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %x", value)
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bs, &result)
	if err == nil {
		buf := new(bytes.Buffer)
		_ = json.Compact(buf, result)
		*j = JSON(buf.Bytes())
	} else {
		*j = bs
	}
	return err
}

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSON) String() string {
	if j == nil || len(*j) == 0 {
		return ""
	}
	return string(*j)
}

func (j *JSON) Bytes() []byte {
	if j == nil || len(*j) == 0 {
		return nil
	}
	return []byte(*j)
}

func (j *JSON) Equal(o *JSON) bool {
	if j == nil && o == nil {
		return true
	}
	if j == nil || o == nil {
		return false
	}
	return bytes.Equal(*j, *o)
}

func (j *JSON) Clone() JSON {
	if j == nil {
		return nil
	}
	c := CopySlice([]byte(*j))
	return JSON(c)
}

func (j JSON) Format(f fmt.State, c rune) {
	s := j.String()
	_, _ = fmt.Fprint(f, s)
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	var result *json.RawMessage
	err := json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	if result == nil {
		*j = nil
	} else {
		buf := new(bytes.Buffer)
		_ = json.Compact(buf, *result)
		*j = buf.Bytes()
	}
	return nil
}

type NotNullJSON JSON

func (n NotNullJSON) IsNull() bool {
	return len(n) == 0
}

func (n *NotNullJSON) Scan(value interface{}) error {
	bs, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal NotNullJSON value: %x", value)
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bs, &result)
	if err == nil {
		buf := new(bytes.Buffer)
		_ = json.Compact(buf, result)
		*n = NotNullJSON(buf.Bytes())
	} else {
		*n = bs
	}
	return err
}

func (n NotNullJSON) Value() (driver.Value, error) {
	return n.Bytes(), nil
}

func (n *NotNullJSON) String() string {
	return string(n.Bytes())
}

func (n *NotNullJSON) Bytes() []byte {
	if n == nil || len(*n) == 0 {
		return []byte("{}")
	}
	return *n
}

func (n *NotNullJSON) Equal(o *NotNullJSON) bool {
	if n == nil && o == nil {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return bytes.Equal(*n, *o)
}

func (n *NotNullJSON) Clone() NotNullJSON {
	if n == nil || len(*n) == 0 {
		return nil
	}
	return CopySlice(*n)
}

func (n NotNullJSON) MarshalJSON() ([]byte, error) {
	if len(n) == 0 {
		return nil, errors.New(`NotNullJSON with null value`)
	}
	return n, nil
}

func (n *NotNullJSON) UnmarshalJSON(data []byte) error {
	var result *json.RawMessage
	err := json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New(`NotNullJSON could not be null`)
	} else {
		buf := new(bytes.Buffer)
		_ = json.Compact(buf, *result)
		bs := buf.Bytes()
		if len(bs) == 0 {
			return errors.New(`NotNullJSON could not be null`)
		}
		*n = bs
	}
	return nil
}

func (n NotNullJSON) ToJSON() JSON {
	return JSON(n)
}

type JSONArray[T comparable] []T

func NewJSONArray[T comparable](data JSON) (ja JSONArray[T], err error) {
	if data.IsNull() {
		return nil, nil
	}
	if err = json.Unmarshal(data, &ja); err != nil {
		return nil, err
	}
	return ja, nil
}

func (ja JSONArray[T]) ToArray() []T {
	return ja
}

func (ja JSONArray[T]) Index(v T) int {
	for i, value := range ja {
		if value == v {
			return i
		}
	}
	return -1
}

func (ja JSONArray[T]) ToJSON() (JSON, error) {
	if ja == nil {
		return nil, nil
	}
	return JSONBuilder{}.FromObj(ja)
}

func (ja JSONArray[T]) ToNotNullJSON() (NotNullJSON, error) {
	jaa := ja
	if ja == nil {
		jaa = []T{}
	}
	js, err := JSONBuilder{}.FromObj(jaa)
	return NotNullJSON(js), err
}

func (ja JSONArray[T]) MarshalJSON() ([]byte, error) {
	s, err := JsonString([]T(ja))
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (ja JSONArray[T]) Equal(jt JSONArray[T]) bool {
	return (KS[T])(ja).Equal((KS[T])(jt))
}
