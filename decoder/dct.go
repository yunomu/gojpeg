package decoder

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

var (
	dctA, dctAT mat.Matrix
)

func init() {
	var data []float64
	m0 := 1 / (2 * math.Sqrt(2))
	for i := 0; i < 8; i++ {
		data = append(data, m0)
	}

	for i := 1; i < 8; i++ {
		for j := 0; j < 8; j++ {
			data = append(data, 1/2.0*math.Cos(math.Pi*float64((2*j+1)*i)/16.0))
		}
	}

	dctA = mat.NewDense(8, 8, data)
	dctAT = dctA.T()
}

func toMatInt(s [][]int16) mat.Matrix {
	var data []float64
	for _, row := range s {
		for _, v := range row {
			data = append(data, float64(v))
		}
	}
	return mat.NewDense(8, 8, data)
}

func dct(s [][]int16) [][]float64 {
	x := toMatInt(s)

	var r mat.Dense
	r.Mul(dctA, x)
	r.Mul(&r, dctAT)

	var ret [][]float64
	for i := 0; i < 8; i++ {
		ret = append(ret, r.RawRowView(i))
	}

	return ret
}

func toMat(s [][]float64) mat.Matrix {
	var data []float64
	for _, row := range s {
		data = append(data, row...)
	}
	return mat.NewDense(8, 8, data)
}

func idct(b [][]float64) [][]float64 {
	x := toMat(b)

	r := idct_(x)

	var ret [][]float64
	for i := 0; i < 8; i++ {
		ret = append(ret, r.RawRowView(i))
	}

	return ret
}

func idct_(b mat.Matrix) *mat.Dense {
	var r mat.Dense
	r.Mul(dctAT, b)
	r.Mul(&r, dctA)

	return &r
}
