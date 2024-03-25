package decoder

import "fmt"

type Marker byte

const (
	Marker_SOF0   Marker = 0xC0
	Marker_SOF1          = 0xC1
	Marker_SOF2          = 0xC2
	Marker_SOF3          = 0xC3
	Marker_SOF5          = 0xC5
	Marker_SOF6          = 0xC6
	Marker_SOF7          = 0xC7
	Marker_JPG           = 0xC8
	Marker_SOF9          = 0xC9
	Marker_SOF10         = 0xCA
	Marker_SOF11         = 0xCB
	Marker_SOF13         = 0xCD
	Marker_SOF14         = 0xCE
	Marker_SOF15         = 0xCF
	Marker_DHT           = 0xC4
	Marker_DAC           = 0xCC
	Marker_RST_0         = 0xD0
	Marker_SOI           = 0xD8
	Marker_EOI           = 0xD9
	Marker_SOS           = 0xDA
	Marker_DQT           = 0xDB
	Marker_DNL           = 0xDC
	Marker_DRI           = 0xDD
	Marker_DHP           = 0xDE
	Marker_EXP           = 0xDF
	Marker_APP_n         = 0xE0
	Marker_JPG_n         = 0xF0
	Marker_COM           = 0xFE
	Marker_TEM           = 0x01
	Marker_FF            = 0x00
	Marker_Prefix        = 0xFF
)

func (m Marker) String() string {
	return fmt.Sprintf("0xFF%02X", int(m))
}

func (m Marker) RST() int {
	i := int(m - Marker_RST_0)
	if i < 0 || i >= 8 {
		return -1
	}
	return i
}

var frameMarkers = []Marker{
	Marker_SOF0,
	Marker_SOF1,
	Marker_SOF2,
	Marker_SOF3,
	Marker_SOF5,
	Marker_SOF6,
	Marker_SOF7,
	Marker_SOF9,
	Marker_SOF10,
	Marker_SOF11,
	Marker_SOF13,
	Marker_SOF14,
	Marker_SOF15,
}

func (m Marker) isFrameMarker() bool {
	for _, fm := range frameMarkers {
		if m == fm {
			return true
		}
	}
	return false
}
