package frame

import (
	"errors"
	"fmt"
	"math"

	"github.com/yunomu/jpeg/internal/marker"
	"github.com/yunomu/jpeg/internal/misc"
	"github.com/yunomu/jpeg/internal/reader"
	"github.com/yunomu/jpeg/internal/types"
)

const (
	MaxComponents = 4
)

var (
	ErrUnexpectedMarker = errors.New("Unexpected marker")
)

type ComponentParam struct {
	Cs   uint8
	H, V uint8
	Tq   uint8

	X, Y uint16
}

func (p *ComponentParam) String() string {
	return fmt.Sprintf("(C=%d H=%d V=%d Tq=%d x=%d y=%d)", p.Cs, p.H, p.V, p.Tq, p.X, p.Y)
}

type Header struct {
	Marker     marker.Marker
	P          uint8
	Y, X       uint16
	Nf         uint8
	HMax, VMax uint8
	Params     [MaxComponents]*ComponentParam
}

func (h *Header) String() string {
	return fmt.Sprintf("(marker=%v p=%d (x, y)=(%d, %d) Nf=%d params=%v)", h.Marker, h.P, h.X, h.Y, h.Nf, h.Params)
}

func readFrameHeader(r reader.Reader) (*Header, error) {
	m, err := r.ReadMarker()
	if err != nil {
		return nil, err
	}

	if !m.IsFrameMarker() {
		return nil, ErrUnexpectedMarker
	}

	lf, err := r.ReadUint16()
	if err != nil {
		return nil, err
	}

	p, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}

	y, err := r.ReadUint16()
	if err != nil {
		return nil, err
	}

	x, err := r.ReadUint16()
	if err != nil {
		return nil, err
	}

	nf, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}

	var params [MaxComponents]*ComponentParam
	var hmax, vmax uint8
	for i := 0; i < int(nf); i++ {
		c, err := r.ReadUint8()
		if err != nil {
			return nil, err
		}

		hv, err := r.ReadUint8()
		if err != nil {
			return nil, err
		}
		h := hv >> 4
		v := hv & 0xF

		hmax = max(h, hmax)
		vmax = max(v, vmax)

		tq, err := r.ReadUint8()
		if err != nil {
			return nil, err
		}

		params[int(c-1)] = &ComponentParam{
			Cs: c,
			H:  h,
			V:  v,
			Tq: tq,
		}
	}
	for _, p := range params {
		if p == nil {
			continue
		}
		p.X = uint16(math.Ceil(float64(x) * float64(p.H) / float64(hmax)))
		p.Y = uint16(math.Ceil(float64(y) * float64(p.V) / float64(vmax)))
	}

	var _ = lf
	return &Header{
		Marker: m,
		P:      p,
		Y:      y,
		X:      x,
		Nf:     nf,
		HMax:   hmax,
		VMax:   vmax,
		Params: params,
	}, nil
}

var nilCs = [MaxComponents][]types.Block{}

func ReadFrame(r reader.Reader, miscTables *misc.Tables) (*Header, [MaxComponents][]types.Block, error) {
	header, err := readFrameHeader(r)
	if err != nil {
		return nil, nilCs, err
	}

	for {
		misc1, err := misc.ReadMiscTables(r)
		if err != nil {
			return nil, nilCs, err
		}

		m, err := r.ReadMarker()
		if err != nil {
			return nil, nilCs, err
		}

		if m != marker.SOS {
			return nil, nilCs, ErrUnexpectedMarker
		}

		cs, err := readScan(r, header, miscTables.Cascade(misc1))
		if err != nil {
			return nil, nilCs, err
		}
	}
}
