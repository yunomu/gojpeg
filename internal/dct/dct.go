package dct

import (
	"math"

	"github.com/yunomu/jpeg/internal/types"
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

var unzig []int = []int{
	0, 1, 5, 6, 14, 15, 27, 28,
	2, 4, 7, 13, 16, 26, 29, 42,
	3, 8, 12, 17, 25, 30, 41, 43,
	9, 11, 18, 24, 31, 40, 44, 53,
	10, 19, 23, 32, 39, 45, 52, 54,
	20, 22, 33, 38, 46, 51, 55, 60,
	21, 34, 37, 47, 50, 56, 59, 61,
	35, 36, 48, 49, 57, 58, 62, 63,
}

func zzToMatrix(zz types.Block) *mat.Dense {
	var data [types.BlockSize]float64
	for i := range zz {
		data[i] = float64(zz[unzig[i]])
	}

	return mat.NewDense(8, 8, data[:])
}

func Idct(b types.Block) types.Block {
	var ret types.Block
	r := idct_(zzToMatrix(b))
	r.Apply(func(i, j int, v float64) float64 {
		ret[i*8+j] = int32(math.Round(v))
		return v
	}, r)

	return ret
}

func idct_(b mat.Matrix) *mat.Dense {
	var r mat.Dense
	r.Mul(dctAT, b)
	r.Mul(&r, dctA)

	return &r
}
