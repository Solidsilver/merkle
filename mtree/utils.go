package mtree

import (
	"encoding/base64"
	"fmt"
	"slices"
	"strings"
)

func CompareTrees(t1, t2 *MTree) {
	// fmt.Println(cmpNodes(t1.Root, t2.Root, 0))
	fmt.Printf("Root hash matches: %t\n", slices.Equal(t1.Root.Hash, t2.Root.Hash))
}

func cmpNodes(n1, n2 *Node, depth int) string {
	str := strings.Repeat("-", depth)
	slicesEqual := slices.Equal(n1.Hash, n2.Hash)
	if slicesEqual {
		str += "[•]" + base64.StdEncoding.EncodeToString(n1.Hash[:])
	} else {
		str += "[x]" + base64.StdEncoding.EncodeToString(n1.Hash[:]) + " | " + base64.StdEncoding.EncodeToString(n2.Hash[:])
	}
	if n1.left != nil && n2.left != nil {
		str += fmt.Sprintf("\n" + cmpNodes(n1.left, n2.left, depth+1))
	} else if !slicesEqual {
		str += "(LBM)"
	}
	if n1.right != nil && n2.right != nil {
		str += fmt.Sprintf("\n" + cmpNodes(n1.right, n2.right, depth+1))
	} else if !slicesEqual {
		str += "(RBM)"
	}
	return str
}
