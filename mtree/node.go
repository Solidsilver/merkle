package mtree

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

type Node struct {
	// Val of the data for the given segment
	Val   []byte
	Left  *Node
	Right *Node
	// Caches the cache value so it doesn't need to be recomputed.
	// hashedVal []byte
	// Depth of the node (how many layers are below it). Used to speed up insertion.
	depth int
}

func NewNode(val []byte, left, right *Node) Node {
	return Node{
		Val:   val,
		Left:  left,
		Right: right,
	}
}

func (n Node) IsLeaf() bool {
	return n.Left == nil && n.Right == nil
}

// isBalanced returns the balance
// state of the current node
func (n *Node) isBalanced() bool {
	if n.depth != 0 {
		return true
	}
	depR, balR := n.Right.depthBalance()
	if balR {
		depL, _ := n.Left.depthBalance()
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
	depthR, balR := n.Right.depthBalance()
	if balR {
		depthL, _ := n.Left.depthBalance()
		if depthL == depthR {
			n.depth = 1 + depthR
			return n.depth, true
		}
	}
	return 0, false
}

func (n Node) ComputeHash() []byte {
	if n.IsLeaf() {
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
			Left: n,
			Right: &Node{
				Val:   hashVal,
				depth: 1,
			},
		}
		if n.depth == 1 {
			newParent.depth = 2
		}
		return newParent, true
	}
	if newP, ok := n.Right.add(hashVal); ok {
		n.Right = newP
	}
	n.Val = cat(n.Left.ComputeHash(), n.Right.ComputeHash())
	return nil, false
}

// sRec returns a depth-first string representation
// of the given node and it's children recursively
func (n Node) sRec(depth int) string {
	str := strings.Repeat("—", depth)
	if n.IsLeaf() {
		str += "——Node with val [" + base64.StdEncoding.EncodeToString(n.Val[:]) + "]"
	} else {
		str += "|-Node with val [" + base64.StdEncoding.EncodeToString(n.Val[:32]) + "|" + base64.StdEncoding.EncodeToString(n.Val[32:]) + "]"
	}
	// if !n.IsLeaf() {
	// 	str += "|-Node with val [" + base64.StdEncoding.EncodeToString(n.Val[:]) + "]"
	// }
	if n.Left != nil {
		str += fmt.Sprintf("\n" + n.Left.sRec(depth+1))
	}
	if n.Right != nil {
		str += fmt.Sprintf("\n" + n.Right.sRec(depth+1))
	}
	return str
}

func (cur *Node) trimLeaves() {
	if cur.Left != nil {
		if cur.Left.IsLeaf() {
			cur.Left = nil
		} else {
			cur.Left.trimLeaves()
		}
	}
	if cur.Right != nil {
		if cur.Right.IsLeaf() {
			cur.Right = nil
		} else {
			cur.Right.trimLeaves()
		}
	}
}

func (n *Node) deepEqual(n2 *Node) bool {
	if !bytes.Equal(n.Val, n2.Val) {
		return false
	}
	leftMatch, rightMatch := true, true
	if n.Left != nil && n2.Left != nil {
		leftMatch = n.Left.deepEqual(n2.Left)
	}
	if leftMatch && n.Right != nil && n2.Right != nil {
		rightMatch = n.Right.deepEqual(n2.Right)
	}

	return rightMatch && leftMatch
}
