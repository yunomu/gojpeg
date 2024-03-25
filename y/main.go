package main

import (
	"flag"
	"image"
	"image/color"
	"image/jpeg"
	"log/slog"
	"os"
)

var (
	debug = flag.Bool("debug", false, "")
)

func init() {
	flag.Parse()

	if *debug {
		level := new(slog.LevelVar)
		level.Set(slog.LevelDebug)
		slog.SetDefault(slog.New(
			slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: level,
			}),
		))
	}
}

func main() {
	img, fmt, err := image.Decode(os.Stdin)
	if err != nil {
		slog.Error("image.Decode", "err", err)
		return
	}
	slog.Info("Decode", "format", fmt, "colorModel", img.ColorModel(), "bounds", img.Bounds())

	//out := image.NewNRGBA(img.Bounds())
	out := image.NewYCbCr(img.Bounds(), image.YCbCrSubsampleRatio420)
	for j := 0; j < img.Bounds().Max.Y; j++ {
		for i := 0; i < img.Bounds().Max.X; i++ {
			r, g, b, a := img.At(i, j).RGBA()
			slog.Debug("Color", "x", i, "y", j, "r", r, "g", g, "b", b, "a", a)
			y, _, _ := color.RGBToYCbCr(uint8(r>>8), uint8(g>>8), uint8(b>>8))
			out.Y[j*img.Bounds().Max.Y+i] = y
		}
	}

	if err := jpeg.Encode(os.Stdout, out, nil); err != nil {
		slog.Error("jpeg.Encode", "err", err)
		return
	}
}
