package shortproof

import (
	"crypto/sha256"
	"github.com/google/trillian/merkle/hashers"
	"math/big"
)

type diffieHasher struct {
	g, p *big.Int
}

var _ hashers.LogHasher = diffieHasher{}

func NewDiffieHasher(g, p *big.Int) hashers.LogHasher {
	return &diffieHasher{
		g: g,
		p: p,
	}
}

func (c diffieHasher) Size() int {
	return sha256.Size
}

func (c diffieHasher) EmptyRoot() []byte {
	eint := big.NewInt(0)
	z := big.NewInt(0).Exp(c.g, eint, c.p)
	zh := sha256.Sum256(z.Bytes())
	return zh[:]
}

func (c diffieHasher) HashLeaf(l []byte) ([]byte, error) {
	h := sha256.Sum256(append([]byte{0}, l...))
	return h[:], nil
}

func (c diffieHasher) HashChildren(l []byte, r []byte) []byte {
	lint := big.NewInt(0).SetBytes(l[:])
	rint := big.NewInt(0).SetBytes(r[:])
	e := big.NewInt(0).Mul(lint, rint)
	return e.Bytes()
}
