package tools

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

type ID int64

func (id ID) IsValid() bool     { return id > 0 }
func (id ID) PadZeros() string  { return fmt.Sprintf("%08d", id) }
func (id ID) String() string    { return strconv.FormatInt(int64(id), 10) }
func (id ID) Int() int64        { return int64(id) }
func (id ID) Incr() int64       { return int64(id) }
func (id ID) Decr() int64       { return -int64(id) }
func (id ID) FullBytes() []byte { return binary.BigEndian.AppendUint64(nil, uint64(id)) }

// F for float64
type F float64

func (f F) ID() (i ID, overflow bool) {
	if f > math.MaxInt64 {
		return math.MaxInt64, true
	} else if f < 0 {
		return 0, true
	}
	return ID(int64(f)), false
}

func (f F) MustID() ID {
	i, _ := f.ID()
	return i
}

func (f F) Int() (i int64, flowed bool) {
	if f < math.MinInt64 {
		return math.MinInt64, true
	}
	if f > math.MaxInt64 {
		return math.MaxInt64, true
	}
	return int64(f), false
}

func (f F) MustInt() int64 {
	i, _ := f.Int()
	return i
}

type (
	// B for byte
	B  byte
	Bs []byte
)

func createPopCountDict() func(b byte) byte {
	dict := [256]byte{
		0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4, // 0xX
		1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, // 0x1X
		1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, // 0x2X
		2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, // 0x3X
		1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, // 0x4X
		2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, // 0x5X
		2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, // 0x6X
		3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, // 0x7X
		1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, // 0x8X
		2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, // 0x9X
		2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, // 0xAX
		3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, // 0xBX
		2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, // 0xCX
		3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, // 0xDX
		3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, // 0xEX
		4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8, // 0xFX
	}
	return func(b byte) byte { return dict[b] }
}

var popCountFunc = createPopCountDict()

func (b B) PopCount() int {
	return int(popCountFunc(byte(b)))
}

func (bs Bs) PopCount() int {
	var cnt int
	for _, b := range bs {
		cnt += int(popCountFunc(b))
	}
	return cnt
}
