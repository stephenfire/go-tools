package tools

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"hash"
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

func AllMapValueSlices[K comparable, T comparable](m map[K][]T, needDedup ...bool) []T {
	if len(m) == 0 {
		return nil
	}
	if VariadicParam(needDedup) {
		idset := make(KSet[T])
		for _, ids := range m {
			idset.Add(ids...)
		}
		return idset.Slice()
	} else {
		var idslice []T
		for _, ids := range m {
			idslice = append(idslice, ids...)
		}
		return idslice
	}
}

func Hash256(hasher hash.Hash, in ...[]byte) []byte {
	for _, n := range in {
		hasher.Write(n)
	}
	return hasher.Sum(nil)
}
