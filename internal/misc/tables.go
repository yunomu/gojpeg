package misc

import (
	"errors"

	"github.com/yunomu/jpeg/internal/huffman"
	"github.com/yunomu/jpeg/internal/marker"
	"github.com/yunomu/jpeg/internal/quantization"
	"github.com/yunomu/jpeg/internal/reader"
)

type Tables struct {
	HuffmanTables      [huffman.MaxClasses][huffman.MaxTargets]*huffman.Table
	QuantizationTables [quantization.MaxTargets]*quantization.Table
	Interval           int
	Comments           [][]byte
	APs                [][]byte
}

func (t *Tables) clone() *Tables {
	var ret Tables

	for c, hts := range t.HuffmanTables {
		for h, ht := range hts {
			ret.HuffmanTables[c][h] = ht
		}
	}

	for h, qt := range t.QuantizationTables {
		ret.QuantizationTables[h] = qt
	}

	ret.Interval = t.Interval

	for _, com := range t.Comments {
		ret.Comments = append(ret.Comments, com)
	}

	for _, ap := range t.APs {
		ret.APs = append(ret.APs, ap)
	}

	return &ret
}

func (t *Tables) Cascade(o *Tables) *Tables {
	ret := t.clone()

	for c, hts := range o.HuffmanTables {
		for h, ht := range hts {
			ret.HuffmanTables[c][h] = ht
		}
	}

	for h, qt := range o.QuantizationTables {
		ret.QuantizationTables[h] = qt
	}

	ret.Interval = o.Interval

	for _, com := range o.Comments {
		ret.Comments = append(ret.Comments, com)
	}

	for _, ap := range o.APs {
		ret.APs = append(ret.APs, ap)
	}

	return ret
}

func readDRI(r reader.Reader) (uint16, error) {
	lr, err := r.ReadUint16()
	if err != nil {
		return 0, err
	}

	if lr != 4 {
		return 0, errors.New("Invalid DRI")
	}

	return r.ReadUint16()
}

func ReadMiscTables(r reader.Reader) (*Tables, error) {
	var ret Tables
	ret.Interval = -1

	for {
		m, err := r.ReadMarker()
		if err != nil {
			return nil, err
		}

		if m.IsFrameMarker() {
			r.Unread()
			return &ret, nil
		}

		switch m {
		case marker.DQT:
			qts, err := quantization.ReadDQT(r)
			if err != nil {
				return nil, err
			}

			for _, qt := range qts {
				ret.QuantizationTables[qt.Target] = qt
			}

		case marker.DHT:
			hts, err := huffman.ReadDHT(r)
			if err != nil {
				return nil, err
			}

			for _, ht := range hts {
				ret.HuffmanTables[ht.Class][ht.Target] = ht
			}

		case marker.DAC:
			return nil, errors.New("Arithmetic coding is not supported")

		case marker.DRI:
			ri, err := readDRI(r)
			if err != nil {
				return nil, err
			}

			ret.Interval = int(ri)

		case marker.COM:
			lc, err := r.ReadUint16()
			if err != nil {
				return nil, err
			}

			bs, err := r.ReadBytes(int(lc))
			if err != nil {
				return nil, err
			}

			ret.Comments = append(ret.Comments, bs)

		case marker.APP_n:
			lp, err := r.ReadUint16()
			if err != nil {
				return nil, err
			}

			bs, err := r.ReadBytes(int(lp))
			if err != nil {
				return nil, err
			}

			ret.APs = append(ret.APs, bs)

		default:
			r.Unread()
			return &ret, nil
		}
	}
}
