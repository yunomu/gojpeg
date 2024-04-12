package decoder

import (
	"bytes"
	"testing"
)

func TestDecodeHufftable(t *testing.T) {
	bits := []byte{0, 1, 5, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0}
	vals := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	d := New(bytes.NewReader(append(
		[]byte{0x12}, // Tc, Th
		append(bits, vals...)...,
	)))

	hufftable, l, err := d.readHTn()
	if err != nil {
		t.Fatalf("decode hufftable: %v", err)
	}

	if l != 29 {
		t.Errorf("l=%d", l)
	}

	if hufftable.class != 1 {
		t.Errorf("Tc=%v", hufftable.class)
	}
	if hufftable.target != 2 {
		t.Errorf("Th=%v", hufftable.target)
	}

	expcodes := []uint16{
		0b00000000000000_00,
		0b0000000000000_010,
		0b0000000000000_011,
		0b0000000000000_100,
		0b0000000000000_101,
		0b0000000000000_110,
		0b000000000000_1110,
		0b00000000000_11110,
		0b0000000000_111110,
		0b000000000_1111110,
		0b00000000_11111110,
		0b0000000_111111110,
	}
	for i, code := range hufftable.huffcodes {
		if code.value != vals[i] {
			t.Errorf("code[%d]=%d vals[%d]=%d", i, code.value, i, vals[i])
		}

		if code.code != expcodes[i] {
			t.Errorf("code[%d]=0b%016b expcodes[%d]=0b%016b", i, code.code, i, expcodes[i])
		}
	}
}
