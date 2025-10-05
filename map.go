package tools

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

type Exister[K comparable] interface {
	IsExist(k K) bool
}

type KMap[K comparable, V any] map[K]V

func (km KMap[K, V]) Map() map[K]V {
	return km
}

func (km KMap[K, V]) Put(key K, value V) KMap[K, V] {
	m := km
	if m == nil {
		m = make(KMap[K, V])
	}
	m[key] = value
	return m
}

func (km KMap[K, V]) Merge(kmo KMap[K, V]) KMap[K, V] {
	if len(kmo) == 0 {
		return km
	}
	if len(km) == 0 {
		return kmo
	}
	for k, v := range kmo {
		km[k] = v
	}
	return km
}

func (km KMap[K, V]) Puts(it iter.Seq2[K, V]) KMap[K, V] {
	m := km
	for k, v := range it {
		if m == nil {
			m = make(map[K]V)
		}
		m[k] = v
	}
	return m
}

func (km KMap[K, V]) Delete(ks ...K) {
	if len(km) == 0 {
		return
	}
	for _, k := range ks {
		delete(km, k)
	}
}

func (km KMap[K, V]) IsExist(key K) bool {
	_, ok := km[key]
	return ok
}

func (km KMap[K, V]) Get(k K) (v V, exist bool) {
	if len(km) == 0 {
		return
	}
	v, exist = km[k]
	return
}

func (km KMap[K, V]) Keys(filters ...func(k K, v V) bool) []K {
	return slices.Collect(km.KeySeq(filters...))
}

func (km KMap[K, V]) Values(filters ...func(k K, v V) bool) []V {
	return slices.Collect(km.ValuesSeq(filters...))
}

func (km KMap[K, V]) KeySeq(filters ...func(k K, v V) bool) iter.Seq[K] {
	return func(yield func(K) bool) {
		var filter func(k K, v V) bool
		if len(filters) > 0 && filters[0] != nil {
			filter = filters[0]
		}
		for k, v := range km {
			if filter == nil || filter(k, v) {
				if !yield(k) {
					return
				}
			}
		}
	}
}

func (km KMap[K, V]) ValuesSeq(filters ...func(k K, v V) bool) iter.Seq[V] {
	return func(yield func(V) bool) {
		var filter func(k K, v V) bool
		if len(filters) > 0 && filters[0] != nil {
			filter = filters[0]
		}
		for k, v := range km {
			if filter == nil || filter(k, v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func (km KMap[K, V]) List(keys ...K) []V {
	ret := make([]V, len(keys))
	if len(km) == 0 {
		return ret
	}
	for i, k := range keys {
		ret[i] = km[k]
	}
	return ret
}

func (km KMap[K, V]) ExistingList(keys ...K) []V {
	var ret []V
	if len(km) == 0 || len(keys) == 0 {
		return ret
	}
	for _, k := range keys {
		v, ok := km[k]
		if ok {
			ret = append(ret, v)
		}
	}
	return ret
}

// SubMap generate a new map with all keys from ks, even if there's no corresponding value in km.
func (km KMap[K, V]) SubMap(ks ...K) KMap[K, V] {
	if len(ks) == 0 {
		return km
	}
	m := make(KMap[K, V], len(ks))
	for _, k := range ks {
		m[k] = km[k]
	}
	return m
}

func (km KMap[K, V]) RangeSubMap(batchSize int, ranger func(m KMap[K, V]) bool) {
	if len(km) == 0 {
		return
	}
	if len(km) <= batchSize {
		ranger(km)
		return
	}
	sub := make(KMap[K, V], batchSize)
	for k, v := range km {
		sub[k] = v
		if len(sub) >= batchSize {
			ranger(sub)
			sub = make(KMap[K, V], batchSize)
		}
	}
	if len(sub) > 0 {
		ranger(sub)
	}
}

type KKMap[K1 comparable, K2 comparable, V any] map[K1]map[K2]V

func (kkm KKMap[K1, K2, V]) Put(k1 K1, k2 K2, v V) KKMap[K1, K2, V] {
	kkmm := kkm
	if kkmm == nil {
		kkmm = make(KKMap[K1, K2, V])
	}
	kmm, exist := kkmm[k1]
	if !exist {
		kmm = make(map[K2]V)
		kkmm[k1] = kmm
	}
	kmm[k2] = v
	return kkmm
}

func (kkm KKMap[K1, K2, V]) PutKey(k1 K1, km2 map[K2]V) KKMap[K1, K2, V] {
	kkmm := kkm
	if kkmm == nil {
		kkmm = make(KKMap[K1, K2, V])
	}
	kkmm[k1] = km2
	return kkmm
}

func (kkm KKMap[K1, K2, V]) Delete(k1 K1, k2 K2) {
	km, exist := kkm[k1]
	if !exist {
		return
	}
	delete(km, k2)
	if len(km) == 0 {
		delete(kkm, k1)
	}
}

func (kkm KKMap[K1, K2, V]) DeleteKey(k1 K1) {
	if len(kkm) == 0 {
		return
	}
	delete(kkm, k1)
}

func (kkm KKMap[K1, K2, V]) Get(k1 K1, k2 K2) (v V, exist bool) {
	if len(kkm) == 0 {
		return
	}
	if km, ok := kkm[k1]; !ok || len(km) == 0 {
		return
	} else {
		v, exist = km[k2]
		return
	}
}

func (kkm KKMap[K1, K2, V]) IsExist(k1 K1, k2 K2) (exist bool) {
	if len(kkm) == 0 {
		return false
	}
	if km, ok := kkm[k1]; !ok || len(km) == 0 {
		return false
	} else {
		_, exist = km[k2]
		return exist
	}
}

func (kkm KKMap[K1, K2, V]) IsExistKey(k1 K1) (exist bool) {
	_, exist = kkm[k1]
	return
}

func (kkm KKMap[K1, K2, V]) Range(handler func(k1 K1, k2 K2, v V) (goon bool)) {
	if len(kkm) == 0 {
		return
	}
	for k1, km := range kkm {
		for k2, v := range km {
			if !handler(k1, k2, v) {
				return
			}
		}
	}
}

type KSet[K comparable] map[K]struct{}

func NewKSet[K comparable](ks ...K) KSet[K] {
	km := make(KSet[K])
	return km.Add(ks...)
}

func (km KSet[K]) Append(ks ...K) KSet[K] {
	if len(ks) == 0 {
		return km
	}
	m := km
	if m == nil {
		m = make(map[K]struct{})
	}
	for _, k := range ks {
		m[k] = struct{}{}
	}
	return m
}

func (km KSet[K]) Add(ks ...K) KSet[K] {
	for _, k := range ks {
		km[k] = struct{}{}
	}
	return km
}

func (km KSet[K]) Adds(it iter.Seq[K]) KSet[K] {
	for k := range it {
		km[k] = struct{}{}
	}
	return km
}

func (km KSet[K]) AppendSet(s KSet[K]) KSet[K] {
	if len(s) == 0 {
		return km
	}
	if len(km) == 0 {
		return s
	}
	for k := range s {
		km[k] = struct{}{}
	}
	return km
}

func (km KSet[K]) Appends(it iter.Seq[K]) KSet[K] {
	m := km
	for k := range it {
		if m == nil {
			m = make(map[K]struct{})
		}
		m[k] = struct{}{}
	}
	return m
}

func (km KSet[K]) CAS(k K) (set KSet[K], changed bool) {
	if _, exist := km[k]; exist {
		return km, false
	}
	return km.Append(k), true
}

func (km KSet[K]) Delete(ks ...K) KSet[K] {
	for _, k := range ks {
		delete(km, k)
	}
	return km
}

func (km KSet[K]) IsExist(k K) bool {
	_, exist := km[k]
	return exist
}

// In return a sequence that in both input and current set
func (km KSet[K]) In(input iter.Seq[K]) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range input {
			if km.IsExist(k) {
				if !yield(k) {
					return
				}
			}
		}
	}
}

func (km KSet[K]) Keys() iter.Seq[K] {
	return maps.Keys(km)
}

func (km KSet[K]) Slice(emptyNil ...bool) []K {
	if km == nil {
		return nil
	}
	if len(km) == 0 {
		if len(emptyNil) > 0 && emptyNil[0] {
			return nil
		}
		return []K{}
	}
	ks := make([]K, 0, len(km))
	for k := range km {
		ks = append(ks, k)
	}
	return ks
}

func (km KSet[K]) Clone() KSet[K] {
	if km == nil {
		return nil
	}
	ret := make(KSet[K], len(km))
	for k := range km {
		ret[k] = struct{}{}
	}
	return ret
}

func (km KSet[K]) Equal(o KSet[K]) bool {
	if km == nil && o == nil {
		return true
	}
	if km == nil || o == nil {
		return false
	}
	if len(km) != len(o) {
		return false
	}
	for k := range km {
		if _, ok := o[k]; !ok {
			return false
		}
	}
	return true
}

func (km KSet[K]) ExsitingList(ks ...K) []K {
	var ret []K
	if len(km) == 0 || len(ks) == 0 {
		return ret
	}
	for _, k := range ks {
		if _, ok := km[k]; ok {
			ret = append(ret, k)
		}
	}
	return ret
}

func (km KSet[K]) RangeSubSet(batchSize int, ranger func(s KSet[K]) bool) {
	KMap[K, struct{}](km).RangeSubMap(batchSize, func(m KMap[K, struct{}]) bool {
		s := make(KSet[K])
		s.Appends(maps.Keys(m))
		return ranger(s)
	})
}

// OrderMap 一个按插入顺序遍历的map，也可以通过给出的排序器顺序遍历。线程不安全。
type OrderMap[K comparable, V any] struct {
	m map[K]V
	e map[K]*ListElement[K]
	l *List[K]
}

func NewOrderMap[K comparable, V any]() *OrderMap[K, V] {
	return &OrderMap[K, V]{
		m: make(map[K]V),
		e: make(map[K]*ListElement[K]),
		l: NewList[K](),
	}
}

func (m *OrderMap[K, V]) Put(k K, v V) *OrderMap[K, V] {
	_, exist := m.m[k]
	if exist {
		return m
	}
	m.m[k] = v
	m.e[k] = m.l.PushBack(k)
	return m
}

func (m *OrderMap[K, V]) Get(k K) (V, bool) {
	v, exist := m.m[k]
	return v, exist
}

func (m *OrderMap[K, V]) Del(k K) {
	elem, exist := m.e[k]
	if exist {
		m.l.Remove(elem)
		delete(m.m, k)
		delete(m.e, k)
	}
}

func (m *OrderMap[K, V]) Len() int          { return len(m.m) }
func (m *OrderMap[K, V]) Keys() iter.Seq[K] { return m.l.All() }

func (m *OrderMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k := range m.l.All() {
			v := m.m[k]
			if !yield(k, v) {
				return
			}
		}
	}
}

func (m *OrderMap[K, V]) selfCheck() error {
	if len(m.m) != len(m.e) || len(m.m) != m.l.Len() {
		return fmt.Errorf("length mismatch: m:%d e:%d l:%d", len(m.m), len(m.e), m.l.Len())
	}
	for k := range m.l.All() {
		if _, exist := m.m[k]; !exist {
			return fmt.Errorf("element missing in m at k:%v", k)
		}
		if _, exist := m.e[k]; !exist {
			return fmt.Errorf("element missing in e at k:%v", k)
		}
	}
	return nil
}
