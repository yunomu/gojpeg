package main

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func hex(i uint64) string {
	return "0x" + strings.ToUpper(strconv.FormatUint(uint64(i), 16))
}

func main() {
	r := bufio.NewReader(os.Stdin)

	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			slog.Info("FINISHED")
			return
		} else if err != nil {
			slog.Error("ReadByte", "err", err)
			return
		}

		if b != 0xFF {
			continue
		}

		m, err := r.ReadByte()
		if err == io.EOF {
			slog.Error("unexpected EOF")
			return
		} else if err != nil {
			slog.Error("read marker", "err", err)
			return
		}

		if m == 0x0 {
			continue
		}

		slog.Info("marker", "val", hex(uint64(m)))
	}
}
