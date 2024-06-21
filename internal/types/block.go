package types

const BlockSize = 64

type Block [BlockSize]int32

var NilBlock = Block{}
