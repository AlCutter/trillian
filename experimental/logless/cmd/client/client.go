package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/google/trillian/experimental/logless/api"
	"github.com/google/trillian/experimental/logless/internal/client"
	"github.com/google/trillian/experimental/logless/internal/storage/fs"
	"github.com/google/trillian/merkle/logverifier"
	"github.com/google/trillian/merkle/rfc6962/hasher"
)

var (
	storageDir   = flag.String("storage_dir", "", "Root directory to store log data.")
	fromIndex    = flag.Uint64("from_index", 0, "Index for inclusion proof")
	forEntryPath = flag.String("for_entry", "", "Path to entry for inclusion proof")
)

func main() {
	flag.Parse()
	// init storage
	st, err := fs.New(*storageDir)
	if err != nil {
		glog.Exitf("Failed to load storage: %q", err)
	}

	args := flag.Args()
	if len(args) == 0 {
		glog.Exit("Please specify a command from [inclusion]")
	}
	switch args[0] {
	case "inclusion":
		err = inclusionProof(st.LogState(), st.GetTile, args[1:])
	}
	if err != nil {
		glog.Exitf("Command %q failed: %q", args[0], err)
	}
}

func inclusionProof(state api.LogState, f client.GetTileFunc, args []string) error {
	entry, err := ioutil.ReadFile(*forEntryPath)
	if err != nil {
		return fmt.Errorf("failed to read entry from %q: %q", *forEntryPath, err)
	}
	lh := hasher.DefaultHasher.HashLeaf(entry)

	proof, err := client.InclusionProof(*fromIndex, state.Size, f)
	if err != nil {
		return fmt.Errorf("failed to get inclusion proof: %w", err)
	}

	lv := logverifier.New(hasher.DefaultHasher)
	if err := lv.VerifyInclusionProof(int64(*fromIndex), int64(state.Size), proof, state.RootHash, lh); err != nil {
		return fmt.Errorf("failed to verify inclusion proof: %q", err)
	}

	glog.Info("Inclusion verified.")
	return nil
}
