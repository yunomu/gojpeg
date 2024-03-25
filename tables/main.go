package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func hex(i byte) string {
	return "0x" + strings.ToUpper(strconv.FormatUint(uint64(i), 16))
}

func printBS(bs []byte) {
	var ss []string
	for _, b := range bs {
		ss = append(ss, hex(b))
	}

	fmt.Print("[")
	fmt.Print(strings.Join(ss, ", "))
	fmt.Println("]")
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

		switch m {
		case 0xD8, 0xD9, 0xD0:
			slog.Info("marker", "val", hex(m))
			continue
		default:
			lh, err := r.ReadByte()
			if err != nil {
				slog.Error("ReadByte(lh)", "err", err)
				return
			}
			ll, err := r.ReadByte()
			if err != nil {
				slog.Error("ReadByte(ll)", "err", err)
				return
			}
			l := uint16(lh)<<8 + uint16(ll)

			slog.Info("marker", "val", hex(m), "len", l)

			var bs []byte
			for i := 0; i < int(l-2); i++ {
				b, err := r.ReadByte()
				if err != nil {
					slog.Error("ReadByte(table)", "err", err)
					return
				}
				bs = append(bs, b)
			}
			printBS(bs)
		}
	}
}
