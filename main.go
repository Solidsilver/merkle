package main

import (
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
	ver        = flag.String("v", "a", "do the thing")
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
	var controlTree *mtree.Tree
	var err error
	if *ver == "b" {
		controlTree, err = verify.HashFileHarr(*filePath, 1024)
	} else {
		controlTree, err = verify.HashFileLargeReadBuffer(*filePath, 1024)
	}
	if err != nil {
		fmt.Println("Error hashing file:", err.Error())
	}
	// rootHash := bt.RootHash()
	// fmt.Println()
	controlTree.TrimLeaves()
	tArr := controlTree.SerializeToArray()
	// fmt.Println("--- tArr ---")
	// for i := 0; i < len(tArr); i += 64 {
	// 	fmt.Println(base64.StdEncoding.EncodeToString(tArr[i:i+32]) + "|" + base64.StdEncoding.EncodeToString(tArr[i+32:i+64]))
	// }
	// fmt.Println("--- tArr end ---")
	treeFromArr, err := mtree.DeserializeFromArray(tArr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// fmt.Println(hex.EncodeToString(bt2.RootHash()))
	// fmt.Println(bt2.String())
	// fmt.Println()

	// hexHash := hex.EncodeToString(rootHash)
	// fmt.Printf("\nhash: %s\nfile: %s\n", hexHash, *filePath)
	fmt.Println("Comparison of trees:")
	// fmt.Println("Tree 1 (control)")
	// fmt.Println(bt.String())
	// fmt.Println()
	// fmt.Println("Tree 2 (fromArr)")
	// fmt.Println(bt2.String())
	// mtree.CompareTrees(bt, bt2)
	fmt.Printf("Trees are equal: %t", mtree.DeepEquals(controlTree, treeFromArr))
}
