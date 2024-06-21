package reader

import (
	"bufio"
	"errors"
	"io"
	"log/slog"

	"github.com/yunomu/jpeg/internal/marker"
)

type Reader interface {
	ReadByteMarker() (byte, marker.Marker, error)
	Unread()
	ReadByte() (byte, error)
	ReadBytes(n int) ([]byte, error)
	ReadMarker() (marker.Marker, error)
	ReadUint8() (uint8, error)
	ReadUint16() (uint16, error)
}

type Stream struct {
	r *bufio.Reader

	prevByte   byte
	prevMarker marker.Marker
	unreaded   bool
}

var _ Reader = (*Stream)(nil)

func NewStream(r io.Reader) *Stream {
	return &Stream{
		r: bufio.NewReader(r),
	}
}

var (
	ErrUnexpectedByte   = errors.New("unexpected byte")
	ErrUnexpectedMarker = errors.New("unexpected marker")
)

func (r *Stream) ReadByteMarker() (byte, marker.Marker, error) {
	if r.unreaded {
		b, m := r.prevByte, r.prevMarker

		r.unreaded = false
		return b, m, nil
	}

	b, err := r.r.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	if b != marker.Prefix {
		r.prevByte = b
		r.prevMarker = 0
		return b, 0, nil
	}

	m, err := r.r.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	if m == marker.FF {
		r.prevByte = b
		r.prevMarker = 0
		return b, 0, nil
	}

	r.prevByte = 0
	r.prevMarker = marker.Marker(m)
	return 0, marker.Marker(m), nil
}

func (r *Stream) Unread() {
	r.unreaded = true
}

func (r *Stream) ReadByte() (byte, error) {
	b, m, err := r.ReadByteMarker()
	if err != nil {
		return 0, err
	}

	if m != 0 {
		return 0, ErrUnexpectedMarker
	}

	return b, nil
}

func (d *Stream) ReadMarker() (marker.Marker, error) {
	b, m, err := d.ReadByteMarker()
	if err != nil {
		return 0, err
	}

	if m == 0 {
		slog.Error("readMarker()",
			"marker", m,
			"byte", b,
		)
		return 0, ErrUnexpectedByte
	}

	return m, nil
}

func (r *Stream) ReadBytes(n int) ([]byte, error) {
	var ret []byte

	for i := 0; i < n; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		ret = append(ret, b)
	}

	if len(ret) != n {
		return nil, errors.New("unexpected end of stream")
	}

	return ret, nil
}

func (r *Stream) ReadUint16() (uint16, error) {
	bs, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}

	return (uint16(bs[0]) << 8) | uint16(bs[1]), nil
}

func (r *Stream) ReadUint8() (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	return uint8(b), nil
}
