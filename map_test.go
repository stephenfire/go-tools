package tools

import (
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"testing"
)

func TestOrderMap(t *testing.T) {
	ks := []int64{9, 22, 90, 3, 1, 2, 6, 9, 90, 2}
	result := []int64{9, 22, 90, 3, 1, 2, 6}

	m := NewOrderMap[int64, int64]()
	for _, k := range ks {
		m.Put(k, k)
	}

	check := func(ks []int64) {
		if err := m.selfCheck(); err != nil {
			t.Fatal(err)
		}
		for k, v := range m.All() {
			if k != v {
				t.Fatal(fmt.Errorf("ks:%v k %v != %v", ks, k, v))
			}
		}
		if len(ks) != m.Len() {
			t.Fatal(fmt.Errorf("ks:%v len(ks) != m.Len()", ks))
		}
		for _, k := range ks {
			v, ok := m.Get(k)
			if !ok {
				t.Fatal(fmt.Errorf("ks:%v k %v not found", ks, k))
			}
			if k != v {
				t.Fatal(fmt.Errorf("ks:%v k %v != %v", k, k, v))
			}
		}
	}

	ts := slices.Collect(m.Keys())
	check(ts)
	if !slices.Equal(ts, result) {
		t.Fatal(fmt.Errorf("expected %v, got %v", result, ts))
	}

	for i := 0; i < len(ts)/2; i++ {
		r := rand.Intn(len(ts))
		k := ts[r]
		m.Del(k)
		ts = append(ts[:r], ts[r+1:]...)
		check(ts)
	}
}

func TestKMap_RangeSubMap(t *testing.T) {
	var km KMap[int, int]
	size := 1000
	for i := 0; i < size; i++ {
		j := rand.Intn(99999)
		km = km.Put(i, j)
	}
	if len(km) != size {
		t.Fatal(fmt.Errorf("len(km)!=size"))
	}

	var tm KMap[int, int]
	lastLen := 0
	batchSize := 17
	km.RangeSubMap(batchSize, func(m KMap[int, int]) bool {
		tm = tm.Merge(m)
		if len(m) > batchSize || len(m) == 0 {
			t.Fatal(fmt.Errorf("len(m)=%d, (batchSize=%d)", len(m), batchSize))
		}
		if (len(tm) - lastLen) != len(m) {
			t.Fatal(fmt.Errorf("duplicated key found, len(m)=%d, len(tm)=%d, lastLen=%d", len(m), len(tm), lastLen))
		}
		lastLen = len(tm)
		return true
	})

	if len(tm) != size {
		t.Fatal(fmt.Errorf("len(tm)!=size"))
	}

	if !maps.Equal(tm, km) {
		t.Fatal(fmt.Errorf("tm!=km"))
	}
}

func TestKSet_RangeSubSet(t *testing.T) {
	var ks KSet[int]
	size := 1000
	ks = ks.Appends(func(yield func(int) bool) {
		for i := 0; i < size; i++ {
			if !yield(i) {
				return
			}
		}
	})

	if len(ks) != size {
		t.Fatal(fmt.Errorf("len(ks)!=size"))
	}

	var ts KSet[int]
	lastLen := 0
	batchSize := 17
	ks.RangeSubSet(batchSize, func(s KSet[int]) bool {
		ts = ts.AppendSet(s)
		if len(s) > batchSize || len(s) == 0 {
			t.Fatal(fmt.Errorf("len(s)=%d, (batchSize=%d)", len(s), batchSize))
		}
		if (len(ts) - lastLen) != len(s) {
			t.Fatal(fmt.Errorf("duplicated key found, len(s)=%d, len(ts)=%d, lastLen=%d", len(s), len(ts), lastLen))
		}
		lastLen = len(ts)
		return true
	})

	if len(ts) != size {
		t.Fatal(fmt.Errorf("len(ts)!=size"))
	}
	if !maps.Equal(ts, ks) {
		t.Fatal(fmt.Errorf("ts!=tm"))
	}
}
