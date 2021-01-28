package main

import (
	"flag"

	"github.com/google/trillian/merkle/rfc6962"
)

var (
	storageDir = flag.String("storage_dir", "", "Root directory to store log data.")
)

func main() {
	hasher := rfc6962.DefaultHasher
	// init storage
	st, err := fs.New(*storageDir)

	// fetch state

	// look for new sequenced entries and build tree

	// write new completed subtrees

	// create and sign tree head
	// write treehead

	// done.
}
