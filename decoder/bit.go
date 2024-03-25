package decoder

import (
	"errors"
)

var (
	EOS = errors.New("end of scan")
)

func (d *Decoder) readDNL() (uint16, error) {
	ld, err := d.readUint16()
	if err != nil {
		return 0, err
	}

	if ld != 4 {
		return 0, errors.New("Unexpected NDL length")
	}

	return d.readUint16()
}

func (d *Decoder) nextBit() (uint16, error) {
	if d.bitMask == 0 {
		b, err := d.readUint8()
		if err == ErrUnexpectedMarker {
			d.unread()
			m, err := d.readMarker()
			if err != nil {
				return 0, err
			}

			if m == Marker_DNL {
				l, err := d.readDNL()
				if err != nil {
					return 0, err
				}

				d.numLine = l

				return 0, EOS
			}

			return 0, ErrUnexpectedMarker
		} else if err != nil {
			return 0, err
		}

		d.bits = b
		d.bitMask = 0b1000_0000
	}

	bitMask := d.bitMask
	d.bitMask >>= 1
	if d.bits&bitMask == 0 {
		return 0b0, nil
	} else {
		return 0b1, nil
	}
}

func (d *Decoder) receive(l int) (uint8, error) {
	var ret uint8
	for i := 0; i < l; i++ {
		b, err := d.nextBit()
		if err != nil {
			return 0, err
		}

		ret = ret<<1 + uint8(b)
	}

	return ret, nil
}
