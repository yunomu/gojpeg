package main

import (
	"flag"
	"image"
	"image/jpeg"
	"log/slog"
	"os"
)

func init() {
	flag.Parse()
}

func main() {
	m, t, err := image.Decode(os.Stdin)
	if err != nil {
		slog.Error("image.Decode", "err", err)
		return
	}

	slog.Info("image.Decode", "format", t)

	if err := jpeg.Encode(os.Stdout, m, nil); err != nil {
		slog.Error("jpeg.Encode", "err", err)
		return
	}
}
