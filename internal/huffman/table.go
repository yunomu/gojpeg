package huffman

import (
	"fmt"

	"github.com/yunomu/jpeg/internal/reader"
)

const (
	MaxClasses = 2
	MaxTargets = 4
)

type huffval struct {
	i, j int
	v    uint8
}

func (v *huffval) String() string {
	return fmt.Sprintf("L%d[%d]=%d", v.i+1, v.j, v.v)
}

type huffcode struct {
	code  uint16
	value uint8
	size  int
}

func (c *huffcode) String() string {
	format := fmt.Sprintf("%%d:%%0%db", c.size)
	return fmt.Sprintf(format, c.value, c.code)
}

type Table struct {
	Class     uint8
	Target    uint8
	huffcodes []*huffcode

	maxcode map[int]int
	mincode map[int]int
	valptr  map[int]int
}

func (t *Table) String() string {
	return fmt.Sprintf("(class=%v target=%v huffcodes=%v)", t.Class, t.Target, t.huffcodes)
}

func makeDecoderTables(bits [16]uint8, huffcodes []*huffcode) (map[int]int, map[int]int, map[int]int) {
	maxcodes := make(map[int]int)
	mincodes := make(map[int]int)
	valptr := make(map[int]int)
	var j int
	for i := 0; i < 16; i++ {
		if bits[i] == 0 {
			maxcodes[i+1] = -1
			continue
		}

		valptr[i+1] = j
		mincodes[i+1] = int(huffcodes[j].code)
		j += int(bits[i]) - 1
		maxcodes[i+1] = int(huffcodes[j].code)
		j++
	}

	return maxcodes, mincodes, valptr
}

func makeTable(class, target uint8, bits [16]uint8, huffval []*huffval) *Table {
	var huffcodes []*huffcode

	// HUFFSIZE
	for i, l := range bits {
		for j := 0; j < int(l); j++ {
			huffcodes = append(huffcodes, &huffcode{
				size: i + 1,
			})
		}
	}

	// HUFFCODE
	var code uint16
	prev := huffcodes[0]
	for _, huffcode := range huffcodes[1:] {
		code++
		size := huffcode.size
		if prev.size != size {
			code <<= size - prev.size
		}
		huffcode.code = code
		prev = huffcode
	}

	// Order_codes
	for i, v := range huffval {
		huffcodes[i].value = v.v
	}

	maxcode, mincode, valptr := makeDecoderTables(bits, huffcodes)

	return &Table{
		Class:     class,
		Target:    target,
		huffcodes: huffcodes,
		maxcode:   maxcode,
		mincode:   mincode,
		valptr:    valptr,
	}
}

func readTable(r reader.Reader) (*Table, int, error) {
	t, err := r.ReadUint8()
	if err != nil {
		return nil, 0, err
	}
	tc := t >> 4
	th := 0x0F & t

	// BITS
	var bits [16]uint8
	for i := 0; i < 16; i++ {
		l, err := r.ReadUint8()
		if err != nil {
			return nil, 0, err
		}

		bits[i] = l
	}

	// HUFFVAL
	var huffvals []*huffval
	for i, l := range bits {
		for j := 0; j < int(l); j++ {
			v, err := r.ReadUint8()
			if err != nil {
				return nil, 0, err
			}

			huffvals = append(huffvals, &huffval{
				i: i,
				j: j,
				v: v,
			})
		}
	}

	return makeTable(tc, th, bits, huffvals), 17 + len(huffvals), nil
}

func ReadDHT(r reader.Reader) ([]*Table, error) {
	lh, err := r.ReadUint16()
	if err != nil {
		return nil, err
	}

	rem := int(lh) - 2
	var ret []*Table
	for rem > 0 {
		t, l, err := readTable(r)
		if err != nil {
			return nil, err
		}

		ret = append(ret, t)
		rem -= l
	}

	return ret, nil
}
