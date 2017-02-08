// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"github.com/google/trillian"
	"golang.org/x/net/context"
)

// ReadOnlyAdminTX is a transaction capable only of read operations in the
// AdminStorage.
type ReadOnlyAdminTX interface {
	AdminReader

	// Commit applies the operations performed to the underlying storage, or
	// returns an error.
	// A commit must be performed before any reads from storage are
	// considered consistent.
	Commit() error

	// Rollback aborts any performed operations, or returns an error.
	// See Close() for a way to automatically manage transactions.
	Rollback() error

	// IsClosed returns true if the transaction is closed.
	// A transaction is closed when either Commit() or Rollback() are
	// called.
	IsClosed() bool

	// Close rolls back the transaction if it's not yet closed.
	// It's advisable to call "defer tx.Close()" after the creation of
	// transaction to ensure that it's always rolled back if not explicitly
	// committed.
	Close() error
}

// AdminTX is a transaction capable of read and write operations in the
// AdminStorage.
type AdminTX interface {
	ReadOnlyAdminTX
	AdminWriter
}

// AdminStorage represents the persistent storage of tree data.
type AdminStorage interface {
	// Snapshot starts a read-only transaction.
	// A transaction must be explicitly committed before the data read by it
	// is considered consistent.
	Snapshot(ctx context.Context) (ReadOnlyAdminTX, error)

	// Begin starts a read/write transaction.
	// A transaction must be explicitly committed before the data read by it
	// is considered consistent.
	Begin(ctx context.Context) (AdminTX, error)
}

// AdminReader provides a read-only interface for tree data.
type AdminReader interface {
	// GetTree returns the tree corresponding to treeID or an error.
	GetTree(ctx context.Context, treeID int64) (*trillian.Tree, error)

	// ListTreeIDs returns the IDs of all trees in storage.
	// Note that there's no authorization restriction on the IDs returned,
	// so it should be used with caution in production code.
	ListTreeIDs(ctx context.Context) ([]int64, error)

	// ListTrees returns all trees in storage.
	// Note that there's no authorization restriction on the trees returned,
	// so it should be used with caution in production code.
	ListTrees(ctx context.Context) ([]trillian.Tree, error)
}

// AdminWriter provides a write-only interface for tree data.
type AdminWriter interface {
	// CreateTree inserts the specified tree in storage, returning a tree
	// with all storage-generated fields set.
	// Note that treeID and timestamps will be automatically generated by
	// the storage layer, thus may be ignored by the implementation.
	// Remaining fields must be set to valid values.
	// Returns an error if the tree is invalid or creation fails.
	CreateTree(ctx context.Context, tree *trillian.Tree) (*trillian.Tree, error)
}