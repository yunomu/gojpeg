package huffman

import (
	"bytes"
	"testing"
)

func TestHistogram(t *testing.T) {
	r := bytes.NewReader([]byte{
		5, 5, 5, 5, 5, 5,
		4, 4, 4, 4, 4,
		3, 3, 3, 3,
		2, 2, 2,
		1, 1,
		0,
	})

	bins, err := histogram(r)
	if err != nil {
		t.Fatalf("Histogram: %v", err)
	}

	if len(bins) != 6 {
		t.Errorf("length error: exp=%v act=%v", 6, len(bins))
	}

	for i, bin := range bins {
		if i != int(bin.b) {
			t.Errorf("value mismatch")
		}
		if i+1 != int(bin.count) {
			t.Errorf("count mismatch")
		}
	}
}
