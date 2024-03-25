package huffman

import (
	"io"
	"sort"
)

type bin struct {
	b     byte
	count int
}

type binSlice []*bin

func (b binSlice) Len() int               { return len(b) }
func (b binSlice) Less(i int, j int) bool { return b[i].count < b[j].count }
func (b binSlice) Swap(i int, j int)      { b[i], b[j] = b[j], b[i] }

func histogram(r io.ByteReader) ([]*bin, error) {
	m := make(map[byte]int)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		c := m[b]
		m[b] = c + 1
	}

	var ret []*bin
	for b, c := range m {
		ret = append(ret, &bin{
			b:     b,
			count: c,
		})
	}

	sort.Sort(binSlice(ret))

	return ret, nil
}

type node struct {
	zero, one *node
	count     int
	b         byte
}

type nodeSlice []*node

func (n nodeSlice) Len() int               { return len(n) }
func (n nodeSlice) Less(i int, j int) bool { return n[i].count < n[j].count }
func (n nodeSlice) Swap(i int, j int)      { n[i], n[j] = n[j], n[i] }

func makeTree(bins []*bin) (*node, error) {
	if len(bins) == 0 {
		return nil, nil
	}

	var nodes []*node
	for _, bin := range bins {
		nodes = append(nodes, &node{
			count: bin.count,
			b:     bin.b,
		})
	}

	for {
		switch len(nodes) {
		case 0:
			panic("unexpected state")
		case 1:
			return nodes[0], nil
		default:
		}

		zero, one := nodes[0], nodes[1]
		node := &node{
			zero:  zero,
			one:   one,
			count: zero.count + one.count,
		}

		var newNodes []*node
		for i, n_ := range nodes[1:] {
			n := n_
			if node.count <= n.count {
				newNodes = append(newNodes, node, nodes[i+1:]...)
			}
		}

		var _ = node
	}
}

type Entry struct {
	orig byte
}

func Encode(r io.ByteReader) error {
	bins, err := histogram(r)
	if err != nil {
		return nil
	}

	var _ = bins
	panic("not implements")
}
