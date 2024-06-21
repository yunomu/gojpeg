package frame

import (
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/yunomu/jpeg/internal/bitreader"
	"github.com/yunomu/jpeg/internal/dct"
	"github.com/yunomu/jpeg/internal/huffman"
	"github.com/yunomu/jpeg/internal/misc"
	"github.com/yunomu/jpeg/internal/quantization"
	"github.com/yunomu/jpeg/internal/reader"
	"github.com/yunomu/jpeg/internal/types"
)

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
	qt    *quantization.Table // quantization table
	dcHT  *huffman.Table      // huffman code tables for DC
	acHT  *huffman.Table      // huffman code tables for AC
}

func (p *componentParam) String() string {
	return fmt.Sprintf("(cs=%d H=%d V=%d)", p.cs, p.h, p.v)
}

func padding(x int, i int) int {
	m := i % x
	if m == 0 {
		return i
	}
	return i + x - m
}

func getComponentParams(
	frameHeader *Header,
	miscTables *misc.Tables,
	scanHeader *scanHeader,
) ([]*componentParam, int, error) {
	var ret []*componentParam
	for _, sp := range scanHeader.params {
		fp := frameHeader.Params[sp.cs-1]
		if fp == nil {
			return nil, 0, errors.New("component param not found in frame header")
		}

		ret = append(ret, &componentParam{
			cs:    sp.cs,
			h:     fp.H,
			v:     fp.V,
			x:     fp.X,
			y:     fp.Y,
			nunit: int(fp.H) * int(fp.V),
			qt:    miscTables.QuantizationTables[fp.Tq],
			dcHT:  miscTables.HuffmanTables[0][sp.td],
			acHT:  miscTables.HuffmanTables[1][sp.ta],
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

func readScanHeader(r reader.Reader) (*scanHeader, error) {
	ls, err := r.ReadUint16()
	if err != nil {
		return nil, err
	}

	ns, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}

	var params []*scanComponentParam
	for i := 0; i < int(ns); i++ {
		cs, err := r.ReadUint8()
		if err != nil {
			return nil, err
		}

		t, err := r.ReadUint8()
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

	ss, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}

	se, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	a, err := r.ReadUint8()
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

func extend(v_ uint8, t int) int32 {
	if t == 0 {
		return 0
	}

	var v int32 = int32(v_)
	var vt int32 = 1 << (t - 1)
	if v < vt {
		return v + (-1 << t) + 1
	}

	return v
}

func decodeDC(r bitreader.Reader, ht *huffman.Table) (int32, error) {
	l_, err := huffman.Decode(r, ht)
	if err != nil {
		return 0, err
	}
	l := int(l_)
	if l == 0 {
		return 0, nil
	}

	v, err := r.Receive(l)
	if err != nil {
		return 0, err
	}

	return extend(v, l), nil
}

type ob uint8

func (b ob) String() string {
	return fmt.Sprintf("%08b", b)
}

func decodeZZ(r bitreader.Reader, ssss int) (int32, error) {
	v, err := r.Receive(ssss)
	if err != nil {
		return 0, err
	}

	return extend(v, ssss), nil
}

func decodeACs(r bitreader.Reader, ht *huffman.Table) (types.Block, error) {
	k := 1
	var zz types.Block

	for {
		rs, err := huffman.Decode(r, ht)
		if err != nil {
			return types.NilBlock, err
		}

		ssss := int(rs % 16)
		rrrr := rs >> 4
		r_ := int(rrrr)

		if ssss == 0 {
			if r_ == 15 {
				k += 16
				continue
			}
			break
		}

		k += r_

		v, err := decodeZZ(r, ssss)
		if err != nil {
			return types.NilBlock, err
		}
		zz[k] = v

		if k == 63 {
			break
		}
		k++
	}

	return zz, nil
}

func levelShift(p uint8, a types.Block) types.Block {
	s := int32(math.Pow(2, float64(p-1)))

	var ret types.Block
	for i, v := range a {
		if v <= (-s) {
			ret[i] = 0
		} else if v >= s {
			ret[i] = s*2 - 1
		} else {
			ret[i] = v + s
		}
	}

	return ret
}

func decodeDataUnit(r bitreader.Reader, param *componentParam, pred int32, p uint8) (types.Block, error) {
	dc, err := decodeDC(r, param.dcHT)
	if err != nil {
		return types.NilBlock, err
	}

	zz, err := decodeACs(r, param.acHT)
	if err != nil {
		return types.NilBlock, err
	}
	zz[0] = pred + dc

	ret := levelShift(p, dct.Idct(param.qt.Unquantize(zz)))
	return ret, nil
}

func decodeMCU(r bitreader.Reader, params []*componentParam, p uint8) ([MaxComponents][]types.Block, error) {
	var ret [MaxComponents][]types.Block
	for i, b := 0, 0; i < len(params); {
		param := params[i]

		unit, err := decodeDataUnit(r, param, p)
		if err != nil {
			return nil, err
		}

		cidx := int(param.cs - 1)
		ret[cidx] = append(ret[cidx], unit)

		b++
		if b >= param.nunit {
			i++
			b = 0
		}
	}

	return ret, nil
}

type SegmentDecoder struct {
	pred   [MaxComponents]int32
	r      bitreader.Reader
	params [MaxComponents]*componentParam
	ri     int
	p      uint8
}

func (d *SegmentDecoder) Decode() ([MaxComponents][]types.Block, error) {
	var cnt int
	var ret [MaxComponents][]types.Block
	for i := 0; i != d.ri; i++ {
		units, err := decodeMCU(r, params, p)
		if err != nil {
			return nil, err
		}

		for i, us := range units {
			ret[i] = append(ret[i], us...)
		}

		cnt++
		if cnt == d.ri {
			break
		}
	}

	return ret, nil
}

func decodeRestartInterval_(r reader.Reader, params []*componentParam, nmcu int, ri int, p uint8) ([MaxComponents][]types.Block, error) {
	var ret [MaxComponents][]types.Block
	var rst int
	for {
		d := &IntervalDecoder{
			r:      bitreader.New(r),
			params: params,
			ri:     ri,
			p:      p,
		}

		components, err := d.Decode()
		if err != nil {
			return nil, err
		}

		// process RST

		m, err := r.ReadMarker()
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

		rst++
		if rst == 8 {
			rst = 0
		}
	}
}

func decodeRestartInterval(r reader.Reader, params []*componentParam, nmcu int, ri int, p uint8) ([][]block, error) {
	d.pred = make(map[uint8]int16)

	var cnt, rst int
	ret := make([][]block, len(params))
	for i := 0; i < nmcu; i++ {
		units, err := decodeMCU(r, params, p)
		if err != nil {
			return nil, err
		}

		for i, mcu := range units {
			ret[i] = append(ret[i], mcu...)
		}

		cnt++
		if cnt == ri {
			d.pred = make(map[uint8]int16)

			m, err := r.ReadMarker()
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

func readScan(r reader.Reader, frameHeader *Header, misc *misc.Tables) (map[uint8][]block, error) {
	scanHeader, err := readScanHeader(r)
	if err != nil {
		return nil, err
	}

	params, nmcu, err := getComponentParams(frameHeader, misc, scanHeader)
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
