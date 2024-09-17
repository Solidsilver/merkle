package mtree

import (
	"encoding/base64"
	"fmt"
	"slices"
	"strings"
)

func CompareTrees(t1, t2 *Tree) {
	fmt.Println("Comparison of trees:")
	fmt.Println(cmpNodes(t1.Root, t2.Root, 0))
	fmt.Printf("Root hash matches: %t\n", slices.Equal(t1.Root.Val, t2.Root.Val))
}

func cmpNodes(n1, n2 *Node, depth int) string {
	str := strings.Repeat("-", depth)
	slicesEqual := slices.Equal(n1.Val, n2.Val)
	if slicesEqual {
		str += "[•]" + base64.StdEncoding.EncodeToString(n1.Val[:])
	} else {
		str += "[x]" + base64.StdEncoding.EncodeToString(n1.Val[:]) + " | " + base64.StdEncoding.EncodeToString(n2.Val[:])
	}
	if n1.Left != nil && n2.Left != nil {
		str += fmt.Sprintf("\n" + cmpNodes(n1.Left, n2.Left, depth+1))
	} else if !slicesEqual {
		str += "(LBM)"
	}
	if n1.Right != nil && n2.Right != nil {
		str += fmt.Sprintf("\n" + cmpNodes(n1.Right, n2.Right, depth+1))
	} else if !slicesEqual {
		str += "(RBM)"
	}
	return str
}

// DeepEquals compates two trees starting at the root node, and recursively comparing all children.
// This only needs to be used for testing or tree manipulation.
// Otherwise, comparing root nodes is sufficient.
func DeepEquals(t1, t2 *Tree) bool {
	if t1.Root != nil && t2.Root != nil {
		return t1.Root.deepEqual(t2.Root)
	}

	return false
}
