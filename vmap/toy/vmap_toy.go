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

	hasher := merkle.NewMapHasher(merkle.NewRFC6962TreeHasher(trillian.NewSHA256()))

	const batchSize = 1024
	const numBatches = 4
	rev := int64(0)
	var root trillian.Hash
	for x := 0; x < numBatches; x++ {
		w, err := merkle.NewSparseMerkleTreeWriter(rev, hasher,
			func() (storage.TreeTX, error) {
				return ms.Begin()
			})
		if err != nil {
			glog.Fatalf("Failed to create new SMTWriter: %v", err)
		}
		tx, err := ms.Begin()
		if err != nil {
			glog.Fatalf("Failed to Begin() a new tx: %v", err)
		}

		glog.Infof("Starting batch %d...", x)
		h := make([]merkle.HashKeyValue, batchSize)
		for y := 0; y < batchSize; y++ {
			h[y].HashedKey = hasher.HashKey([]byte(fmt.Sprintf("key-%d-%d", x, y)))
			h[y].HashedValue = hasher.TreeHasher.HashLeaf([]byte(fmt.Sprintf("value-%d-%d", x, y)))
			n := storage.NewNodeIDFromHash(h[y].HashedKey)
			glog.Infof("HashKey:\n%s\n%x", n.String(), n.Path)
		}
		glog.Infof("Created %d k/v pairs...", len(h))

		glog.Info("SetLeaves...")
		if err := w.SetLeaves(h); err != nil {
			glog.Fatalf("Failed to batch %d: %v", x, err)
		}
		glog.Info("SetLeaves done.")

		glog.Info("CalculateRoot...")
		root, err = w.CalculateRoot()
		if err != nil {
			glog.Fatalf("Failed to calculate root hash: %v", err)
		}
		glog.Infof("CalculateRoot (%d), root: %s", x, base64.StdEncoding.EncodeToString(root))

		err = tx.Commit()
		if err != nil {
			glog.Fatalf("Failed to Commit() tx: %v", err)
		}
		rev++
	}

	// calculated using python code.
	// const expectedRootB64 = "Av30xkERsepT6F/AgbZX3sp91TUmV1TKaXE6QPFfUZA=" // 4*1024
	// const expectedRootB64 = "6Pk5sprCr3ACfo0OLRZw7sAGdIBTc+7+MxfdW3n76Pc=" // 4*10
	// const expectedRootB64 = "QZJ42Te4bw+uGdUaIqzhelxpERU5Ru6uLdy0ixJAuWQ=" // 4*6
	// const expectedRootB64 = "9myL1k8Ik6m3Q3JXljHLzfNQHS2d5X6CCbpE/x3mixg=" // 4 * 4
	// const expectedRootB64 = "4xyGOe2DQYi2Qb4aBto9R7jSmiRYqfJ+TcMxUZTXMkM=" // 4 * 5
	// const expectedRootB64 = "FeB/9D+Gzo6oYB2Zi2JMHdrr9KvfvMk7o6DOzjPYG4w=" // 3 * 6
	const expectedRootB64 = "RfJ6JPERbkDiwlov8/alCqr4yeYYIWl3dWWS3trHsiY=" // 3 * 10
	// const expectedRootB64 = "pQhTahkoXM3WTeAO1o8BYKhgMNzS1yG03vg/fQSVyIc=" // 4*1
	// const expectedRootB64 = "RdcEkg5qEuW5eV3VJJLr6uSzvlc27D55AZChG76LHGA=" // 4*2
	// const expectedRootB64 = "3dpnVw5Le3HDq/GAkGoSYT9VkzJRV8z18huOk5qMbco=" // 1 * 4
	// const expectedRootB64 = "7R5uvGy5MJ2Y8xrQr4/mnn3aPw39vYscghmg9KBJaKc=" // 1 * 1024
	// const expectedRootB64 = "cZIYiv7ZQ/3rBfpCrha1NKdUnQ8NsTm21WWdV3P4qcU=" // 2 * 1
	// const expectedRootB64 = "KUaQinjLtPQ/ZAek4nHrR7tVXDxLt5QsvZK3vGopDkA=" // 3 * 1
	if expected, got := testonly.MustDecodeBase64(expectedRootB64), root; !bytes.Equal(expected, root) {
		glog.Fatalf("Expected root %s, got root: %s", base64.StdEncoding.EncodeToString(expected), base64.StdEncoding.EncodeToString(got))
	}
	glog.Infof("Finished, root: %s", base64.StdEncoding.EncodeToString(root))

}
