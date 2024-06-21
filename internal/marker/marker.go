package marker

import "fmt"

type Marker byte

const (
	SOF0   Marker = 0xC0
	SOF1          = 0xC1
	SOF2          = 0xC2
	SOF3          = 0xC3
	SOF5          = 0xC5
	SOF6          = 0xC6
	SOF7          = 0xC7
	JPG           = 0xC8
	SOF9          = 0xC9
	SOF10         = 0xCA
	SOF11         = 0xCB
	SOF13         = 0xCD
	SOF14         = 0xCE
	SOF15         = 0xCF
	DHT           = 0xC4
	DAC           = 0xCC
	RST_0         = 0xD0
	SOI           = 0xD8
	EOI           = 0xD9
	SOS           = 0xDA
	DQT           = 0xDB
	DNL           = 0xDC
	DRI           = 0xDD
	DHP           = 0xDE
	EXP           = 0xDF
	APP_n         = 0xE0
	JPG_n         = 0xF0
	COM           = 0xFE
	TEM           = 0x01
	FF            = 0x00
	Prefix        = 0xFF
)

func (m Marker) String() string {
	return fmt.Sprintf("0xFF%02X", int(m))
}

func (m Marker) RST() int {
	i := int(m - RST_0)
	if i < 0 || i >= 8 {
		return -1
	}
	return i
}

var frameMarkers = []Marker{
	SOF0,
	SOF1,
	SOF2,
	SOF3,
	SOF5,
	SOF6,
	SOF7,
	SOF9,
	SOF10,
	SOF11,
	SOF13,
	SOF14,
	SOF15,
}

func (m Marker) IsFrameMarker() bool {
	for _, fm := range frameMarkers {
		if m == fm {
			return true
		}
	}
	return false
}
