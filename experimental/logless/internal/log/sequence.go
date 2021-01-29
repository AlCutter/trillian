package log

import (
	"fmt"

	"github.com/google/trillian/merkle/hashers"
)

type Entry struct {
	LeafData []byte
	LeafHash []byte
}

func Sequence(st LogStorage, h hashers.LogHasher, toAdd <-chan Entry) error {
	for entry := range toAdd {
		// ask storage to sequence
		if err := st.Sequence(entry.LeafHash, entry.LeafData); err != nil {
			return fmt.Errorf("failed to sequence %q: %q", entry.LeafHash, err)
		}
	}
	return nil
}
