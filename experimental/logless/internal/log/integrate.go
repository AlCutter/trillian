package log

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/google/trillian/experimental/logless/api"
	"github.com/google/trillian/merkle/compact"
	"github.com/google/trillian/merkle/hashers"
)

type LogStorage interface {
	GetTile(level, index, logSize uint64) (*api.Tile, error)
	StoreTile(level, index, tileSize uint64, tile *api.Tile) error
	LogState() api.LogState
	UpdateState(newState api.LogState) error
	ScanSequenced(begin uint64, f func(seq uint64, entry []byte) error) (uint64, error)
	Sequence(leafhash []byte, leaf []byte) error
}

// Integrate adds sequenced but not-yet-included entries into the tree state.
func Integrate(st LogStorage, h hashers.LogHasher) error {
	rf := compact.RangeFactory{h.HashChildren}

	// fetch state
	state := st.LogState()
	baseRange, err := rf.NewRange(0, state.Size, state.Hashes)
	if err != nil {
		return fmt.Errorf("failed to create range covering existing log: %q", err)
	}

	r, err := baseRange.GetRootHash(nil)
	if err != nil {
		return fmt.Errorf("invalid log state, unable to recalculate root: %q", err)
	}
	glog.Infof("Loaded state with roothash %x", r)

	tiles := make(map[string]*api.Tile)

	visitor := func(id compact.NodeID, hash []byte) {
		tileLevel := uint64(id.Level / 8)
		tileIndex := uint64(id.Index / 256)
		tileKey := tileKey(tileLevel, tileIndex)
		tile := tiles[tileKey]
		if tile == nil {
			tile, err = st.GetTile(tileLevel, tileIndex, state.Size)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				tile = &api.Tile{
					Nodes: make(map[string][]byte),
				}
			}
			tiles[tileKey] = tile
		}
		tile.Nodes[nodeKey(id.Level%8, id.Index%256)] = hash
	}

	// look for new sequenced entries and build tree
	newRange := rf.NewEmptyRange(state.Size)

	// write new completed subtrees
	n, err := st.ScanSequenced(state.Size,
		func(seq uint64, entry []byte) error {
			lh := h.HashLeaf(entry)
			glog.Infof("new @%d: %x", seq, lh)
			// Set leafhash on zeroth level
			visitor(compact.NodeID{Level: 0, Index: seq}, lh)
			// Update range and set internal nodes
			newRange.Append(lh, visitor)
			return nil
		})
	if err != nil {
		return fmt.Errorf("error while integrating: %q", err)
	}
	if n == 0 {
		glog.Infof("Nothing to do.")
		// Nothing to do, nothing done.
		return nil
	}

	if err := baseRange.AppendRange(newRange, visitor); err != nil {
		return fmt.Errorf("failed to merge new range onto existing log: %q", err)
	}

	newRoot, err := baseRange.GetRootHash(visitor)
	if err != nil {
		return fmt.Errorf("failed to calculate new root hash: %q", err)
	}
	glog.Infof("New log state: size %d hash: %x", baseRange.End(), newRoot)

	for k, t := range tiles {
		l, i := splitTileKey(k)
		s := tileSize(t)
		if err := st.StoreTile(l, i, s, t); err != nil {
			return fmt.Errorf("failed to store tile at level %d index %d: %q", l, i, err)
		}
	}

	newState := api.LogState{
		RootHash: newRoot,
		Size:     baseRange.End(),
		Hashes:   baseRange.Hashes(),
	}
	if err := st.UpdateState(newState); err != nil {
		return fmt.Errorf("failed to update stored state: %q", err)
	}
	return nil
}

func nodeKey(level uint, index uint64) string {
	return fmt.Sprintf("%d-%d", level, index)
}

func tileSize(t *api.Tile) uint64 {
	for i := uint64(0); i < 256; i++ {
		if t.Nodes[nodeKey(0, i)] == nil {
			return i
		}
	}
	return 256
}

func tileKey(level, index uint64) string {
	return fmt.Sprintf("%d/%d", level, index)
}

func splitTileKey(s string) (uint64, uint64) {
	p := strings.Split(s, "/")
	l, err := strconv.ParseUint(p[0], 10, 64)
	if err != nil {
		panic(err)
	}
	i, err := strconv.ParseUint(p[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return l, i

}
