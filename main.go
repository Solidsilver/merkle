package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/Solidsilver/merkle/mtree"
	"github.com/Solidsilver/merkle/verify"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	filePath   = flag.String("f", "", "write cpu profile to file")
	ver        = flag.String("v", "harr", "Specify file hashing strategy. Use 'old' for tree insertion strategy.")
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		fmt.Println("Saving CPU Prof to", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *filePath == "" {
		log.Fatal("You must include a file to hash using the -f parameter. Ex: go run main.go -f <filepath>")
	}
	if _, err := os.Stat(*filePath); errors.Is(err, os.ErrNotExist) {
		log.Fatal("File does not exist.")
		return
		// path/to/whatever does not exist
	}
	var controlTree *mtree.Tree
	var err error
	if *ver == "harr" {
		controlTree, err = verify.HashFileHarr(*filePath, 1024)
	} else {
		controlTree, err = verify.HashFileLargeReadBuffer(*filePath, 1024)
	}
	if err != nil {
		fmt.Println("Error hashing file:", err.Error())
	}
	// rootHash := bt.RootHash()
	// fmt.Println()
	// controlTree.TrimLeaves()
	// tArr := controlTree.ToArray()
	// treeFromArr, err := mtree.FromArray(tArr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// fmt.Println("Comparing trees...")
	// fmt.Printf("Trees are equal: %t", mtree.DeepEquals(controlTree, treeFromArr))
	// fmt.Printf("\nHash: %s\n", base64.RawStdEncoding.EncodeToString(controlTree.RootHash()))
	fmt.Printf("\nHash: %d\n", binary.LittleEndian.Uint64((controlTree.RootHash())))
	fmt.Printf("Hash is %d bytes\n", len(controlTree.RootHash()))
}
