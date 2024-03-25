package decoder

import (
	"bytes"
	"testing"
)

func TestReadByte(t *testing.T) {
	d := New(bytes.NewReader([]byte{
		0x0A,
		0x0B,
		0xFF,
		0xC0,
		0x0C,
		0xFF,
		0x00,
		0x0D,
	}))

	b0, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[0]: %v", err)
	}

	if b0 != 0x0A {
		t.Errorf("readByte[0] mismatch: exp=0x0A act=0x%02X", int(b0))
	}

	b1, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[1]: %v", err)
	}

	if b1 != 0x0B {
		t.Errorf("readByte[1] mismatch: exp=0x0B act=0x%02X", int(b1))
	}

	m0, err := d.readMarker()
	if err != nil {
		t.Fatalf("readMarker[0]: %v", err)
	}

	if m0 != 0xC0 {
		t.Errorf("readMarker[0] mismatch: exp=0xC0 act=0x%02X", int(m0))
	}

	b2, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[2]: %v", err)
	}

	if b2 != 0x0C {
		t.Errorf("readByte[2] mismatch: exp=0x0C act=0x%02X", int(b2))
	}

	b3, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[3]: %v", err)
	}

	if b3 != 0xFF {
		t.Errorf("readByte[3] mismatch: exp=0xFF act=0x%02X", int(b3))
	}

	b4, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[4]: %v", err)
	}

	if b4 != 0x0D {
		t.Errorf("readByte[4] mismatch: exp=0x0D act=0x%02X", int(b4))
	}
}

func TestReadByte_repeat(t *testing.T) {
	d := New(bytes.NewReader([]byte{
		0x0A,
	}))

	b0, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[0]: %v", err)
	}

	if b0 != 0x0A {
		t.Errorf("readByte[0] mismatch: exp=0x0A act=0x%02X", int(b0))
	}

	d.unread()

	b1, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[1]: %v", err)
	}

	if b1 != 0x0A {
		t.Errorf("readByte[1] mismatch: exp=0x0A act=0x%02X", int(b1))
	}

	d.unread()

	b2, err := d.readByte()
	if err != nil {
		t.Fatalf("readByte[2]: %v", err)
	}

	if b2 != 0x0A {
		t.Errorf("readByte[2] mismatch: exp=0x0A act=0x%02X", int(b2))
	}
}
