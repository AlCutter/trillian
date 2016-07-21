package storage

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/google/trillian"
)

type GetSubtreeFunc func(id NodeID) (*SubtreeProto, error)
type SetSubtreeFunc func(s *SubtreeProto) error

type SubtreeCache struct {
	subtrees      map[string]*SubtreeProto
	dirtyPrefixes map[string]bool
	mutex         *sync.RWMutex
}

type Suffix struct {
	bits byte
	path []byte
}

func (s *Suffix) serialize() string {
	r := make([]byte, 1, len(s.path)+1)
	r[0] = s.bits
	r = append(r, s.path...)
	return string(r)
}

const (
	strataDepth = 8
)

// TODO(al): consider supporting different sized subtrees - for now everything's subtrees of 8 levels.
func NewSubtreeCache() SubtreeCache {
	return SubtreeCache{
		subtrees:      make(map[string]*SubtreeProto),
		dirtyPrefixes: make(map[string]bool),
		mutex:         new(sync.RWMutex),
	}
}

/*
func (s *subtreeCache) checkNodeIDMatches(id NodeID) error {
	subtreePrefixLen := len(subtree.Prefix) * 8
	if id.PrefixLenBits > subtreePrefixLen+s.subtree.Depth {
		return fmt.Errorf("node id %s does not lie within subtree", id.String())
	}
	if id.PrefixLenBits < subtreePrefixLen {
		return fmt.Errorf("node id %s is shorter than my subtree prefix length %d", id.String(), subtreePrefixLen)
	}
	if nodePrefix := id.path[:len(s.subtree.Prefix)]; !bytes.Equal(nodePrefix, s.subtree.Prefix) {
		return fmt.Errorf("node id prefix %v does not match subtree prefix %v", nodePrefix, subtreePrefix)
	}
}
*/

// splitNodeID breaks a NodeID out into its prefix and suffix parts.
// unless ID is 0 bits long, Suffix must always contain at least one bit.
func splitNodeID(id NodeID) ([]byte, Suffix) {
	if id.PrefixLenBits == 0 {
		return []byte{}, Suffix{bits: 0, path: []byte{}}
	}
	prefixSplit := (id.PrefixLenBits - 1) / strataDepth
	suffixEnd := (id.PrefixLenBits-1)/8 + 1
	s := Suffix{
		bits: byte((id.PrefixLenBits-1)%strataDepth) + 1,
		path: make([]byte, suffixEnd-prefixSplit),
	}
	// XXX necessary?
	copy(s.path, id.Path[prefixSplit:suffixEnd])
	if id.PrefixLenBits%8 != 0 {
		suffixMask := (byte(0x1<<uint((id.PrefixLenBits%8))) - 1) << uint(8-id.PrefixLenBits%8)
		s.path[len(s.path)-1] &= suffixMask
	}

	// XXX necessary?
	r := make([]byte, prefixSplit)
	copy(r, id.Path[:prefixSplit])
	//return id.Path[:prefixSplit], s
	return r, s
}

func (s *SubtreeCache) GetNodeHash(id NodeID, getSubtree GetSubtreeFunc) (trillian.Hash, int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	px, sx := splitNodeID(id)
	prefixKey := string(px)
	c := s.subtrees[prefixKey]
	if c == nil {
		subID := id
		subID.PrefixLenBits = len(px) * 8
		var err error
		c, err = getSubtree(subID)
		if err != nil {
			return nil, -1, err
		}
		if c == nil {
			// storage didn't have one for us, so we'll store an empty proto here
			// incase we try to update it later on (we won't flush it back to
			// storage if it's empty.)
			c = &SubtreeProto{
				Prefix: px,
				Depth:  strataDepth,
				Nodes:  make(map[string]*HashAndRevision),
			}
		}
		if c.Prefix == nil {
			panic(fmt.Errorf("GetNodeHash nil prefix on %v for id %v with px %#v", c, id.String(), px))
		}

		s.subtrees[prefixKey] = c
	}

	nh := c.Nodes[sx.serialize()]
	if nh == nil {
		return nil, -1, nil
	}
	return nh.Hash, nh.Revision, nil
}

func (s *SubtreeCache) SetNodeHash(id NodeID, rev int64, h trillian.Hash) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	px, sx := splitNodeID(id)
	prefixKey := string(px)
	c := s.subtrees[prefixKey]
	if c == nil {
		// XXX FIX ME
		//return fmt.Errorf("attempting to SetNodeHash for %s without having read any siblings", id.String())
		// This is of, IFF *all* leaves in the subtree are being set...
		c = &SubtreeProto{
			Prefix: px,
			Nodes:  make(map[string]*HashAndRevision),
		}
		s.subtrees[prefixKey] = c
	}
	if c.Prefix == nil {
		panic(fmt.Errorf("nil PREFIX for %v (key %v)", id.String(), prefixKey))
	}
	s.dirtyPrefixes[prefixKey] = true
	c.Nodes[sx.serialize()] = &HashAndRevision{h, rev}
	return nil
}

func (s *SubtreeCache) Flush(setSubtree SetSubtreeFunc) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for k, v := range s.subtrees {
		if s.dirtyPrefixes[k] {
			bk := []byte(k)
			if !bytes.Equal(bk, v.Prefix) {
				return fmt.Errorf("inconsistent cache: prefix key is %v, but cached object claims %v", bk, v.Prefix)
			}
			if len(v.Nodes) > 0 {
				if err := setSubtree(v); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
