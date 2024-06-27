package btree

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

func hash2(val, data []byte) {
	sum := sha256.Sum256(val)
	copy(data, sum[:])
}

func hash(val []byte) []byte {
	sum := sha256.Sum256(val)
	return sum[:]
}

type Node struct {
	Val   []byte
	Left  *Node
	Right *Node
	depth int
}

func (n Node) isLeaf() bool {
	return n.Left == nil && n.Right == nil
}

func (n *Node) IsBalanced() bool {
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

func (n *Node) depthBalance() (int, bool) {
	if n.depth != 0 {
		return n.depth, true
	}
	depthR, balR := n.Right.depthBalance()
	if balR {
		// n.Right.isBalanced = true
		// n.Right.depth = depthR
		depthL, _ := n.Left.depthBalance()
		// n.Left.depth = depthL
		// n.Left.isBalanced = true
		if depthL == depthR {
			n.depth = 1 + depthR
			return n.depth, true
		}
	}
	return 0, false
}

type BTree struct {
	Root   *Node
	tmpBoi []byte
}

func New() *BTree {
	return &BTree{
		tmpBoi: make([]byte, 64),
	}
}

func (b BTree) RootHash() []byte {
	if b.Root != nil {
		return b.Root.Val
	}
	return []byte{}
}

func (n *Node) add(data []byte) (newParent *Node, hasNewParent bool) {
	if n.IsBalanced() {
		hashCat := append(n.Val, data...)
		// copy(hashCat[:32], n.Val)
		// hashCat := make([]byte, 64)
		// copy(hashCat[32:], data)
		newParent = &Node{
			Val:  hash(hashCat),
			Left: n,
			Right: &Node{
				Val:   data,
				depth: 1,
			},
		}
		if n.depth == 1 {
			newParent.depth = 2
		}
		return newParent, true
	}
	if newP, ok := n.Right.add(data); ok {
		n.Right = newP
		n.Val = hash(append(n.Left.Val, n.Right.Val...))
	}
	return nil, false

}

func (bt *BTree) AddData(val []byte) {
	data := hash(val)
	if bt.Root == nil {
		bt.Root = &Node{
			Val:   data,
			depth: 1,
		}
	} else if newP, ok := bt.Root.add(data); ok {
		bt.Root = newP
	}
}

func (bt *BTree) AddDataIter(val []byte) {
	data := make([]byte, 32)
	hash2(val, data)
	if bt.Root == nil {
		bt.Root = &Node{
			Val:   data,
			depth: 1,
		}
		return
	}

	if bt.Root.IsBalanced() {
		copy(bt.tmpBoi, bt.Root.Val)
		copy(bt.tmpBoi[32:], data)
		newParent := &Node{
			Val:  hash(bt.tmpBoi),
			Left: bt.Root,
			Right: &Node{
				Val:   data,
				depth: 1,
			},
		}
		if bt.Root.depth == 1 {
			newParent.depth = 2
		}
		bt.Root = newParent
		return
	}

	prevNode := bt.Root
	curNode := bt.Root.Right
	for {
		if curNode.IsBalanced() {
			copy(bt.tmpBoi, bt.Root.Val)
			copy(bt.tmpBoi[32:], data)
			prevNode.Right = &Node{
				Val:  hash(bt.tmpBoi),
				Left: curNode,
				Right: &Node{
					Val:   data,
					depth: 1,
				},
			}
			prevNode.Val = hash(append(prevNode.Left.Val, prevNode.Right.Val...))
			if curNode.depth == 1 {
				prevNode.Right.depth = 2
			}
			return
		}
		prevNode = curNode
		curNode = curNode.Right
	}
}

func (n Node) sRec(depth int) string {
	str := strings.Repeat("—", depth)
	if n.isLeaf() {
		str += "——Node with val [" + base64.StdEncoding.EncodeToString(n.Val[:]) + "]"
	} else {
		str += "|-Node with val [" + base64.StdEncoding.EncodeToString(n.Val[:]) + "]"
	}
	if n.Left != nil {
		str += fmt.Sprintf("\n" + n.Left.sRec(depth+1))
	}
	if n.Right != nil {
		str += fmt.Sprintf("\n" + n.Right.sRec(depth+1))
	}
	return str

}

func (bt BTree) String() string {
	if bt.Root != nil {
		return bt.Root.sRec(0)
	}
	return "Empty Tree"
}
