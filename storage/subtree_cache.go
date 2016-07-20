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
	subtrees map[string]*SubtreeProto
	mutex    *sync.RWMutex
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
		subtrees: make(map[string]*SubtreeProto),
		mutex:    new(sync.RWMutex),
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
func splitNodeID(id NodeID) ([]byte, Suffix) {
	prefixSplit := (id.PrefixLenBits - 1) / strataDepth
	suffixEnd := (id.PrefixLenBits-1)/8 + 1
	s := Suffix{
		bits: byte((id.PrefixLenBits-1)%strataDepth) + 1,
		path: id.Path[prefixSplit:suffixEnd],
	}
	if id.PrefixLenBits%8 != 0 {
		suffixMask := byte(0x1<<uint((id.PrefixLenBits%8))) - 1
		s.path[len(s.path)-1] &= suffixMask
	}
	return id.Path[:prefixSplit], s
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
	suffixKey := string(px)
	c := s.subtrees[suffixKey]
	if c == nil {
		return fmt.Errorf("attempting to SetNodeHash for %s without having read any siblings", id.String())
	}
	c.Nodes[sx.serialize()] = &HashAndRevision{h, rev}
	return nil
}

func (s *SubtreeCache) Flush(setSubtree SetSubtreeFunc) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for k, v := range s.subtrees {
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
	return nil
}
