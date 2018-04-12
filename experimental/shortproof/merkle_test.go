package shortproof

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/trillian/merkle"
)

func TestDiffieMerkleThing(t *testing.T) {
	leaves := [][]byte{
		[]byte("one"),
		[]byte("two"),
		[]byte("three"),
		[]byte("four"),
		[]byte("five"),
	}

	g := big.NewInt(5)
	p := big.NewInt(23)

	hasher := NewDiffieHasher(g, p)
	tree := merkle.NewInMemoryMerkleTree(hasher)

	t.Logf("root %x", tree.CurrentRoot())

	for _, l := range leaves {
		s, te, err := tree.AddLeaf(l)
		if err != nil {
			t.Fatalf("Couldn't add leaf: %v", err)
		}
		t.Logf("leaf %d = %x", s, te)
	}

	t.Logf("root %x", tree.CurrentRoot())
}

func TestDiffieMerkleInclusion(t *testing.T) {
	N := 5
	leaves := make([][]byte, 0, N)
	for i := 0; i < N; i++ {
		leaves = append(leaves, []byte(fmt.Sprintf("leaf %d", i)))
	}

	// openssl dhparam 256 | openssl asn1parse for g because apparently it should be a multiple of a large enough
	// to be sure it'll wrap around p when we raise it by node hashes.
	q, ok := big.NewInt(0).SetString("9D42CF01A6D486A9F92637162D9DA58C1A60F78BD758D95D1F53709155A00833", 16)
	if !ok {
		t.Fatalf("Failed to set q")
	}
	// some arbitrary multiple of q
	g := big.NewInt(0).Mul(big.NewInt(82347834587235234), q)
	// openssl dhparam 1024 | openssl asn1parse for p
	p, ok := big.NewInt(0).SetString("F2FE2D9FE5FA7FCC13C4E4ADE37F1F291F541DF969F638E75314BF45ABA6490644A9AFD90521FA7B00409C55B646BE50F0EB48951DB01851DE73D52CA618871E4C6EDDB0488DB5653CB5F04A019AD6FA249BF84D2B8F05A7D2C4BD57DA64F817F97EB5D56C2491264ED54952A4FDB0827592529F4F526DAE1B3F7B93C50320FB", 16)
	if !ok {
		t.Fatalf("Failed to set p")
	}

	hasher := NewDiffieHasher(g, p)
	tree := merkle.NewInMemoryMerkleTree(hasher)
	verifier := merkle.NewLogVerifier(hasher)

	for _, l := range leaves {
		s, te, err := tree.AddLeaf(l)
		if err != nil {
			t.Fatalf("Couldn't add leaf: %v", err)
		}
		t.Logf("leaf %d = %x", s, te)
	}

	hRootNode := sha256.Sum256(tree.CurrentRoot().Hash())
	rValue := big.NewInt(0).SetBytes(hRootNode[:])
	rootHash := sha256.Sum256(big.NewInt(0).Exp(g, rValue, p).Bytes())

	t.Logf("rootHash  %x", rootHash)

	for i := int64(0); i < int64(len(leaves)); i++ {
		t.Logf("Checking leaf %d", i)
		// "precalculate the mini proof:
		path := tree.PathToCurrentRoot(i + 1)
		proof := make([][]byte, len(path))

		miniMe := path[0].Value.Hash()
		proof[0] = miniMe
		for i := 1; i < len(path); i++ {
			miniMe = hasher.HashChildren(miniMe, path[i].Value.Hash())
			proof[i] = path[i].Value.Hash()

		}

		t.Logf("Proof (%3d bytes): %x", len(miniMe), miniMe)

		lh, err := hasher.HashLeaf(leaves[i])
		if err != nil {
			t.Errorf("%d: Failed to hash leaf: %v", i, err)
			continue
		}

		vRoot := sha256.Sum256(hasher.HashChildren(lh, miniMe))
		vRootValue := big.NewInt(0).SetBytes(vRoot[:])
		vRootHash := sha256.Sum256(big.NewInt(0).Exp(g, vRootValue, p).Bytes())
		//if want := root; vRoot.Cmp(want) != 0 {
		if !bytes.Equal(rootHash[:], vRootHash[:]) {
			t.Errorf("%d: got root %x want %x", i, vRootHash, rootHash)
		}

		// old way:
		if err := verifier.VerifyInclusionProof(i, tree.LeafCount(), proof, tree.CurrentRoot().Hash(), lh); err != nil {
			t.Errorf("%d: Failed to calculate root from proof: %v", i, err)
			continue
		}
	}
}
