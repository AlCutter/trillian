package merkle

import (
	_ "bytes"
	"errors"
	"fmt"
	"github.com/google/trillian"
	"github.com/google/trillian/storage"
)

type SparseMerkleTreeReader struct {
	tx           storage.ReadOnlyTreeTX
	hasher       MapHasher
	treeRevision int64
}

type SparseMerkleTreeWriter struct {
	tx           storage.TreeTX
	hasher       MapHasher
	treeRevision uint64
}

var (
	NoSuchRevision = errors.New("no such revision")
)

func NewSparseMerkleTreeReader(rev int64, h MapHasher, tx storage.ReadOnlyTreeTX) *SparseMerkleTreeReader {
	return &SparseMerkleTreeReader{
		tx:           tx,
		hasher:       h,
		treeRevision: rev,
	}
}

func NewSparseMerkleTreeWriter(rev int64, h MapHasher, tx storage.TreeTX) *SparseMerkleTreeWriter {
	return &SparseMerkleTreeWriter{
		tx:     tx,
		hasher: h,
	}
}

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
	kh := s.hasher.keyHasher(key)
	nid := storage.NewNodeIDFromHash(kh)
	sibs := nid.Siblings()
	nodes, err := s.tx.GetMerkleNodes(rev, sibs)
	if err != nil {
		return nil, err
	}
	r := make([]trillian.Hash, len(sibs), len(sibs))
	for i := 0; i < len(r); i++ {
		if !sibs[i].Equivalent(nodes[i].NodeID) {
			return nil, fmt.Errorf("expected node ID %v, but got %v", sibs[i].String(), nodes[i].NodeID.String())
		}
		r[i] = nodes[i].Hash
	}
	return r, nil
}

func (s *SparseMerkleTreeWriter) SetLeaves(newRevision int64, leaves []trillian.MapLeaf) (trillian.TreeRoot, error) {

	return trillian.TreeRoot{}, errors.New("unimplemented")
}
