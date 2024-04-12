package decoder

import (
	"errors"
	"fmt"
	"log/slog"
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

type hufftable struct {
	class     uint8
	target    uint8
	huffcodes []*huffcode

	maxcode map[int]int
	mincode map[int]int
	valptr  map[int]int
}

func (t *hufftable) String() string {
	return fmt.Sprintf("(class=%v target=%v huffcodes=%v)", t.class, t.target, t.huffcodes)
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

func makeHufftable(class, target uint8, bits [16]uint8, huffval []*huffval) *hufftable {
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

	return &hufftable{
		class:     class,
		target:    target,
		huffcodes: huffcodes,
		maxcode:   maxcode,
		mincode:   mincode,
		valptr:    valptr,
	}
}

func (d *Decoder) readHTn() (*hufftable, int, error) {
	t, err := d.readUint8()
	if err != nil {
		return nil, 0, err
	}
	tc := t >> 4
	th := 0x0F & t

	// BITS
	var bits [16]uint8
	for i := 0; i < 16; i++ {
		l, err := d.readUint8()
		if err != nil {
			return nil, 0, err
		}

		bits[i] = l
	}

	// HUFFVAL
	var huffvals []*huffval
	for i, l := range bits {
		for j := 0; j < int(l); j++ {
			v, err := d.readUint8()
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

	slog.Debug("huffman table",
		"Tc", tc,
		"Th", th,
		"size", 17+len(huffvals),
		"BITS", bits,
		"HUFFVAL", huffvals,
	)

	return makeHufftable(tc, th, bits, huffvals), 17 + len(huffvals), nil
}

func (d *Decoder) readDHT() ([]*hufftable, error) {
	lh, err := d.readUint16()
	if err != nil {
		return nil, err
	}

	slog.Info("define huffman tables",
		"Lh", lh,
	)

	rem := int(lh) - 2
	var ret []*hufftable
	for rem > 0 {
		t, l, err := d.readHTn()
		if err != nil {
			return nil, err
		}

		slog.Debug("hufftable",
			"size", l,
			"Tc", t.class,
			"Th", t.target,
			"huffcode", t.huffcodes,
			"maxcode", t.maxcode,
			"mincode", t.mincode,
			"valptr", t.valptr,
		)
		ret = append(ret, t)
		rem -= l
	}

	return ret, nil
}

func (d *Decoder) decodeHuffval(
	ht *hufftable,
) (uint8, error) {
	l := 1
	code, err := d.nextBit()
	if err != nil {
		return 0, err
	}

	for {
		maxcd, ok := ht.maxcode[l]
		if !ok {
			return 0, errors.New("unexpected length in maxcode")
		}
		if int(code) <= maxcd {
			break
		}

		nextbit, err := d.nextBit()
		if err != nil {
			return 0, err
		}

		l++
		code = (code << 1) + nextbit
	}

	mincd, ok := ht.mincode[l]
	if !ok {
		return 0, errors.New("unexpected length in mincode")
	}
	j := ht.valptr[l]
	j += int(code) - int(mincd)
	val := ht.huffcodes[j].value

	return val, nil
}
