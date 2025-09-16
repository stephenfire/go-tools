package tools

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type NullString sql.Null[string]

func (n *NullString) Scan(value any) error {
	return (*sql.Null[string])(n).Scan(value)
}

func (n NullString) Value() (driver.Value, error) {
	return sql.Null[string](n).Value()
}

func (n NullString) IsNull() bool { return !n.Valid }

func (n NullString) String() string {
	if n.Valid {
		return n.V
	}
	return ""
}

func (n NullString) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte("null"), nil
	}
	return json.Marshal(n.V)
}

func (n *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == nil {
		n.Valid, n.V = false, ""
	} else {
		n.Valid, n.V = true, *s
	}
	return nil
}
