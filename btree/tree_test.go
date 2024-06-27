package btree

import (
	"fmt"
	"testing"
)

func TestBtree(t *testing.T) {
	startStr := "abbaboneutc"
	mt := New()
	mt.Add(startStr)

	for range 10 {
		startStr += "123"
		mt.Add(startStr)
	}
	fmt.Println(mt.String())
}
