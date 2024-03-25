package decoder

import (
	"fmt"
	"log/slog"
	"math"
)

type frameComponentParam struct {
	c    uint8
	h, v uint8
	tq   uint8

	x, y uint16
}

func (p *frameComponentParam) String() string {
	return fmt.Sprintf("(C=%d H=%d V=%d Tq=%d x=%d y=%d)", p.c, p.h, p.v, p.tq, p.x, p.y)
}

type frameHeader struct {
	marker     Marker
	p          uint8
	y, x       uint16
	hMax, vMax uint8
	params     []*frameComponentParam
}

func (h *frameHeader) String() string {
	return fmt.Sprintf("(marker=%v p=%d (x, y)=(%d, %d) params=%v)", h.marker, h.p, h.x, h.y, h.params)
}

func (d *Decoder) readFrameHeader() (*frameHeader, error) {
	m, err := d.readMarker()
	if err != nil {
		return nil, err
	}

	if !m.isFrameMarker() {
		slog.Error("marker is not start of frame", "marker", m)
		return nil, ErrUnexpectedMarker
	}

	lf, err := d.readUint16()
	if err != nil {
		return nil, err
	}

	p, err := d.readUint8()
	if err != nil {
		return nil, err
	}

	y, err := d.readUint16()
	if err != nil {
		return nil, err
	}

	x, err := d.readUint16()
	if err != nil {
		return nil, err
	}

	nf, err := d.readUint8()
	if err != nil {
		return nil, err
	}

	var params []*frameComponentParam
	var hmax, vmax uint8
	for i := 0; i < int(nf); i++ {
		c, err := d.readUint8()
		if err != nil {
			return nil, err
		}

		hv, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		h := hv >> 4
		v := hv & 0xF

		hmax = max(h, hmax)
		vmax = max(v, vmax)

		tq, err := d.readUint8()
		if err != nil {
			return nil, err
		}

		params = append(params, &frameComponentParam{
			c:  c,
			h:  h,
			v:  v,
			tq: tq,
		})
	}
	for _, p := range params {
		p.x = uint16(math.Ceil(float64(x) * float64(p.h) / float64(hmax)))
		p.y = uint16(math.Ceil(float64(y) * float64(p.v) / float64(vmax)))
	}

	ret := &frameHeader{
		marker: m,
		p:      p,
		y:      y,
		x:      x,
		hMax:   hmax,
		vMax:   vmax,
		params: params,
	}

	slog.Info("frame header",
		"header", ret,
		"Lf", lf,
		"Nf", nf,
	)

	return ret, nil
}
