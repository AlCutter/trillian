package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/google/trillian/experimental/logless/internal/log"
	"github.com/google/trillian/experimental/logless/internal/storage/fs"
	"github.com/google/trillian/merkle/rfc6962/hasher"
)

var (
	storageDir = flag.String("storage_dir", "", "Root directory to store log data.")
)

func main() {
	flag.Parse()
	h := hasher.DefaultHasher
	// init storage
	st, err := fs.New(*storageDir)
	if err != nil {
		glog.Exitf("Failed to load storage: %q", err)
	}

	if err := log.Integrate(st, h); err != nil {
		glog.Exitf("Failed to integrate: %q", err)
	}
}
