package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/google/trillian/merkle/rfc6962"
)

var (
	storageDir = flag.String("storage_dir", "", "Root directory to store log data.")
	entries    = flag.String("entries", "", "File path glob of entries to add to the log.")
)

func main() {
	hasher := rfc6962.DefaultHasher
	// init storage
	st, err := fs.New(*storageDir)

	// sequence entries
	toAdd, err := filepath.Glob(*entries)
	if err != nil {
		glog.Fatalf("Failed to glob entries %q: %q", *entries, err)
	}

	for _, fp := range toAdd {
		entry, err := ioutil.ReadFile(fp)
		if err != nil {
			glog.Fatalf("Failed to read entry file %q: %q", fp, err)
		}
		lh := hasher.LeafHash(entry)

		// ask storage to sequence
		seq, err := st.Sequence(lh, entry)
		if err != nil {
			glog.Fatalf("Failed to sequence %q: %q", fp, err)
		}
	}

	// done.
}
