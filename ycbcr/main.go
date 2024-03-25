package main

import (
	"flag"
	"image"
	"log/slog"
)

func init() {
	flag.Parse()
}

func main() {
	img := image.NewYCbCr(image.Rect(0, 0, 400, 400), image.YCbCrSubsampleRatio420)

	slog.Info("Image",
		"len(Y)", len(img.Y),
		"len(Cb)", len(img.Cb),
		"len(Cr)", len(img.Cr),
		"YStride", img.YStride,
		"CStride", img.CStride,
		"Rect", img.Rect,
	)
}
