package huffman

import (
	"errors"

	"github.com/yunomu/jpeg/internal/bitreader"
)

func Decode(
	r bitreader.Reader,
	ht *Table,
) (uint8, error) {
	l := 1
	code, err := r.NextBit()
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

		nextbit, err := r.NextBit()
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
