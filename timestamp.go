package tools

import "time"

type Timestamp int64

const (
	msThreshold = 99999999999
)

func (t Timestamp) ToMilliSecond() Timestamp {
	if t.IsMilli() {
		return t
	}
	return t * 1000
}

func (t Timestamp) ToSecond() Timestamp {
	if t.IsMilli() {
		return t / 1000
	}
	return t
}

func (t Timestamp) IsMilli() bool {
	return int64(t) > msThreshold
}

func (t Timestamp) ToTime() Time {
	if t.IsMilli() {
		return Time(time.UnixMilli(int64(t)).Truncate(time.Millisecond))
	}
	return Time(time.UnixMilli(int64(t)).Truncate(time.Second))
}

func (t Timestamp) DiffSeconds(o Timestamp) int64 {
	t1 := t.ToSecond()
	t2 := o.ToSecond()
	if t1 >= t2 {
		return int64(t1) - int64(t2)
	}
	return int64(t2) - int64(t1)
}

func (t Timestamp) DiffMillis(o Timestamp) int64 {
	t1 := t.ToMilliSecond()
	t2 := o.ToMilliSecond()
	if t1 >= t2 {
		return int64(t1) - int64(t2)
	}
	return int64(t2) - int64(t1)
}
