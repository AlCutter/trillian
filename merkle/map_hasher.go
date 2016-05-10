package merkle

import (
	"github.com/google/trillian"
)

type MapHasher struct {
	TreeHasher
	nullHashes []trillian.Hash
}

func NewMapHasher(hasher trillian.Hasher) MapHasher {
	th := NewTreeHasher(hasher)
	return MapHasher{
		TreeHasher: th,
		nullHashes: createNullHashes(th),
	}
}

func createNullHashes(th TreeHasher) []trillian.Hash {
	numEntries := th.Size * 8
	r := make([]trillian.Hash, numEntries, numEntries)
	r[numEntries-1] = th.HashLeaf([]byte{})
	for i := numEntries - 2; i >= 0; i-- {
		r[i] = th.HashChildren(r[i+1], r[i+1])
	}
	return r
}
