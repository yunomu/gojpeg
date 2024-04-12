package decoder

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"

	"github.com/yunomu/jpeg/jpeg"
)

func TestDCT_jpeg(t *testing.T) {
	var c int16
	mat := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		mat[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			mat[i][j] = c
			c++
		}
	}

	t.Logf("o=%v", mat)
	b := dct(mat)
	for i, y := range b {
		for j, x := range y {
			b[i][j] = math.Round(x)
		}
	}
	t.Logf("b=%v", b)
	a := jpeg.Idct(b)
	for i, y := range a {
		for j, x := range y {
			if mat[i][j] != int16(math.Round(x)) {
				t.Errorf("mismatch a[%d][%d]", i, j)
			}
		}
	}
	t.Logf("a=%v", a)
}

func TestDCT(t *testing.T) {
	var c int16
	mat := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		mat[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			mat[i][j] = c
			c++
		}
	}

	t.Logf("o=%v", mat)
	b := dct(mat)
	for i, y := range b {
		for j, x := range y {
			b[i][j] = math.Round(x)
		}
	}
	t.Logf("b=%v", b)
	a := idct(b)
	for i, y := range a {
		for j, x := range y {
			if mat[i][j] != int16(math.Round(x)) {
				t.Errorf("mismatch a[%d][%d]", i, j)
			}
		}
	}
	t.Logf("a=%v", a)
}

func dct0(s [][]int16) [][]float64 {
	if len(s) == 0 {
		return nil
	}

	mmax := len(s)
	nmax := len(s[0])

	ret := make([][]float64, mmax)
	for v := 0; v < mmax; v++ {
		ret[v] = make([]float64, nmax)
		for u := 0; u < nmax; u++ {
			var cv float64
			if v == 0 {
				cv = 1 / math.Sqrt(2)
			} else {
				cv = 1
			}
			var cu float64
			if u == 0 {
				cu = 1 / math.Sqrt(2)
			} else {
				cu = 1
			}
			for y := 0; y < mmax; y++ {
				for x := 0; x < nmax; x++ {
					ret[v][u] += float64(s[y][x]) *
						math.Cos(math.Pi*float64((2*y+1)*v)/16) *
						math.Cos(math.Pi*float64((2*x+1)*u)/16)
				}
			}
			ret[v][u] *= cv * cu / 4
		}
	}

	return ret
}

func dct1(a [][]int16) [][]float64 {
	if len(a) == 0 {
		return nil
	}

	mmax := len(a)
	nmax := len(a[0])

	ap := make(map[int]float64)
	ap[0] = 1 / math.Sqrt(float64(mmax))
	apv := math.Sqrt(2 / float64(mmax))
	for i := 1; i < mmax; i++ {
		ap[i] = apv
	}

	aq := make(map[int]float64)
	aq[0] = 1 / math.Sqrt(float64(nmax))
	aqv := math.Sqrt(2 / float64(nmax))
	for i := 1; i < nmax; i++ {
		aq[i] = aqv
	}

	b := make([][]float64, mmax)
	for p := 0; p < mmax; p++ {
		b[p] = make([]float64, nmax)
		for q := 0; q < nmax; q++ {
			for m := 0; m < mmax; m++ {
				for n := 0; n < nmax; n++ {
					b[p][q] += float64(a[m][n]) *
						math.Cos(math.Pi*float64((2*m+1)*p)/float64(2*mmax)) *
						math.Cos(math.Pi*float64((2*n+1)*q)/float64(2*nmax))
				}
			}
			b[p][q] *= ap[p] * aq[q]
		}
	}

	return b
}

func dct4(s [][]int16) [][]float64 {
	if len(s) == 0 {
		return nil
	}

	mmax := len(s)
	nmax := len(s[0])

	cv := make([]float64, mmax)
	cv[0] = 1 / math.Sqrt(2)
	for i := 1; i < mmax; i++ {
		cv[i] = 1
	}

	cu := make([]float64, nmax)
	cu[0] = 1 / math.Sqrt(2)
	for i := 1; i < nmax; i++ {
		cu[i] = 1
	}

	vy := make([][]float64, mmax)
	for p := 0; p < mmax; p++ {
		vy[p] = make([]float64, mmax)
		for m := 0; m < mmax; m++ {
			vy[p][m] = math.Cos(math.Pi * float64((2*m+1)*p) / 16)
		}
	}

	vx := make([][]float64, nmax)
	for q := 0; q < nmax; q++ {
		vx[q] = make([]float64, nmax)
		for n := 0; n < nmax; n++ {
			vx[q][n] = math.Cos(math.Pi * float64((2*n+1)*q) / 16)
		}
	}

	ret := make([][]float64, mmax)
	for v := 0; v < mmax; v++ {
		ret[v] = make([]float64, nmax)
		for u := 0; u < nmax; u++ {
			for y := 0; y < mmax; y++ {
				for x := 0; x < nmax; x++ {
					ret[v][u] += float64(s[y][x]) * vy[v][y] * vx[u][x]
				}
			}
			ret[v][u] *= cv[v] * cu[u] / 4
		}
	}

	return ret
}

func idct4(b [][]float64) [][]float64 {
	if len(b) == 0 {
		return nil
	}

	mmax := len(b)
	nmax := len(b[0])

	ap := make([]float64, mmax)
	ap[0] = 1 / math.Sqrt(2)
	for i := 1; i < mmax; i++ {
		ap[i] = 1
	}

	aq := make([]float64, nmax)
	aq[0] = 1 / math.Sqrt(2)
	for i := 1; i < nmax; i++ {
		aq[i] = 1
	}

	vy := make([][]float64, mmax)
	for m := 0; m < mmax; m++ {
		vy[m] = make([]float64, mmax)
		for p := 0; p < mmax; p++ {
			vy[m][p] = math.Cos(math.Pi * float64((2*m+1)*p) / 16)
		}
	}

	vx := make([][]float64, nmax)
	for n := 0; n < nmax; n++ {
		vx[n] = make([]float64, nmax)
		for q := 0; q < nmax; q++ {
			vx[n][q] = math.Cos(math.Pi * float64((2*n+1)*q) / 16)
		}
	}

	a := make([][]float64, mmax)
	for m := 0; m < mmax; m++ {
		a[m] = make([]float64, nmax)
		for n := 0; n < nmax; n++ {
			for p := 0; p < mmax; p++ {
				for q := 0; q < nmax; q++ {
					a[m][n] += ap[p] * aq[q] * b[p][q] * vy[m][p] * vx[n][q]
				}
			}
			a[m][n] /= 4
		}
	}

	return a
}

func BenchmarkDCT4(b *testing.B) {
	var c int16
	a := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		a[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			a[i][j] = c
			c++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dct4(a)
	}
}

func BenchmarkDCT0(b *testing.B) {
	var c int16
	a := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		a[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			a[i][j] = c
			c++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dct0(a)
	}
}

func BenchmarkDCT1(b *testing.B) {
	var c int16
	a := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		a[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			a[i][j] = c
			c++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dct1(a)
	}
}

var (
	cv_, cu_ [8]float64
	vy_, vx_ [8][8]float64
)

func init() {
	cv_[0] = 1 / math.Sqrt(2)
	for i := 1; i < 8; i++ {
		cv_[i] = 1
	}

	cu_[0] = 1 / math.Sqrt(2)
	for i := 1; i < 8; i++ {
		cu_[i] = 1
	}

	for v := 0; v < 8; v++ {
		for y := 0; y < 8; y++ {
			vy_[v][y] = math.Cos(math.Pi * float64((2*y+1)*v) / 16)
		}
	}

	for u := 0; u < 8; u++ {
		for x := 0; x < 8; x++ {
			vx_[u][x] = math.Cos(math.Pi * float64((2*x+1)*u) / 16)
		}
	}
}

func dct2(s [][]int16) [][]float64 {
	if len(s) == 0 {
		return nil
	}

	mmax := len(s)
	nmax := len(s[0])

	ret := make([][]float64, mmax)
	for v := 0; v < mmax; v++ {
		ret[v] = make([]float64, nmax)
		for u := 0; u < nmax; u++ {
			for y := 0; y < mmax; y++ {
				for x := 0; x < nmax; x++ {
					ret[v][u] += float64(s[y][x]) * vy_[v][y] * vx_[u][x]
				}
			}
			ret[v][u] *= cv_[v] * cu_[u] / 4
		}
	}

	return ret
}

func BenchmarkDCT2(b *testing.B) {
	var c int16
	a := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		a[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			a[i][j] = c
			c++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dct2(a)
	}
}

var (
	cosFactor [8][8][8][8]float64
	cvcu      [8][8]float64
)

func init() {
	for v := 0; v < 8; v++ {
		for u := 0; u < 8; u++ {
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					cosFactor[v][u][y][x] = vy_[v][y] * vx_[u][x]
				}
			}
			cvcu[v][u] = cv_[v] * cu_[u] / 4
		}
	}
}

func dct3(s [][]int16) [][]float64 {
	ret := make([][]float64, 8)
	for v := 0; v < 8; v++ {
		ret[v] = make([]float64, 8)
		for u := 0; u < 8; u++ {
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					ret[v][u] += float64(s[y][x]) * cosFactor[v][u][y][x]
				}
			}
			ret[v][u] *= cvcu[v][u]
		}
	}

	return ret
}

func BenchmarkDCT3(b *testing.B) {
	var c int16
	a := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		a[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			a[i][j] = c
			c++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dct3(a)
	}
}

func TestDCTMat_withorig(t *testing.T) {
	var c int16
	mat := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		mat[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			mat[i][j] = c
			c++
		}
	}

	t.Logf("o=%v", mat)
	b := dct(mat)
	for i, y := range b {
		for j, x := range y {
			b[i][j] = math.Round(x)
		}
	}
	t.Logf("b=%v", b)
	a := idct(b)
	if len(a) == 0 {
		t.Fatalf("a is empty")
	}
	for i, y := range a {
		for j, x := range y {
			if mat[i][j] != int16(math.Round(x)) {
				t.Errorf("mismatch a[%d][%d]", i, j)
			}
		}
	}
	t.Logf("a=%v", a)
}

func TestDCTMat(t *testing.T) {
	var c int16
	mat := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		mat[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			mat[i][j] = c
			c++
		}
	}

	t.Logf("o=%v", mat)
	b := dct(mat)
	for i, y := range b {
		for j, x := range y {
			b[i][j] = math.Round(x)
		}
	}
	t.Logf("b=%v", b)
	a := idct(b)
	if len(a) == 0 {
		t.Fatalf("a is empty")
	}
	for i, y := range a {
		for j, x := range y {
			if mat[i][j] != int16(math.Round(x)) {
				t.Errorf("mismatch a[%d][%d]", i, j)
			}
		}
	}
	t.Logf("a=%v", a)
}

func BenchmarkDCTMAT(b *testing.B) {
	var c int16
	a := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		a[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			a[i][j] = c
			c++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dct(a)
	}
}

func mul(a, b [][]float64) [][]float64 {
	ret := make([][]float64, len(a))
	for i, row := range a {
		ret[i] = make([]float64, len(b[0]))
		for j := 0; j < len(b[0]); j++ {
			for k, v := range row {
				ret[i][j] += v * b[k][j]
			}
		}
	}
	return ret
}

func TestMul(t *testing.T) {
	a := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}
	b := [][]float64{
		{1, 2},
		{3, 4},
		{5, 6},
	}
	r := mul(a, b)

	a0 := mat.NewDense(2, 3, []float64{1, 2, 3, 4, 5, 6})
	b0 := mat.NewDense(3, 2, []float64{1, 2, 3, 4, 5, 6})
	var r0 mat.Dense
	r0.Mul(a0, b0)

	for i, row := range r {
		for j, v := range row {
			if v != r0.At(i, j) {
				t.Errorf("mismatch r[%d][%d]=%v rr=%v", i, j, v, r0.At(i, j))
			}
		}
	}
	t.Logf("r=%v rr=%v", r, r0)
}

func transpose(a [][]float64) [][]float64 {
	var ret [][]float64
	for i, row := range a {
		if i == 0 {
			ret = make([][]float64, len(row))
		}
		for j, v := range row {
			ret[j] = append(ret[j], v)
		}
	}
	return ret
}

func TestTranspose(t *testing.T) {
	a := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}
	r := transpose(a)

	a0 := mat.NewDense(2, 3, []float64{1, 2, 3, 4, 5, 6})
	r0 := a0.T()

loop:
	for i, row := range r {
		for j, v := range row {
			if v != r0.At(i, j) {
				t.Errorf("r[%d][%d]=%v != %v", i, j, v, r0.At(i, j))
				t.Logf("r=%v", r)
				t.Logf("r =\n%v", mat.Formatted(r0, mat.Prefix(""), mat.Squeeze()))
				break loop
			}
		}
	}
}

var (
	a0, a0T [][]float64
)

func init() {
	var a00 []float64
	m0 := 1 / (2 * math.Sqrt(2))
	for i := 0; i < 8; i++ {
		a00 = append(a00, m0)
	}
	a0 = append(a0, a00)

	for i := 1; i < 8; i++ {
		var r []float64
		for j := 0; j < 8; j++ {
			r = append(r, 1/2.0*math.Cos(math.Pi*float64((2*j+1)*i)/16.0))
		}
		a0 = append(a0, r)
	}

	a0T = transpose(a0)
}

func dctmat0(si [][]int16) [][]float64 {
	var s [][]float64
	for _, row := range si {
		var r []float64
		for _, v := range row {
			r = append(r, float64(v))
		}
		s = append(s, r)
	}

	return mul(mul(a0, s), a0T)
}

func TestDCTMat0(t *testing.T) {
	var c int16
	mat := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		mat[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			mat[i][j] = c
			c++
		}
	}

	t.Logf("o=%v", mat)
	b := dctmat0(mat)
	for i, y := range b {
		for j, x := range y {
			b[i][j] = math.Round(x)
		}
	}
	t.Logf("b=%v", b)
	a := idct(b)
	if len(a) == 0 {
		t.Fatalf("a is empty")
	}
	for i, y := range a {
		for j, x := range y {
			if mat[i][j] != int16(math.Round(x)) {
				t.Errorf("mismatch a[%d][%d]", i, j)
			}
		}
	}
	t.Logf("a=%v", a)
}

func BenchmarkDCTMAT0(b *testing.B) {
	var c int16
	a := make([][]int16, 8)
	for i := 0; i < 8; i++ {
		a[i] = make([]int16, 8)
		for j := 0; j < 8; j++ {
			a[i][j] = c
			c++
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dctmat0(a)
	}
}

func matRound(m mat.Matrix) *mat.Dense {
	var ret mat.Dense
	ret.Apply(func(i, j int, v float64) float64 {
		return math.Round(v)
	}, m)
	return &ret
}

func TestIDCT_Value0(t *testing.T) {
	raw := mat.NewDense(8, 8, []float64{328, 0, 10, 0, 0, 0, 0, 0, 18, 0, 0, 0, 0, 0, 0, 0, -7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	exp := mat.NewDense(8, 8, []float64{45, 44, 42, 41, 41, 42, 44, 45, 45, 44, 43, 42, 42, 43, 44, 45, 45, 44, 43, 42, 42, 43, 44, 45, 44, 43, 42, 41, 41, 42, 43, 44, 43, 42, 41, 40, 40, 41, 42, 43, 41, 40, 39, 38, 38, 39, 40, 41, 40, 39, 37, 36, 36, 37, 39, 40, 38, 37, 36, 35, 35, 36, 37, 38})

	r := idct_(raw)
	act := matRound(r)

	if !mat.Equal(exp, act) {
		t.Logf("exp=%v", exp)
		t.Logf("act=%v", act)
		t.Logf("r=\n%v", mat.Formatted(r))
		t.Errorf("mismatch")
	}
}

func TestIDCT_Value150(t *testing.T) {
	raw := mat.NewDense(8, 8, []float64{288, -6, 10, 0, 0, 0, 0, 0, -6, 0, 0, 0, 0, 0, 0, 0, 7, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	exp := mat.NewDense(8, 8, []float64{38, 37, 36, 35, 34, 35, 36, 37, 37, 36, 35, 34, 34, 35, 37, 38, 35, 34, 33, 33, 34, 35, 37, 38, 34, 33, 32, 32, 34, 35, 38, 39, 34, 34, 33, 33, 34, 36, 38, 39, 36, 35, 34, 34, 35, 36, 38, 39, 39, 38, 36, 36, 36, 37, 38, 39, 40, 39, 38, 37, 36, 37, 38, 39})

	r := idct_(raw)
	act := matRound(r)

	if !mat.Equal(exp, act) {
		t.Logf("exp=%v", exp)
		t.Logf("act=%v", act)
		t.Logf("r=%v", r)
		t.Errorf("mismatch")
	}
}
