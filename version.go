package tools

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Version 使用64位有符号整数，分为3部分，major.minor.patch，每部分使用6位数，最大支持999,999
// patch 使用十进制最低6位数
// minor 使用十进制第12位到第7位数
// major 使用十进制第18位到第13位数
// 符号为+时，为正式版，为-时，为alpha版，0.0.0没有alpha版
type Version int64

func (v Version) Major() uint64 {
	return uint64(Abs(int64(v)) / 1_000_000_000_000)
}

func (v Version) Minor() uint64 {
	return uint64((Abs(int64(v)) % 1_000_000_000_000) / 1_000_000)
}

func (v Version) Patch() uint64 {
	return uint64(Abs(int64(v)) % 1_000_000)
}

func (v Version) Alpha() Version {
	if v < 0 {
		return v
	}
	return -v
}

func NewVersion(major, minor, patch uint64, alpha bool) (Version, error) {
	if major > 999_999 || minor > 999_999 || patch > 999_999 {
		return 0, errors.New("tools: out of range")
	}
	if major == 0 && minor == 0 && patch == 0 && alpha {
		return 0, errors.New("tools: unsupported version")
	}
	v := patch + minor*1_000_000 + major*1_000_000_000_000
	if alpha {
		v = -v
	}
	return Version(v), nil
}

func ParseVersion(s string) (Version, error) {
	ss := strings.Split(S(s).Trim().ToLower().String(), ".")
	if !(len(ss) == 3 || (len(ss) == 4 && ss[3] == "a")) {
		return 0, errors.New("invalid version string")
	}
	var ns []uint64
	for i := 0; i < 3; i++ {
		u, err := strconv.ParseUint(ss[i], 10, 64)
		if err != nil {
			return 0, err
		}
		ns = append(ns, u)
	}
	return NewVersion(ns[0], ns[1], ns[2], len(ss) == 4)
}

func (v Version) String() string {
	if v < 0 {
		return fmt.Sprintf("%d.%d.%d.a", v.Major(), v.Minor(), v.Patch())
	}
	return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
}
