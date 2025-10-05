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
	km.RangeSubMap(17, func(m KMap[int, int]) bool {
		tm = tm.Merge(m)
		t.Logf("len(tm)=%d", len(tm))
		return true
	})

	if len(tm) != size {
		t.Fatal(fmt.Errorf("len(tm)!=size"))
	}

	if !maps.Equal(tm, km) {
		t.Fatal(fmt.Errorf("tm!=km"))
	}
}
