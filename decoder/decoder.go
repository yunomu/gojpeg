package decoder

import (
	"bufio"
	"errors"
	"image/jpeg"
	"io"
	"log/slog"
	"os"
)

type Decoder struct {
	r *bufio.Reader

	prevByte   byte
	prevMarker Marker
	unreaded   bool

	bitMask uint8
	bits    uint8
	numLine uint16

	pred int16
}

func New(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

var (
	ErrUnexpectedByte   = errors.New("unexpected byte")
	ErrUnexpectedMarker = errors.New("unexpected marker")
)

func (d *Decoder) readByteMarker() (byte, Marker, error) {
	if d.unreaded {
		b, m := d.prevByte, d.prevMarker

		d.unreaded = false
		return b, m, nil
	}

	b, err := d.r.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	if b != Marker_Prefix {
		d.prevByte = b
		d.prevMarker = 0
		return b, 0, nil
	}

	m, err := d.r.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	if m == Marker_FF {
		d.prevByte = b
		d.prevMarker = 0
		return b, 0, nil
	}

	d.prevByte = 0
	d.prevMarker = Marker(m)
	return 0, Marker(m), nil
}

func (d *Decoder) unread() {
	d.unreaded = true
}

func (d *Decoder) readByte() (byte, error) {
	b, m, err := d.readByteMarker()
	if err != nil {
		return 0, err
	}

	if m != 0 {
		return 0, ErrUnexpectedMarker
	}

	return b, nil
}

func (d *Decoder) readMarker() (Marker, error) {
	b, m, err := d.readByteMarker()
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

func (d *Decoder) readBytes(n int) ([]byte, error) {
	var ret []byte

	for i := 0; i < n; i++ {
		b, err := d.readByte()
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

func (d *Decoder) readUint16() (uint16, error) {
	bs, err := d.readBytes(2)
	if err != nil {
		return 0, err
	}

	return (uint16(bs[0]) << 8) | uint16(bs[1]), nil
}

func (d *Decoder) readUint8() (uint8, error) {
	b, err := d.readByte()
	if err != nil {
		return 0, err
	}

	return uint8(b), nil
}

func (d *Decoder) readDRI() (uint16, error) {
	lr, err := d.readUint16()
	if err != nil {
		return 0, err
	}

	if lr != 4 {
		return 0, errors.New("Invalid DRI")
	}

	return d.readUint16()
}

type miscTables struct {
	hufftables         []*hufftable
	quantizationTables []*quantizationTable
	interval           int
}

func (t *miscTables) cascade(t1 *miscTables) *miscTables {
	ret := &miscTables{
		hufftables:         t.hufftables,
		quantizationTables: t.quantizationTables,
		interval:           t.interval,
	}

	if t1.hufftables != nil {
		ret.hufftables = t1.hufftables
	}
	if t1.quantizationTables != nil {
		ret.quantizationTables = t1.quantizationTables
	}
	if t1.interval != -1 {
		ret.interval = t1.interval
	}

	return ret
}

func (d *Decoder) decodeMisc() (*miscTables, error) {
	var ret miscTables
	ret.interval = -1

	for {
		m, err := d.readMarker()
		if err != nil {
			return nil, err
		}

		if m.isFrameMarker() {
			d.unread()
			return &ret, nil
		}

		switch m {
		case Marker_DQT:
			qts, err := d.readDQT()
			if err != nil {
				return nil, err
			}
			slog.Debug("DQT", "tables", qts)
			ret.quantizationTables = append(ret.quantizationTables, qts...)

		case Marker_DHT:
			hts, err := d.readDHT()
			if err != nil {
				return nil, err
			}
			ret.hufftables = append(ret.hufftables, hts...)

		case Marker_DRI:
			ri, err := d.readDRI()
			if err != nil {
				return nil, err
			}
			ret.interval = int(ri)

		case Marker_DAC, Marker_COM, Marker_APP_n:
			l, err := d.readUint16()
			if err != nil {
				return nil, err
			}

			slog.Info("other header",
				"marker", m,
				"length", l,
			)

			// skip header
			if _, err := d.readBytes(int(l - 2)); err != nil {
				slog.Error("skip")
				return nil, err
			}
		default:
			d.unread()
			return &ret, nil
		}
	}
}

func (d *Decoder) decodeFrame(misc *miscTables) (*frameHeader, map[uint8][]block, error) {
	header, err := d.readFrameHeader()
	if err != nil {
		return nil, nil, err
	}

	for {
		misc1, err := d.decodeMisc()
		if err != nil {
			return nil, nil, err
		}

		m, err := d.readMarker()
		if err != nil {
			return nil, nil, err
		}

		if m != Marker_SOS {
			slog.Error("marker is not start of scan", "marker", m)
			return nil, nil, ErrUnexpectedMarker
		}

		cs, err := d.decodeScan(header, misc.cascade(misc1))
		if err != nil {
			return nil, nil, err
		}

		if err := d.readEOI(); err == ErrUnexpectedMarker {
			d.unread()
			continue
		} else if err != nil {
			return nil, nil, err
		}

		return header, cs, nil
	}
}

func (d *Decoder) readSOI() error {
	m, err := d.readMarker()
	if err != nil {
		return err
	}

	if m != Marker_SOI {
		slog.Debug("readSOI() unexpected marker", "marker", m)
		return ErrUnexpectedMarker
	}

	return nil
}

func (d *Decoder) readEOI() error {
	m, err := d.readMarker()
	if err != nil {
		return err
	}

	if m != Marker_EOI {
		slog.Debug("readEOI() unexpected marker", "marker", m)
		return ErrUnexpectedMarker
	}

	return nil
}

func (d *Decoder) Decode() error {
	if err := d.readSOI(); err != nil {
		return err
	}

	misc, err := d.decodeMisc()
	if err != nil {
		return err
	}

	hdr, cs, err := d.decodeFrame(misc)
	if err != nil {
		return err
	}

	img := makeImage_(hdr, cs)
	return jpeg.Encode(os.Stdout, img, nil)
}
