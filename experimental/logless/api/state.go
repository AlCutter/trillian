package api

import "fmt"

// LogState represents the state of a logless log
type LogState struct {
	// Size is the number of leaves in the log
	Size uint64

	// SHA256 log root, RFC6962 flavour.
	RootHash []byte

	// Hashes are the roots of the minimal set of perfect subtrees contained by this log.
	Hashes [][]byte
}

type Tile struct {
	// RootHash is the hash of the root of this subtree tile
	RootHash []byte
	// Nodes stores the log tree nodes.
	// Keys are "<level>-<index>" where level is 0 ("leaf") through 6, and
	// index is 0 (left most node) through 2^(7-<level>).
	// Only non-ephemeral nodes are stored.
	Nodes map[string][]byte
}

// TileNodeKey generates keys used in Tile.Nodes map.
func TileNodeKey(level uint, index uint64) string {
	return fmt.Sprintf("%d-%d", level, index)
}
