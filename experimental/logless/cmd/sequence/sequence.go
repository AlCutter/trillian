package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/google/trillian/experimental/logless/internal/storage/fs"

	"github.com/golang/glog"
	"github.com/google/trillian/experimental/logless/internal/log"
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

	entries := make(chan log.Entry, 100)
	go func() {
		for _, fp := range toAdd {
			entry, err := ioutil.ReadFile(fp)
			if err != nil {
				glog.Exitf("Failed to read entry file %q: %q", fp, err)
			}
			entries <- log.Entry{
				LeafData: entry,
				LeafHash: h.HashLeaf(entry),
			}
		}
		close(entries)
	}()

	if err := log.Sequence(st, h, entries); err != nil {
		glog.Exitf("Failed to sequence entries: %q", err)
	}
}
