package hash

import (
	"math"
	"sync"

	"github.com/Solidsilver/merkle/mtree"
)

// catHash concatenates the given hashes
// (padding if one is smaller) and returns
// a hash of the concatenated values.
func catHash(h1, h2 []byte) []byte {
	catHash := make([]byte, 64)
	copy(catHash[:32], h1)
	copy(catHash[32:], h2)
	return Do(catHash)
}

type HashArray struct {
	nodeList   []mtree.Node
	curNodeIdx int
}

func NewHashArray(chunks int) *HashArray {
	return &HashArray{
		nodeList: make([]mtree.Node, chunks),
	}
}

type HashJob struct {
	data []byte
	idx  int
}

// HashWorker pulls an available job off
// of the job queue, hashes the data,
// and inserts it at the proper location in the HashArray
func HashWorker(jobs chan HashJob, harr *HashArray, wg *sync.WaitGroup) {
	for hj := range jobs {
		harr.nodeList[hj.idx].Hash = Do(hj.data)
	}
	wg.Done()
}

func (harr *HashArray) QueueHashInsert(val []byte, jobs chan HashJob) {
	jobs <- HashJob{
		data: val,
		idx:  harr.curNodeIdx,
	}
	harr.curNodeIdx++
}

// BuildTree creates a MTree using
// an array of nodes as the leaves
// and building up from there
func (harr *HashArray) BuildTree() *mtree.MTree {
	bt := mtree.New()
	curLen := len(harr.nodeList)
	if curLen == 1 {
		bt.Root = &harr.nodeList[0]
		return bt
	}
	newLen := int(math.Ceil(float64(curLen) / 2))
	ch := make([]byte, 64)
	for newLen > 1 {
		for i := 0; i < curLen-1; i += 2 {
			nL := harr.nodeList[i]
			nR := harr.nodeList[i+1]
			copy(ch[:32], nL.Hash)
			copy(ch[32:], nR.Hash)
			parent := mtree.NewNode(Do(ch), &nL, &nR)
			harr.nodeList[i/2] = parent
		}
		if curLen%2 == 1 {
			harr.nodeList[newLen-1] = harr.nodeList[curLen-1]
		}
		harr.nodeList = harr.nodeList[:newLen]
		curLen = newLen
		newLen = int(math.Ceil(float64(newLen) / 2))
	}
	nL := harr.nodeList[0]
	nR := harr.nodeList[1]

	ch2 := catHash(nL.Hash, nR.Hash)
	btr := mtree.NewNode(ch2, &nL, &nR)
	bt.Root = &btr
	return bt
}
