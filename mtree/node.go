package mtree

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type Node struct {
	// Hash of the data for the given segment
	Hash  []byte
	left  *Node
	right *Node
	// Depth of the node (how many layers are below it). Used to speed up insertion.
	depth int
}

func NewNode(hash []byte, left, right *Node) Node {
	return Node{
		Hash:  hash,
		left:  left,
		right: right,
	}
}

func (n Node) isLeaf() bool {
	return n.left == nil && n.right == nil
}

// isBalanced returns the balance
// state of the current node
func (n *Node) isBalanced() bool {
	if n.depth != 0 {
		return true
	}
	depR, balR := n.right.depthBalance()
	if balR {
		depL, _ := n.left.depthBalance()
		if depL == depR {
			n.depth = 1 + depR
			return true
		}
	}
	return false
}

// depthBalance searches the nodes children
// to calculate the current depth and check if this node is balanced.
// This function will update the depth value for nodes as it searches.
func (n *Node) depthBalance() (int, bool) {
	if n.depth != 0 {
		return n.depth, true
	}
	depthR, balR := n.right.depthBalance()
	if balR {
		depthL, _ := n.left.depthBalance()
		if depthL == depthR {
			n.depth = 1 + depthR
			return n.depth, true
		}
	}
	return 0, false
}

// catHash concatenates the given hashes
// (padding if one is smaller) and returns
// a hash of the concatenated values.
func catHash(h1, h2 []byte) []byte {
	catHash := make([]byte, 64)
	copy(catHash[:32], h1)
	copy(catHash[32:], h2)
	return doHash(catHash)
}

// add inserts a leaf to the current node structure
// and returns if the insertion requires a new parent node.
func (n *Node) add(hashVal []byte) (newParent *Node, hasNewParent bool) {
	if n.isBalanced() {
		newParent = &Node{
			Hash: catHash(n.Hash, hashVal),
			left: n,
			right: &Node{
				Hash:  hashVal,
				depth: 1,
			},
		}
		if n.depth == 1 {
			newParent.depth = 2
		}
		return newParent, true
	}
	if newP, ok := n.right.add(hashVal); ok {
		n.right = newP
	}
	n.Hash = catHash(n.left.Hash, n.right.Hash)
	return nil, false
}

// sRec returns a depth-first string representation
// of the given node and it's children recursively
func (n Node) sRec(depth int) string {
	str := strings.Repeat("—", depth)
	if n.isLeaf() {
		str += "——Node with val [" + base64.StdEncoding.EncodeToString(n.Hash[:]) + "]"
	} else {
		str += "|-Node with val [" + base64.StdEncoding.EncodeToString(n.Hash[:]) + "]"
	}
	if n.left != nil {
		str += fmt.Sprintf("\n" + n.left.sRec(depth+1))
	}
	if n.right != nil {
		str += fmt.Sprintf("\n" + n.right.sRec(depth+1))
	}
	return str
}
