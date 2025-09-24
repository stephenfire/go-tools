package tools

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
)

func JsonString(m any) (string, error) {
	if m == nil {
		return "", nil
	}
	bs, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = json.Compact(buf, bs); err != nil {
		return string(bs), nil
	} else {
		return buf.String(), nil
	}
}

func MustJsonString(m any) string {
	a, _ := JsonString(m)
	return a
}

func JsonPrettyString(m any) (string, error) {
	if m == nil {
		return "", nil
	}
	bs, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = json.Indent(buf, bs, "", "  "); err != nil {
		return string(bs), nil
	} else {
		return buf.String(), nil
	}
}

func MustJsonPrettyString(m any) string {
	a, _ := JsonPrettyString(m)
	return a
}

func RandomBytes(length int) []byte {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return b
}
