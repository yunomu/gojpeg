package main

import (
	"flag"
	"fmt"

	"gonum.org/v1/gonum/mat"
)

func init() {
	flag.Parse()
}

func main() {
	a := mat.NewDense(8, 8, nil)
	fmt.Println(mat.Formatted(a))
	a.Pow(a, 0)
	fmt.Println(mat.Formatted(a))
	a.Apply(func(i, j int, v float64) float64 {
		return 128
	}, a)
	fmt.Println(mat.Formatted(a))
}
