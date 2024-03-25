package decoder

import (
	"fmt"
	"log/slog"

	"gonum.org/v1/gonum/mat"
)

type quantizationTable struct {
	precision, target uint8
	qs                [64]uint16

	mat mat.Matrix
}

func (t *quantizationTable) String() string {
	return fmt.Sprintf("(Pq=%d Tq=%d Q=%v)", t.precision, t.target, t.qs)
}

func (t *quantizationTable) Unquantize(zz []int16) []int16 {
	var ret [64]int16
	for i, q := range t.qs {
		ret[i] = zz[i] * int16(q)
	}

	return ret[:]
}

func (d *Decoder) readQT() (*quantizationTable, int, error) {
	t, err := d.readUint8()
	if err != nil {
		return nil, 0, nil
	}
	pq := t >> 4
	tq := 0x0F & t

	var qs [64]uint16
	size := 1
	for i := 0; i < 64; i++ {
		var v uint16
		var err error
		if pq == 0 {
			v_, err_ := d.readUint8()
			v = uint16(v_)
			err = err_
			size++
		} else {
			v_, err_ := d.readUint16()
			v = v_
			err = err_
			size += 2
		}
		if err != nil {
			return nil, 0, nil
		}

		qs[i] = v
	}

	return &quantizationTable{
		precision: pq,
		target:    tq,
		qs:        qs,
	}, size, nil
}

func (d *Decoder) readDQT() ([]*quantizationTable, error) {
	lq, err := d.readUint16()
	if err != nil {
		return nil, err
	}

	slog.Info("define quantization tables",
		"Lq", lq,
	)

	rem := int(lq) - 2
	var ret []*quantizationTable
	for rem > 0 {
		qt, l, err := d.readQT()
		if err != nil {
			return nil, err
		}

		rem -= l
		ret = append(ret, qt)
	}

	return ret, nil
}
