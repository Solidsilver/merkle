package mtree

import (
	"crypto/sha256"
)

func doHash(val []byte) []byte {
	sum := sha256.Sum256(val)
	return sum[:]
}

type MTree struct {
	Root *Node
}

func New() *MTree {
	return &MTree{}
}

func (b MTree) RootHash() []byte {
	if b.Root != nil {
		return b.Root.Hash
	}
	return []byte{}
}

// AddData inserts a new leaf node into the merkle tree
// with the hash of the given piece of data.
func (bt *MTree) AddData(val []byte) {
	hashVal := doHash(val)
	if bt.Root == nil {
		bt.Root = &Node{
			Hash:  hashVal,
			depth: 1,
		}
	} else if newP, ok := bt.Root.add(hashVal); ok {
		bt.Root = newP
	}
}

// Returns a depth-first string representation
// of the entire tree.
func (bt MTree) String() string {
	if bt.Root != nil {
		return bt.Root.sRec(0)
	}
	return "Empty Tree"
}
