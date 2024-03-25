package decoder

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestFindHufftable(t *testing.T) {
	ht := findHufftable([]*hufftable{
		{
			class:  0,
			target: 0,
		},
		{
			class:  0,
			target: 1,
		},
		{
			class:  1,
			target: 0,
		},
		{
			class:  1,
			target: 1,
		},
	}, 0, 1)
	if ht == nil {
		t.Fatalf("ht is nill")
	}

	if ht.class != 0 {
		t.Errorf("class is not 0")
	}
	if ht.target != 1 {
		t.Errorf("target is not 1")
	}
}

func TestPow(t *testing.T) {
	for i := 1; i < 8; i++ {
		exp := int(math.Pow(2, float64(i-1)))
		act := 1 << (i - 1)
		if exp != act {
			t.Errorf("i=%v exp=%v act=%v", i, exp, act)
		}
	}
}

func BenchmarkPow_Math(b *testing.B) {
	t := 7
	var v int
	for i := 0; i < b.N; i++ {
		v = int(math.Pow(2, float64(t)))
	}
	var _ = v
}

func BenchmarkPow_Shift(b *testing.B) {
	t := 7
	var v int
	for i := 0; i < b.N; i++ {
		v = 1 << t
	}
	var _ = v
}

func TestExtend(t *testing.T) {
	exp := []int16{-3, -2, 2, 3}
	for i := 0; i < 4; i++ {
		v := extend(uint8(i), 2)
		if exp[i] != v {
			t.Errorf("extend(%v,2)=%08b %v", i, v, int8(v))
		}
	}
}

func TestZZMatrix(t *testing.T) {
	var zz []int16
	for i := int16(0); i < 64; i++ {
		zz = append(zz, i)
	}

	act := zzToMatrix(zz)
	for i, v := range zzOrder {
		y := i / 8
		x := i % 8
		if float64(v) != act.At(y, x) {
			t.Errorf("[%d] x=%v y=%v exp=%v act=%v", i, x, y, v, act.At(x, x))
		}
	}
	t.Logf("\n%v", mat.Formatted(act, mat.Squeeze()))
}

func TestPadding_8(t *testing.T) {
	v := padding(8, 8)
	if v == 8 {
		return
	}
	t.Errorf("exp=%v act=%v", 8, v)
}

func TestPadding_5(t *testing.T) {
	v := padding(8, 5)
	if v == 8 {
		return
	}
	t.Errorf("exp=%v act=%v", 8, v)
}
