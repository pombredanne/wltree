// Package wltree provides an implementation of Wavelet Tree.
package wltree

import (
	"fmt"

	"github.com/mozu0/bitvector"
	"github.com/mozu0/huffman"
)

// Wltree represents a Wavelet Tree.
type Wltree struct {
	nodes [256][]*bitvector.BitVector
	codes [256]string
}

// New makes a Wavelet Tree from s.
func New(s []byte) *Wltree {
	w := &Wltree{}

	// Generate huffman tree based on character occurrences in s.
	charset, counts := freq(s)
	codes := huffman.FromInts(counts)
	for i, code := range codes {
		w.codes[charset[i]] = code
	}

	// Count number of bits in each node of the wavelet tree.
	sizes := make(map[string]int)
	for i := range charset {
		var (
			code  = codes[i]
			count = counts[i]
		)
		for j := 0; j < len(code)-1; j++ {
			sizes[code[:j]] += count
		}
	}

	// Assign BitVector Builders to each wavelet tree node.
	nodes := make(map[string]*node)
	for key, size := range sizes {
		nodes[key] = &node{
			builder: bitvector.NewBuilder(size),
		}
	}

	// Flip bits in each BitVector Builder in each wavelet tree node.
	for _, c := range s {
		code := w.codes[c]
		for j := 0; j < len(code)-1; j++ {
			n := nodes[code[:j]]
			if code[j] == '1' {
				n.builder.Set(n.index)
			}
			n.index++
		}
	}

	// Build all BitVectors.
	for key := range nodes {
		nodes[key].bv = nodes[key].builder.Build()
	}

	// For each charactor, register the path from wavelet tree root, through wavelet tree nodes, and
	// to the leaf.
	for i, c := range charset {
		code := codes[i]
		for j := 0; j < len(code)-1; j++ {
			w.nodes[c] = append(w.nodes[c], nodes[code[:j]].bv)
		}
	}

	return w
}

// Rank returns the count of the character c in s[0:i].
func (w *Wltree) Rank(c byte, i int) int {
	code := w.codes[c]
	if code == "" {
		return 0
	}

	nodes := w.nodes[c]
	for j := range nodes {
		if code[j] == '1' {
			i = nodes[j].Rank1(i)
		} else {
			i = nodes[j].Rank0(i)
		}
	}
	return i
}

// Select returns i such that Rank(c, i) = r, i.e. it returns the index of r-th occurrence of the
// character c.
func (w *Wltree) Select(c byte, r int) int {
	code := w.codes[c]
	if code == "" {
		panic(fmt.Sprintf("wltree: no such character '%v' in s.", c))
	}

	nodes := w.nodes[c]
	for j := len(nodes) - 1; j >= 0; j-- {
		if code[j] == '1' {
			r = nodes[j].Select1(r)
		} else {
			r = nodes[j].Select0(r)
		}
	}
	return r
}

type node struct {
	builder *bitvector.Builder
	bv      *bitvector.BitVector
	index   int
}

func freq(s []byte) (charset []byte, counts []int) {
	freqs := make(map[byte]int)
	for _, c := range s {
		freqs[c]++
	}
	for c, w := range freqs {
		charset = append(charset, c)
		counts = append(counts, w)
	}
	return
}
