package merkle

import (
	_ "bytes"
	"errors"
	"fmt"
	"github.com/google/trillian"
	"github.com/google/trillian/storage"
)

type SparseMerkleTreeReader struct {
	tx     storage.ReadOnlyTreeTX
	hasher MapHasher
}

type SparseMerkleTreeWriter struct {
	tx             storage.TreeTX
	hasher         MapHasher
	pendingKeys    map[string][]byte
	pendingVersion uint64
}

var (
	NoSuchRevision = errors.New("no such revision")
)

func (s SparseMerkleTreeReader) RootAtRevision(rev int64) (trillian.Hash, error) {
	nodes, err := s.tx.GetMerkleNodes(rev, []storage.NodeID{storage.NewEmptyNodeID(256)})
	if err != nil {
		return nil, err
	}
	switch {
	case len(nodes) == 0:
		return nil, NoSuchRevision
	case len(nodes) > 1:
		return nil, fmt.Errorf("expected 1 node, but got %d", len(nodes))
	}
	return nodes[0].Hash, nil
}

func (s SparseMerkleTreeReader) InclusionProof(rev int64, key trillian.Key) ([]trillian.Hash, error) {
	kh := s.hasher.keyHasher.Hash(key)
	nid := storage.NewNodeIDFromHash(kh)
	sibs := nid.Siblings()
	nodes, err := s.tx.GetMerkleNodes(rev, sibs)
	if err != nil {
		return nil, err
	}
	r := make([]trillian.Hash, len(sibs), len(sibs))
	for i := 0; i < len(r); i++ {
		if !sibs[i].Equivalent(nodes[i].NodeID) {
			panic(fmt.Errorf("expected node ID %v, but got %v", sibs[i].String(), nodes[i].NodeID.String()))
		}
		r[i] = nodes[i].Hash
	}
	return r, nil
}

func NewSparseMerkleTreeReader(h MapHasher, tx storage.ReadOnlyTreeTX) *SparseMerkleTreeReader {
	return &SparseMerkleTreeReader{
		tx:     tx,
		hasher: h,
	}
}

func NewSparseMerkleTreeWriter(h MapHasher, tx storage.TreeTX) *SparseMerkleTreeWriter {
	return &SparseMerkleTreeWriter{
		tx:          tx,
		hasher:      h,
		pendingKeys: make(map[string][]byte),
	}
}
