package tools

import (
	"bytes"
	"encoding/json"
)

func CopySlice[T any](src []T) []T {
	if src == nil {
		return nil
	}
	ret := make([]T, len(src))
	if len(src) == 0 {
		return ret
	}
	copy(ret, src)
	return ret
}

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

func SsToTs[S, T any](transmitter func(S) T, ss ...S) []T {
	if ss == nil {
		return nil
	}
	r := make([]T, len(ss))
	for i, bs := range ss {
		r[i] = transmitter(bs)
	}
	return r
}

func TsToSs[T, S any](transmitter func(T) (S, bool), ts ...T) []S {
	if ts == nil {
		return nil
	}
	r := make([]S, 0, len(ts))
	for _, t := range ts {
		s, ok := transmitter(t)
		if ok {
			r = append(r, s)
		}
	}
	return r
}

func IF[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}
