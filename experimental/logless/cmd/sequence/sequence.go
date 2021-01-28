package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/google/trillian/experimental/logless/internal/storage/fs"

	"github.com/golang/glog"
	"github.com/google/trillian/merkle/rfc6962/hasher"
)

var (
	storageDir = flag.String("storage_dir", "", "Root directory to store log data.")
	entries    = flag.String("entries", "", "File path glob of entries to add to the log.")
	create     = flag.Bool("create", false, "Set when creating a new log to initialise the structure.")
)

func main() {
	flag.Parse()

	h := hasher.DefaultHasher
	// init storage
	var st *fs.FS
	var err error
	if *create {
		st, err = fs.Create(*storageDir, h.EmptyRoot())
	} else {
		st, err = fs.New(*storageDir)
	}
	if err != nil {
		glog.Exitf("Failed to initialise storage: %q", err)
	}

	// sequence entries
	toAdd, err := filepath.Glob(*entries)
	if err != nil {
		glog.Exitf("Failed to glob entries %q: %q", *entries, err)
	}

	for _, fp := range toAdd {
		entry, err := ioutil.ReadFile(fp)
		if err != nil {
			glog.Exitf("Failed to read entry file %q: %q", fp, err)
		}
		lh := h.HashLeaf(entry)

		// ask storage to sequence
		err = st.Sequence(lh, entry)
		if err != nil {
			glog.Exitf("Failed to sequence %q: %q", fp, err)
		}
	}

	// done.
}
