/*
Package wltree provides an implementation of Wavelet Tree.
See http://en.wikipedia.org/wiki/Wavelet_Tree for details.

Example

    const s = "abracadabra"
    wt := wltree.New(s)
    // The number of 'a' in s.
    wt.Rank('a', len(s)) //=> 5
    // The number of 'a' in s[3:8] = "acada"
    wt.Rank('a', 8) - wt.Rank('a', 3) //=> 3
    // The index of the 3rd occurrence of 'a' in s. 0-origin, thus 2 means 3rd.
    wt.Select('a', 2) //=> 5
*/
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
		code := codes[i]
		count := counts[i]
		for j := range code {
			sizes[code[:j]] += count
		}
	}

	// Assign BitVector Builders to each wavelet tree node.
	builders := make(map[string]*bitvector.Builder)
	var keys []string
	for key, size := range sizes {
		keys = append(keys, key)
		builders[key] = bitvector.NewBuilder(size)
	}

	// Set bits in each BitVector Builder.
	index := make(map[string]int)
	for _, c := range s {
		code := w.codes[c]
		for j := range code {
			if code[j] == '1' {
				builders[code[:j]].Set(index[code[:j]])
			}
			index[code[:j]]++
		}
	}

	// Build all BitVectors.
	bvs := make(map[string]*bitvector.BitVector)
	for key, builder := range builders {
		bvs[key] = builder.Build()
	}

	// For each charactor, register the path from wavelet tree root, through wavelet tree nodes, and
	// to the leaf.
	for i, c := range charset {
		code := codes[i]
		for j := range code {
			w.nodes[c] = append(w.nodes[c], bvs[code[:j]])
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
