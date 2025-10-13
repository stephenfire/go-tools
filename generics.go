package tools

func VariadicParam[T any](params []T, defaultValue ...T) T {
	var defaultVal T
	if len(defaultValue) > 0 {
		defaultVal = defaultValue[0]
	}
	if len(params) == 0 {
		return defaultVal
	}
	return params[0]
}

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

func Abs[I ~int | ~int8 | ~int16 | ~int32 | ~int64](i I) I {
	if i < 0 {
		return -i
	}
	return i
}

func BatchCall[T any](op func([]T) bool, batchSize int, ts ...T) {
	if batchSize <= 0 {
		batchSize = 100
	}
	if len(ts) <= batchSize {
		op(ts)
		return
	}
	for start := 0; start < len(ts); start += batchSize {
		if !op(ts[start:min(start+batchSize, len(ts))]) {
			return
		}
	}
}
