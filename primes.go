package tools

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type S string

func (s S) IsValid() bool {
	return s != ""
}

func (s S) Trim() S {
	return S(strings.TrimSpace(string(s)))
}

func (s S) ToLower() S {
	return S(strings.ToLower(string(s)))
}

func (s S) ToUpper() S {
	return S(strings.ToUpper(string(s)))
}

func (s S) Replace(source, target string) S {
	return S(strings.ReplaceAll(string(s), source, target))
}

func (s S) Formalize() S {
	return S(strings.ToLower(strings.TrimSpace(string(s))))
}

func (s S) CamelToSnake() S {
	if len(s) == 0 {
		return ""
	}
	var rs []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				if len(rs) > 0 && rs[len(rs)-1] != '_' {
					rs = append(rs, '_')
				}
			}
			rs = append(rs, unicode.ToLower(r))
		} else {
			rs = append(rs, r)
		}
	}
	return S(rs)
}

func (s S) Includes(ss ...string) bool {
	str := string(s)
	for _, one := range ss {
		if one == "" {
			continue
		}
		if !strings.Contains(str, one) {
			return false
		}
	}
	return true
}

func (s S) SplitBy(sep string) []string {
	return strings.Split(string(s), sep)
}

func (s S) RegexpSplit(expr string) ([]string, error) {
	reg, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return reg.Split(string(s), -1), nil
}

func (s S) CSVSplit() []string {
	reg := regexp.MustCompile("[,，、]")
	return reg.Split(string(s), -1)
}

func (s S) String() string {
	return string(s)
}

func (s S) Bytes() []byte {
	return []byte(s)
}

func (s S) FirstByte() byte {
	bs := []byte(s)
	if len(bs) > 0 {
		return bs[0]
	}
	return 0
}

func (s S) Int64() (int64, error) {
	ss := s.Trim().String()
	return strconv.ParseInt(ss, 10, 64)
}

func (s S) CSVLike() string {
	return "%," + string(s) + ",%"
}

func (s S) Like() string {
	return "%" + string(s) + "%"
}

func (s S) ID() (ID, error) {
	var i int64
	ss := s.Trim().Bytes()
	for _, sss := range ss {
		if sss >= '0' && sss <= '9' {
			i *= 10
			i += int64(sss - '0')
		} else {
			return 0, errors.New("tools: invalid id string")
		}
	}
	return ID(i), nil
}

func (s S) MustID() ID {
	i, err := s.ID()
	if err != nil {
		return 0
	}
	return i
}

func (s S) JSON() (JSON, error) {
	var j *JSON
	err := json.Unmarshal(s.Bytes(), &j)
	if err != nil {
		return nil, err
	}
	if j == nil {
		return nil, nil
	}
	return *j, nil
}

func (s S) SplitLines() (SS, error) {
	if len(s) == 0 {
		return nil, nil
	}
	rd := bufio.NewReader(strings.NewReader(string(s)))
	var ss SS
	var oneline []byte
	for {
		line, prefixing, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		oneline = append(oneline, line...)
		if !prefixing {
			ss = append(ss, string(oneline))
			if len(oneline) > 0 {
				oneline = oneline[:0]
			}
		}
	}
	if len(oneline) > 0 {
		ss = append(ss, string(oneline))
	}
	return ss, nil
}

func (s S) FirstRune() rune {
	if len(s) == 0 {
		return 0
	}
	return []rune(s)[0]
}

func (s S) FirstRuneString() S {
	if len(s) == 0 {
		return ""
	}
	return S([]rune(s)[0])
}

type SS []string

// Join strings seperated by sep, ignoreing empty strings
func (ss SS) Join(sep string) string {
	if len(ss) == 0 {
		return ""
	}
	var b strings.Builder
	for i, s := range ss {
		if s == "" {
			continue
		}
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(s)
	}
	return b.String()
}

func (ss SS) Map() KSet[string] {
	if len(ss) == 0 {
		return nil
	}
	m := make(KSet[string])
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func (ss SS) Remove(strs ...string) SS {
	return KS[string](ss).Remove(strs...).Slice()
}

func (ss SS) Append(s ...string) SS {
	if len(s) == 0 {
		return ss
	}
	r := append(ss, s...)
	return r
}

func (ss SS) Sort() SS        { sort.Strings(ss); return ss }
func (ss SS) Slice() []string { return ss }

func (ss SS) Clone() SS {
	if ss == nil {
		return nil
	}
	ret := make(SS, len(ss))
	copy(ret, ss)
	return ret
}

func (ss SS) Dedup(transfers ...func(input string) (string, bool)) SS {
	if len(ss) == 0 {
		return ss
	}
	var transfer func(input string) (string, bool)
	if len(transfers) > 0 && transfers[0] != nil {
		transfer = transfers[0]
	} else {
		transfer = func(input string) (string, bool) {
			s := strings.TrimSpace(input)
			if s == "" {
				return s, true
			}
			return "", false
		}
	}

	var set KSet[string]
	var ok, changed bool
	var rr SS
	for _, s := range ss {
		s, ok = transfer(s)
		if !ok {
			continue
		}
		set, changed = set.CAS(s)
		if changed {
			rr = append(rr, s)
		}
	}
	return rr
}

func (ss SS) Has(s string) bool {
	if len(ss) == 0 {
		return false
	}
	for _, sss := range ss {
		if sss == s {
			return true
		}
	}
	return false
}

// Matchs if any s in ss contained by str
func (ss SS) Matchs(str string) bool {
	if len(ss) == 0 || str == "" {
		return false
	}
	for _, s := range ss {
		if s != "" && strings.Contains(str, s) {
			return true
		}
	}
	return false
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
	var filter func(k K, v V) bool
	if len(filters) > 0 && filters[0] != nil {
		filter = filters[0]
	}
	var rs []K
	for k, v := range km {
		if filter == nil || filter(k, v) {
			rs = append(rs, k)
		}
	}
	return rs
}

func (km KMap[K, V]) Values(filters ...func(k K, v V) bool) []V {
	return slices.Collect(km.ValuesSeq(filters...))
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

type KS[K comparable] []K

func (ks KS[K]) All() iter.Seq[K] {
	return func(yield func(K) bool) {
		for _, k := range ks {
			if !yield(k) {
				return
			}
		}
	}
}

func (ks KS[K]) Dedup(validators ...func(k K) bool) KS[K] {
	if len(ks) == 0 {
		return nil
	}
	var validate func(k K) bool
	if len(validators) > 0 && validators[0] != nil {
		validate = validators[0]
	}
	if validate == nil {
		if len(ks) == 1 {
			return ks
		}
		if len(ks) == 2 {
			if ks[0] == ks[1] {
				return ks[:1]
			}
			return ks
		}
		r := make(KS[K], 0, len(ks))
		dedup := make(KSet[K])
		for _, k := range ks {
			if !dedup.IsExist(k) {
				r = append(r, k)
				dedup.Add(k)
			}
		}
		return r
	} else {
		r := make(KS[K], 0, len(ks))
		dedup := make(KSet[K])
		for _, k := range ks {
			if !validate(k) {
				continue
			}
			if !dedup.IsExist(k) {
				r = append(r, k)
				dedup.Add(k)
			}
		}
		if len(r) == 0 {
			return nil
		}
		return r
	}
}

func (ks KS[K]) Equal(os KS[K]) bool {
	if ks == nil && os == nil {
		return true
	}
	if ks == nil || os == nil {
		return false
	}
	if len(ks) != len(os) {
		return false
	}
	for i := 0; i < len(ks); i++ {
		if ks[i] != os[i] {
			return false
		}
	}
	return true
}

func (ks KS[K]) DedupRange(one func(k K) error) error {
	if len(ks) == 0 {
		return nil
	}
	dedup := make(map[K]struct{})
	for _, k := range ks {
		if _, exist := dedup[k]; !exist {
			dedup[k] = struct{}{}
			if err := one(k); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ks KS[K]) Map(validators ...func(k K) bool) KSet[K] {
	if len(ks) == 0 {
		return nil
	}
	var validate func(k K) bool
	if len(validators) > 0 && validators[0] != nil {
		validate = validators[0]
	}
	m := make(KSet[K], len(ks))
	if validate == nil {
		for _, k := range ks {
			m[k] = struct{}{}
		}
	} else {
		for _, k := range ks {
			if validate(k) {
				m[k] = struct{}{}
			}
		}
	}
	return m
}

func (ks KS[K]) Slice() []K {
	return ks
}

func (ks KS[K]) Clone() KS[K] {
	if ks == nil {
		return nil
	}
	rs := make(KS[K], len(ks))
	copy(rs[:], ks[:])
	return rs
}

func (ks KS[K]) Contains(tgt K) bool {
	for _, k := range ks {
		if k == tgt {
			return true
		}
	}
	return false
}

func (ks KS[K]) IterateRemove(tgt K) KS[K] {
	var ret = make(KS[K], 0, len(ks))
	last := 0
	for i := 0; i < len(ks); i++ {
		if ks[i] == tgt {
			if i > last {
				ret = append(ret, ks[last:i]...)
			}
			last = i + 1
		}
	}
	if last < len(ks) {
		ret = append(ret, ks[last:]...)
	}
	return ret
}

func (ks KS[K]) Remove(tgts ...K) KS[K] {
	if len(ks) == 0 || len(tgts) == 0 {
		return ks
	}
	if len(tgts) == 1 {
		return ks.IterateRemove(tgts[0])
	} else if len(ks) == 1 {
		if KS[K](tgts).Contains(ks[0]) {
			return nil
		}
		return ks
	} else {
		set := KSet[K](nil).Append(tgts...)
		ret := make(KS[K], 0, len(ks))
		for _, k := range ks {
			if !set.IsExist(k) {
				ret = append(ret, k)
			}
		}
		return ret
	}
}

func (ks KS[K]) Sub(offset, limit int) KS[K] {
	if len(ks) == 0 {
		return nil
	}
	if offset < len(ks) {
		end := offset + limit
		if end > len(ks) {
			end = len(ks)
		}
		return ks[offset:end]
	}
	return nil
}

func (ks KS[K]) Append(k K) KS[K] {
	return append(ks, k)
}

type KV[K comparable, V any] struct{}

func (KV[K, V]) OneOfMap(k K, mapper func([]K) (map[K]V, error)) (v V, err error) {
	var m map[K]V
	m, err = mapper([]K{k})
	if err != nil {
		return
	}
	return m[k], nil
}

func (KV[K, V]) List(keys []K, valuesMap map[K]V) []V {
	if len(keys) == 0 {
		return []V{}
	}
	r := make([]V, len(keys))
	for i := 0; i < len(keys); i++ {
		r[i] = valuesMap[keys[i]]
	}
	return r
}

func (KV[K, V]) MapKeys(m map[K]V) []K {
	if len(m) == 0 {
		return nil
	}
	ks := make([]K, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func (KV[K, V]) MapValues(m map[K]V) []V {
	if len(m) == 0 {
		return nil
	}
	vs := make([]V, 0, len(m))
	for _, v := range m {
		vs = append(vs, v)
	}
	return vs
}

func (KV[K, V]) RangeByK(ks []K, m map[K]V, dowork func(k K, v V) error) error {
	for _, k := range ks {
		if err := dowork(k, m[k]); err != nil {
			return err
		}
	}
	return nil
}

func (KV[K, V]) KeyNotInMap(m map[K]V, ks ...K) []K {
	if len(ks) == 0 {
		return nil
	}
	if len(m) == 0 {
		return ks
	}
	var r []K
	for _, k := range ks {
		if _, exist := m[k]; !exist {
			r = append(r, k)
		}
	}
	return r
}

func (KV[K, V]) PutToMap(m map[K]V, k K, v V) map[K]V {
	if m == nil {
		m = make(map[K]V)
	}
	m[k] = v
	return m
}
