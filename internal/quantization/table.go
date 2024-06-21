package quantization

import (
	"fmt"

	"github.com/yunomu/jpeg/internal/reader"
	"github.com/yunomu/jpeg/internal/types"
)

const (
	MaxTargets = 4
)

type Table struct {
	Precision uint8
	Target    uint8
	qs        types.Block
}

func (t *Table) String() string {
	return fmt.Sprintf("(Pq=%d Tq=%d Q=%v)", t.Precision, t.Target, t.qs)
}

func (t *Table) Unquantize(zz types.Block) types.Block {
	var ret types.Block
	for i, q := range t.qs {
		ret[i] = zz[i] * int32(q)
	}

	return ret
}

func readTable(r reader.Reader) (*Table, int, error) {
	var size int

	t, err := r.ReadUint8()
	if err != nil {
		return nil, 0, err
	}
	pq := t >> 4
	tq := 0x0F & t
	size++

	var qs types.Block
	for i := 0; i < types.BlockSize; i++ {
		var v int32
		var err error
		if pq == 0 {
			v_, err_ := r.ReadUint8()
			v = int32(v_)
			err = err_
			size++
		} else {
			v_, err_ := r.ReadUint8()
			v = int32(v_)
			err = err_
			size += 2
		}
		if err != nil {
			return nil, 0, err
		}

		qs[i] = v
	}

	return &Table{
		Precision: pq,
		Target:    tq,
		qs:        qs,
	}, size, nil
}

func ReadDQT(r reader.Reader) ([]*Table, error) {
	lq, err := r.ReadUint16()
	if err != nil {
		return nil, err
	}

	rem := int(lq) - 2
	var ret []*Table
	for rem > 0 {
		qt, l, err := readTable(r)
		if err != nil {
			return nil, err
		}

		rem -= l
		ret = append(ret, qt)
	}

	return ret, nil
}
