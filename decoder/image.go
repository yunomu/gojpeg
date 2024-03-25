package decoder

import (
	"image"
	"log/slog"
)

func makeImage_(h *frameHeader, cs map[uint8][]block) image.Image {
	vx := padding(8*int(h.hMax), int(h.x))
	vy := padding(8*int(h.vMax), int(h.y))
	img := image.NewYCbCr(image.Rect(0, 0, vx, vy), image.YCbCrSubsampleRatio420)

	for _, fp := range h.params {
		slog.Debug("component", "fp", fp)
		blocks := cs[fp.c]
		var x, y int
		var cnt int
		for {
			for i := 0; i < int(fp.v); i++ {
				for j := 0; j < int(fp.h); j++ {
					slog.Debug("block",
						"i", i,
						"j", j,
						"x", x,
						"y", y,
						"cnt", cnt,
					)
					for by, row := range blocks[cnt] {
						for bx, v := range row {
							idx := x + j*8 + bx + (y+i*8+by)*int(fp.x)
							switch fp.c {
							case 1:
								img.Y[idx] = uint8(v)
							case 2:
								//img.Cb[idx] = uint8(v)
							case 3:
								//img.Cr[idx] = uint8(v)
							}
						}
					}
					cnt++
				}
			}

			x += int(fp.h) * 8
			if x >= int(fp.x) {
				x = 0
				y += int(fp.v) * 8
				if y >= int(fp.y) {
					break
				}
			}
		}
	}

	return img.SubImage(image.Rect(0, 0, int(h.x), int(h.y)))
}
