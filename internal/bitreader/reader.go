package bitreader

import (
	"errors"

	"github.com/yunomu/jpeg/internal/marker"
	"github.com/yunomu/jpeg/internal/reader"
)

var (
	EOS = errors.New("end of scan")
)

type Reader interface {
	NextBit() (uint16, error)
	Receive(int) (uint8, error)
}

type ByteReader struct {
	r reader.Reader

	bitMask uint8
	bits    uint8
	numLine uint16
}

var _ Reader = (*ByteReader)(nil)

func New(r reader.Reader) *ByteReader {
	return &ByteReader{
		r: r,
	}
}

func (r *ByteReader) readDNL() (uint16, error) {
	ld, err := r.r.ReadUint16()
	if err != nil {
		return 0, err
	}

	if ld != 4 {
		return 0, errors.New("Unexpected NDL length")
	}

	return r.r.ReadUint16()
}

func (r *ByteReader) NextBit() (uint16, error) {
	if r.bitMask == 0 {
		b, err := r.r.ReadUint8()
		if err == reader.ErrUnexpectedMarker {
			r.r.Unread()
			m, err := r.r.ReadMarker()
			if err != nil {
				return 0, err
			}

			if m == marker.DNL {
				l, err := r.readDNL()
				if err != nil {
					return 0, err
				}

				r.numLine = l

				return 0, EOS
			}

			// XXX
			//if m == marker.EOI {
			//	return 0, EOS
			//}

			return 0, reader.ErrUnexpectedMarker
		} else if err != nil {
			return 0, err
		}

		r.bits = b
		r.bitMask = 0b1000_0000
		r.numLine++
	}

	bitMask := r.bitMask
	r.bitMask >>= 1
	if r.bits&bitMask == 0 {
		return 0b0, nil
	} else {
		return 0b1, nil
	}
}

func (r *ByteReader) Receive(l int) (uint8, error) {
	var ret uint8
	for i := 0; i < l; i++ {
		b, err := r.NextBit()
		if err != nil {
			return 0, err
		}

		ret = ret<<1 + uint8(b)
	}

	return ret, nil
}
