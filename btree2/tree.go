package btree2

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

func hashCopy(val, data []byte) {
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
		depthL, _ := n.Left.depthBalance()
		if depthL == depthR {
			n.depth = 1 + depthR
			return n.depth, true
		}
	}
	return 0, false
}

type BTree struct {
	Root       *Node
	nodeList   []Node
	curNodeIdx int
}

func New(chunks int) *BTree {
	return &BTree{
		nodeList: make([]Node, chunks),
	}
}

func (b BTree) RootHash() []byte {
	if b.Root != nil {
		return b.Root.Val
	}
	return []byte{}
}

type HashJob struct {
	data []byte
	idx  int
}

func HashWorker(jobs chan HashJob, bt *BTree, wg *sync.WaitGroup) {
	for hj := range jobs {
		bt.nodeList[hj.idx].Val = hash(hj.data)
	}
	wg.Done()
}

func (bt *BTree) AddData(val []byte, jobs chan HashJob) {
	jobs <- HashJob{
		data: val,
		idx:  bt.curNodeIdx,
	}

	bt.curNodeIdx++
}

func (bt *BTree) BuildTree() {
	start := time.Now()
	defer func() {
		fmt.Printf("Built tree in %s\n", time.Since(start).String())
	}()
	curLen := len(bt.nodeList)
	if curLen == 1 {
		bt.Root = &bt.nodeList[0]
		return
	}
	newLen := int(math.Ceil(float64(curLen) / 2))
	catHash := make([]byte, 64)
	for newLen > 1 {
		for i := 0; i < curLen-1; i += 2 {
			nL := bt.nodeList[i]
			nR := bt.nodeList[i+1]
			copy(catHash, nL.Val)
			copy(catHash[32:], nR.Val)
			parent := Node{
				Val:   hash(catHash),
				Left:  &nL,
				Right: &nR,
			}
			bt.nodeList[i/2] = parent
		}
		if newLen%2 == 1 {
			bt.nodeList[newLen-1] = bt.nodeList[curLen-1]
		}
		bt.nodeList = bt.nodeList[:newLen]
		curLen = newLen
		newLen = int(math.Ceil(float64(newLen) / 2))
	}
	nL := bt.nodeList[0]
	nR := bt.nodeList[1]
	copy(catHash, nL.Val)
	copy(catHash[32:], nR.Val)
	bt.Root = &Node{
		Val:   hash(catHash),
		Left:  &nL,
		Right: &nR,
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
