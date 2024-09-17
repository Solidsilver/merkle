package mtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

func doHash(val []byte) []byte {
	sum := sha256.Sum256(val)
	return sum[:]
}

// A representation of a Merkle Tree.
type Tree struct {
	Root *Node
}

// Creates a new empty Merkle tree.
func NewEmpty() *Tree {
	return &Tree{}
}

// Creates a new merkle tree with the given node.
func New(root *Node) *Tree {
	return &Tree{
		Root: root,
	}
}

// Returns the root hash of the merkle tree.
// If the root is nil, it will return an empty byte array.
func (b Tree) RootHash() []byte {
	if b.Root != nil {
		return b.Root.ComputeHash()
	}
	return []byte{}
}

// AddData inserts a new leaf node into the merkle tree
// with the hash of the given piece of data.
func (bt *Tree) AddData(val []byte) {
	hashVal := doHash(val)
	if bt.Root == nil {
		bt.Root = &Node{
			Val:   hashVal,
			depth: 1,
		}
	} else if newP, ok := bt.Root.add(hashVal); ok {
		bt.Root = newP
	}
}

// Returns a depth-first string representation
// of the entire tree.
func (bt Tree) String() string {
	if bt.Root != nil {
		return bt.Root.sRec(0)
	}
	return "Empty Tree"
}

var (
	chunkSize = 64
	nilMarker = make([]byte, 64)
)

// FromArray converts an array of bytes produced by
// [ToArray] back into a Merkle tree.
func FromArray(arr []byte) (*Tree, error) {
	if len(arr)%chunkSize != 0 || len(arr) == 0 {
		return nil, fmt.Errorf("Invalid array length, must be a multiple of 64 bytes, len(arr)=%d", len(arr)) // Invalid data length, must be a multiple of 64 bytes
	}
	newTree := New(&Node{
		Val: arr[:chunkSize],
	})

	queue := []*Node{newTree.Root}

	for byteIdx := chunkSize; byteIdx < len(arr); {
		curParent := queue[0]
		queue = queue[1:]

		leftData := arr[byteIdx : byteIdx+chunkSize]
		if !bytes.Equal(leftData, nilMarker) {
			leftNode := &Node{Val: leftData}
			curParent.Left = leftNode
			queue = append(queue, leftNode)
		}

		byteIdx += chunkSize
		if byteIdx < len(arr) {
			rightData := arr[byteIdx : byteIdx+chunkSize]
			if !bytes.Equal(rightData, nilMarker) {
				rightNode := &Node{Val: rightData}
				curParent.Right = rightNode
				queue = append(queue, rightNode)
			}
			byteIdx += chunkSize
		}

	}
	return newTree, nil
}

// ToArray serializes a merkle tree
// into an array-of-bytes representation.
// Use [FromArray] to convert back into a tree.
func (t Tree) ToArray() []byte {
	queue := []*Node{t.Root}
	arr := []byte{}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if cur == nil {
			arr = append(arr, nilMarker...)
		} else {
			arr = append(arr, cur.Val...)
			// left and right may be nil
			queue = append(queue, cur.Left, cur.Right)
		}
	}
	return arr
}

// TrimLeaves removes the bottom-most layer of the merkle tree.
// This is typically used prior to serialization, since the parents
// of the leaves contain the same information.
func (bt *Tree) TrimLeaves() {
	if bt.Root != nil {
		bt.Root.trimLeaves()
	}
}
