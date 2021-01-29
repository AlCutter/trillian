package client

import (
	"errors"
	"fmt"

	"github.com/google/trillian/experimental/logless/api"
	"github.com/google/trillian/merkle"
	"github.com/google/trillian/merkle/compact"
)

type GetTileFunc func(level, index, logSize uint64) (*api.Tile, error)

func InclusionProof(index, size uint64, f GetTileFunc) ([][]byte, error) {
	nodes, err := merkle.CalcInclusionProofNodeAddresses(int64(size), int64(index), int64(size))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate inclusion proof node list: %w", err)
	}

	nc := newNodeCache(f)
	ret := make([][]byte, 0)
	// TODO(al) parallelise
	for _, n := range nodes {
		h, err := nc.GetNode(n.ID, size)
		if err != nil {
			return nil, fmt.Errorf("failed to get node (%v): %w", n.ID, err)
		}
		ret = append(ret, h)
	}
	return ret, nil
}

func ConsistencyProof(smaller, larger uint64, f GetTileFunc) ([][]byte, error) {
	return nil, errors.New("unimpl")
}

// nodeCache hides the tiles abstraction away, and improves
// performance by caching tiles it's seen.
// Not threadsafe, and intended to be only used throughout the course
// of a single request.
type nodeCache struct {
	tiles   map[string]api.Tile
	getTile GetTileFunc
}

func newNodeCache(f GetTileFunc) nodeCache {
	return nodeCache{
		tiles:   make(map[string]api.Tile),
		getTile: f,
	}
}

func tileKey(l int, i uint64) string {
	return fmt.Sprintf("%d/%d", l, i)
}

func (n *nodeCache) GetNode(id compact.NodeID, logSize uint64) ([]byte, error) {
	tLevel, tIndex := id.Level/8, id.Index/256
	tKey := tileKey(int(tLevel), tIndex)
	t, ok := n.tiles[tKey]
	if !ok {
		tile, err := n.getTile(uint64(tLevel), tIndex, logSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tile: %w", err)
		}
		t = *tile
		n.tiles[tKey] = *tile
	}
	node := t.Nodes[api.TileNodeKey(id.Level%8, id.Index%256)]
	if node == nil {
		return nil, fmt.Errorf("node %v unknown", id)
	}
	return node, nil
}
