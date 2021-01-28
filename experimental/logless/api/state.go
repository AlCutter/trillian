package api

// LogState represents the state of a logless log
type LogState struct {
	// Size is the number of leaves in the log
	Size uint64

	// SHA256 log root, RFC6962 flavour.
	RootHash []byte

	// Hashes are the roots of the minimal set of perfect subtrees contained by this log.
	Hashes [][]byte
}
