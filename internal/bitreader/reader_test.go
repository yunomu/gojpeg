package bitreader

import (
	"bytes"
	"testing"

	"github.com/yunomu/jpeg/internal/reader"
)

func TestReadBit(t *testing.T) {
	d := New(
		reader.NewStream(
			bytes.NewReader([]byte{
				0b1000_0001,
				0b1010_0000,
			}),
		),
	)

	b0, err := d.NextBit()
	if err != nil {
		t.Fatalf("NextBit[0]: %v", err)
	}

	if b0 != 1 {
		t.Errorf("bits[0] is not 1")
	}

	for i := 1; i <= 6; i++ {
		b, err := d.NextBit()
		if err != nil {
			t.Fatalf("NextBit[%d]: %v", i, err)
		}

		if b != 0 {
			t.Errorf("bits[%d] is not 0", i)
		}
	}

	b7, err := d.NextBit()
	if err != nil {
		t.Fatalf("NextBit[7]: %v", err)
	}

	if b7 != 1 {
		t.Errorf("bits[7] is not 1")
	}

	b8, err := d.NextBit()
	if err != nil {
		t.Fatalf("NextBit[8]: %v", err)
	}

	if b8 != 1 {
		t.Errorf("bits[8] is not 1")
	}

	b9, err := d.NextBit()
	if err != nil {
		t.Fatalf("NextBit[9]: %v", err)
	}

	if b9 != 0 {
		t.Errorf("bits[9] is not 0")
	}

	b10, err := d.NextBit()
	if err != nil {
		t.Fatalf("NextBit[10]: %v", err)
	}

	if b10 != 1 {
		t.Errorf("bits[10] is not 1")
	}
}

func TestReadBits(t *testing.T) {
	d := New(
		reader.NewStream(
			bytes.NewReader([]byte{
				0b1000_0001,
				0b1010_0000,
			}),
		),
	)

	b0, err := d.NextBit()
	if err != nil {
		t.Fatalf("NextBit[0]: %v", err)
	}
	if b0 != 1 {
		t.Errorf("bits[0] is not 1")
	}

	b1, err := d.NextBit()
	if err != nil {
		t.Fatalf("NextBit[1]: %v", err)
	}

	if b1 != 0 {
		t.Errorf("bits[1] is not 0")
	}

	v, err := d.Receive(7)
	if err != nil {
		t.Fatalf("NextBit[2:8]: %v", err)
	}

	if v != 3 {
		t.Errorf("bits[2:8] is not 3, actual=%v", v)
	}
}
