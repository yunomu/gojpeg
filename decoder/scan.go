package decoder

import (
	"errors"
	"fmt"
	"log/slog"
	"math"

	"gonum.org/v1/gonum/mat"
)

var unzig []int = []int{
	0, 1, 5, 6, 14, 15, 27, 28,
	2, 4, 7, 13, 16, 26, 29, 42,
	3, 8, 12, 17, 25, 30, 41, 43,
	9, 11, 18, 24, 31, 40, 44, 53,
	10, 19, 23, 32, 39, 45, 52, 54,
	20, 22, 33, 38, 46, 51, 55, 60,
	21, 34, 37, 47, 50, 56, 59, 61,
	35, 36, 48, 49, 57, 58, 62, 63,
}

func zzToMatrix(zz block) *mat.Dense {
	var data [blockSize]float64
	for i, n := range unzig {
		data[i] = float64(zz[n])
	}

	return mat.NewDense(8, 8, data[:])
}

type scanComponentParam struct {
	cs     uint8
	td, ta uint8
}

func (p *scanComponentParam) String() string {
	return fmt.Sprintf("(Cs=%d Td=%d Ta=%d)", p.cs, p.td, p.ta)
}

type scanHeader struct {
	n      uint8
	params []*scanComponentParam
	ss     uint8
	se     uint8
	ah     uint8
	al     uint8
}

func (h *scanHeader) String() string {
	return fmt.Sprintf("(Ns=%d params=%v Ss=%d Se=%d Ah=%d Al=%d)", h.n, h.params, h.ss, h.se, h.ah, h.al)
}

type componentParam struct {
	cs    uint8
	h, v  uint8
	x, y  uint16
	nunit int
	qt    *quantizationTable // quantization table
	dcHT  *hufftable         // huffman code tables for DC
	acHT  *hufftable         // huffman code tables for AC
}

func (p *componentParam) String() string {
	return fmt.Sprintf("(cs=%d H=%d V=%d)", p.cs, p.h, p.v)
}

func findFrameComponentParam(params []*frameComponentParam, cs uint8) *frameComponentParam {
	for _, p := range params {
		if p.c == cs {
			return p
		}
	}
	return nil
}

func findHufftable(tables []*hufftable, class, target uint8) *hufftable {
	for _, t := range tables {
		if class == t.class && target == t.target {
			return t
		}
	}
	return nil
}

func findQuantizationTable(tables []*quantizationTable, target uint8) *quantizationTable {
	for _, t := range tables {
		if t.target == target {
			return t
		}
	}
	return nil
}

func padding(x int, i int) int {
	m := i % x
	if m == 0 {
		return i
	}
	return i + x - m
}

func getComponentParams(
	frameHeader *frameHeader,
	quantizationTables []*quantizationTable,
	hufftables []*hufftable,
	scanHeader *scanHeader,
) ([]*componentParam, int, error) {
	var ret []*componentParam
	for _, sp := range scanHeader.params {
		fp := findFrameComponentParam(frameHeader.params, sp.cs)
		if fp == nil {
			return nil, 0, errors.New("component param not found in frame header")
		}

		ret = append(ret, &componentParam{
			cs:    sp.cs,
			h:     fp.h,
			v:     fp.v,
			x:     fp.x,
			y:     fp.y,
			nunit: int(fp.h) * int(fp.v),
			qt:    findQuantizationTable(quantizationTables, fp.tq),
			dcHT:  findHufftable(hufftables, 0, sp.td),
			acHT:  findHufftable(hufftables, 1, sp.ta),
		})
	}

	if len(ret) == 1 {
		// non-interleave
		ret[0].nunit = 1
		return ret, padding(8, int(ret[0].x)) * padding(8, int(ret[0].y)) / (8 * 8), nil
	}

	var nmcu int
	for _, p := range ret {
		tx := 8 * int(p.h)
		ty := 8 * int(p.v)
		n := padding(tx, int(p.x)) * padding(ty, int(p.y)) / (tx * ty)
		if nmcu == 0 {
			nmcu = n
		} else if n != nmcu {
			return nil, 0, errors.New("number of MCU mismatch")
		}
	}

	return ret, nmcu, nil
}

func (d *Decoder) decodeScanHeader() (*scanHeader, error) {
	ls, err := d.readUint16()
	if err != nil {
		return nil, err
	}

	ns, err := d.readUint8()
	if err != nil {
		return nil, err
	}

	var params []*scanComponentParam
	for i := 0; i < int(ns); i++ {
		cs, err := d.readUint8()
		if err != nil {
			return nil, err
		}

		t, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		td := t >> 4
		ta := t & 0xF

		params = append(params, &scanComponentParam{
			cs: cs,
			td: td,
			ta: ta,
		})
	}

	ss, err := d.readUint8()
	if err != nil {
		return nil, err
	}

	se, err := d.readUint8()
	if err != nil {
		return nil, err
	}
	a, err := d.readUint8()
	if err != nil {
		return nil, err
	}
	ah := a >> 4
	al := a & 0xF

	var _ = ls

	return &scanHeader{
		n:      ns,
		params: params,
		ss:     ss,
		se:     se,
		ah:     ah,
		al:     al,
	}, nil
}

func extend(v_ uint8, t int) int16 {
	if t == 0 {
		return 0
	}

	var v int16 = int16(v_)
	var vt int16 = 1 << (t - 1)
	if v < vt {
		return v + (-1 << t) + 1
	}

	return v
}

func (d *Decoder) decodeDC(ht *hufftable) (int16, error) {
	l_, err := d.decodeHuffval(ht)
	if err != nil {
		return 0, err
	}
	l := int(l_)
	if l == 0 {
		return 0, nil
	}

	v, err := d.receive(l)
	if err != nil {
		return 0, err
	}

	return extend(v, l), nil
}

type ob uint8

func (b ob) String() string {
	return fmt.Sprintf("%08b", b)
}

func (d *Decoder) decodeZZ(ssss int) (int16, error) {
	v, err := d.receive(ssss)
	if err != nil {
		return 0, err
	}

	return extend(v, ssss), nil
}

func (d *Decoder) decodeACs(ht *hufftable) (block, error) {
	k := 1
	var zz block

	for {
		rs, err := d.decodeHuffval(ht)
		if err != nil {
			return block{}, err
		}

		ssss := int(rs % 16)
		rrrr := rs >> 4
		r := int(rrrr)

		if ssss == 0 {
			if r == 15 {
				k += 16
				continue
			}
			break
		}

		k += r

		v, err := d.decodeZZ(ssss)
		if err != nil {
			return block{}, err
		}
		zz[k] = v

		if k == 63 {
			break
		}
		k++
	}

	return zz, nil
}

func levelShift(p uint8, a mat.Matrix) mat.Matrix {
	s := math.Pow(2, float64(p-1))

	var ret mat.Dense
	ret.Apply(func(i, j int, v float64) float64 {
		v = math.Trunc(v)
		if v <= (-s) {
			return 0
		} else if v >= s {
			return s*2 - 1
		}
		return v + s
	}, a)

	return &ret
}

const blockSize = 64

type block [blockSize]int16

func reconstruct(a mat.Matrix) block {
	var ret block
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			v := a.At(y, x)
			ret[y*8+x] = int16(v)
		}
	}
	return ret
}

func (d *Decoder) decodeDataUnit(param *componentParam, p uint8) (block, error) {
	dc, err := d.decodeDC(param.dcHT)
	if err != nil {
		return block{}, err
	}
	d.pred[param.cs] += dc

	zz, err := d.decodeACs(param.acHT)
	if err != nil {
		return block{}, err
	}
	zz[0] = d.pred[param.cs]

	ret := reconstruct(levelShift(p, idct_(zzToMatrix(param.qt.Unquantize(zz)))))
	return ret, nil
}

func (d *Decoder) decodeMCU(params []*componentParam, p uint8) ([][]block, error) {
	ret := make([][]block, len(params))
	param := params[0]
	for i, cidx, b := 0, 0, 0; true; i++ {
		if param.nunit <= i-b {
			cidx++
			if cidx >= len(params) {
				break
			}

			b += param.nunit
			param = params[cidx]
		}

		unit, err := d.decodeDataUnit(param, p)
		if err != nil {
			return nil, err
		}

		ret[cidx] = append(ret[cidx], unit)
	}

	return ret, nil
}

func (d *Decoder) decodeRestartInterval(params []*componentParam, nmcu int, ri int, p uint8) ([][]block, error) {
	d.pred = make(map[uint8]int16)

	var cnt, rst int
	ret := make([][]block, len(params))
	for i := 0; i < nmcu; i++ {
		units, err := d.decodeMCU(params, p)
		if err != nil {
			return nil, err
		}

		for i, mcu := range units {
			ret[i] = append(ret[i], mcu...)
		}

		cnt++
		if cnt == ri {
			d.pred = make(map[uint8]int16)

			m, err := d.readMarker()
			if err != nil {
				return nil, err
			}

			rst1 := m.RST()
			if rst1 == -1 {
				return nil, ErrUnexpectedMarker
			}

			if rst != rst1 {
				return nil, errors.New("Invalid reset marker")
			}

			cnt = 0
			rst++
			if rst == 8 {
				rst = 0
			}
		}
	}

	return ret, nil
}

func (d *Decoder) decodeScan(frameHeader *frameHeader, misc *miscTables) (map[uint8][]block, error) {
	scanHeader, err := d.decodeScanHeader()
	if err != nil {
		return nil, err
	}

	params, nmcu, err := getComponentParams(frameHeader, misc.quantizationTables, misc.hufftables, scanHeader)
	if err != nil {
		return nil, err
	}
	slog.Info("decode scan",
		"header", scanHeader,
	)

	ret := make(map[uint8][]block)
	for {
		components, err := d.decodeRestartInterval(params, nmcu, misc.interval, frameHeader.p)
		if err == ErrUnexpectedMarker {
			d.unread()
			return ret, nil
		} else if err == EOS {
			return ret, nil
		} else if err != nil {
			return nil, err
		}

		for i, param := range params {
			ret[param.cs] = append(ret[param.cs], components[i]...)
		}
	}
}
