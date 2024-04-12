package decoder

import (
	"image"
)

func makeImage_(h *frameHeader, cs map[uint8][]block) image.Image {
	vx := padding(8*int(h.hMax), int(h.x))
	vy := padding(8*int(h.vMax), int(h.y))
	img := image.NewYCbCr(image.Rect(0, 0, vx, vy), image.YCbCrSubsampleRatio420)

	for _, fp := range h.params {
		blocks := cs[fp.c]
		var x, y int
		var cnt int
		for {
			for i := 0; i < int(fp.v); i++ {
				for j := 0; j < int(fp.h); j++ {
					for bi, v := range blocks[cnt] {
						by := bi / 8
						bx := bi % 8
						idx := x + j*8 + bx + (y+i*8+by)*int(fp.x)
						switch fp.c {
						case 1:
							img.Y[idx] = uint8(v)
						case 2:
							img.Cb[idx] = uint8(v)
						case 3:
							img.Cr[idx] = uint8(v)
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
