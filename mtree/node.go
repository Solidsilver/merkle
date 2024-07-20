package mtree

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type Node struct {
	// Val of the data for the given segment
	Val   []byte
	left  *Node
	right *Node
	// Caches the cache value so it doesn't need to be recomputed.
	// hashedVal []byte
	// Depth of the node (how many layers are below it). Used to speed up insertion.
	depth int
}

func NewNode(hash []byte, left, right *Node) Node {
	return Node{
		Val:   hash,
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

func (n Node) ComputeHash() []byte {
	if n.isLeaf() {
		return n.Val
	}
	return doHash(n.Val)
}

func (n Node) String() string {
	return base64.StdEncoding.EncodeToString(n.Val)
}

func cat(h1, h2 []byte) []byte {
	cat := make([]byte, 64)
	copy(cat[:32], h1)
	copy(cat[32:], h2)
	return cat
}

// add inserts a leaf to the current node structure
// and returns if the insertion requires a new parent node.
func (n *Node) add(hashVal []byte) (newParent *Node, hasNewParent bool) {
	if n.isBalanced() {
		newParent = &Node{
			Val:  cat(n.ComputeHash(), hashVal),
			left: n,
			right: &Node{
				Val:   hashVal,
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
	n.Val = cat(n.left.ComputeHash(), n.right.ComputeHash())
	return nil, false
}

// sRec returns a depth-first string representation
// of the given node and it's children recursively
func (n Node) sRec(depth int) string {
	str := strings.Repeat("—", depth)
	if n.isLeaf() {
		str += "——Node with val [" + base64.StdEncoding.EncodeToString(n.Val[:]) + "]"
	} else {
		str += "|-Node with val [" + base64.StdEncoding.EncodeToString(n.Val[:]) + "]"
	}
	if n.left != nil {
		str += fmt.Sprintf("\n" + n.left.sRec(depth+1))
	}
	if n.right != nil {
		str += fmt.Sprintf("\n" + n.right.sRec(depth+1))
	}
	return str
}

func (n Node) toArray(isRoot ...bool) [][]byte {
	arr := [][]byte{}
	if len(isRoot) > 0 {
		arr = append(arr, n.Val)
	}

	if !n.isLeaf() {
		arr = append(arr, n.left.Val, n.right.Val)
		arr = append(arr, n.left.toArray()...)
		arr = append(arr, n.right.toArray()...)
	}
	return arr
}

func (cur *Node) fromArray(arr [][]byte, curIdx int) int {
	if len(arr) > curIdx {
		cur.left = &Node{
			Val: arr[curIdx],
		}
		cur.right = &Node{
			Val: arr[curIdx+1],
		}
		curIdx += 2
		if len(cur.left.Val) == 64 {
			curIdx = cur.left.fromArray(arr, curIdx)
		}
		if len(cur.right.Val) == 64 {
			curIdx = cur.right.fromArray(arr, curIdx)
		}
		return curIdx
	}
	return curIdx
}
