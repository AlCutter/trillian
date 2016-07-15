package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"

	"github.com/golang/glog"
	"github.com/google/trillian"
	"github.com/google/trillian/merkle"
	"github.com/google/trillian/storage"
	"github.com/google/trillian/storage/mysql"
	"github.com/google/trillian/testonly"
)

var mysqlUriFlag = flag.String("mysql_uri", "test:zaphod@tcp(127.0.0.1:3306)/test", "")

func main() {
	flag.Parse()
	glog.Info("Starting...")
	ms, err := mysql.NewMapStorage(trillian.MapID{[]byte("TODO"), 1}, *mysqlUriFlag)
	if err != nil {
		glog.Fatalf("Failed to open mysql storage: %v", err)
	}

	tx, err := ms.Begin()
	if err != nil {
		glog.Fatalf("Failed to Begin() a new tx: %v", err)
	}

	hasher := merkle.NewMapHasher(merkle.NewRFC6962TreeHasher(trillian.NewSHA256()))
	w, err := merkle.NewSparseMerkleTreeWriter(100, hasher,
		func() (storage.TreeTX, error) {
			return ms.Begin()
		})
	if err != nil {
		glog.Fatalf("Failed to create new SMTWriter: %v", err)
	}

	const batchSize = 1024
	const numBatches = 4
	for x := 0; x < numBatches; x++ {
		glog.Infof("Starting batch %d...", x)
		h := make([]merkle.HashKeyValue, batchSize)
		for y := 0; y < batchSize; y++ {
			h[y].HashedKey = hasher.HashKey([]byte(fmt.Sprintf("key-%d-%d", x, y)))
			h[y].HashedValue = hasher.TreeHasher.HashLeaf([]byte(fmt.Sprintf("value-%d-%d", x, y)))
		}
		glog.Infof("Created %d k/v pairs...", len(h))

		glog.Info("SetLeaves...")
		if err := w.SetLeaves(h); err != nil {
			glog.Fatalf("Failed to batch %d: %v", x, err)
		}
		glog.Info("SetLeaves done.")

	}

	glog.Info("CalculateRoot...")
	root, err := w.CalculateRoot()
	if err != nil {
		glog.Fatalf("Failed to calculate root hash: %v", err)
	}
	glog.Info("CalculateRoot done.")

	err = tx.Commit()
	if err != nil {
		glog.Fatalf("Failed to Begin() a new tx: %v", err)
	}

	// calculated using python code.
	const expectedRootB64 = "Av30xkERsepT6F/AgbZX3sp91TUmV1TKaXE6QPFfUZA="
	if expected, got := testonly.MustDecodeBase64(expectedRootB64), root; !bytes.Equal(expected, root) {
		glog.Fatalf("Expected root %s, got root: %s", base64.StdEncoding.EncodeToString(expected), base64.StdEncoding.EncodeToString(got))
	}
	glog.Infof("Finished, root: %s", base64.StdEncoding.EncodeToString(root))

}
