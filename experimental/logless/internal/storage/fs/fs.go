package fs

import (
	"fmt"
	"os"
)

// FS is a logless storage implementation
type FS struct {
	rootDir string
}

// New creates a new FS storage entry.
func New(rootDir string) (*FS, error) {
	f, err := os.Open(rootDir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if !f.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", rootDir)
	}

	return &FS{
		rootDir: rootDir,
	}, nil
}

// Sequence assigns the given leafhash and entry to the next available sequence number.
func (fs *FS) Sequence(leafHash []byte, leaf []byte) (uint64, error) {
	return 0, nil
}
