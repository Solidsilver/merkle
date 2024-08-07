package main

import (
	"encoding/hex"
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
	var bt *mtree.MTree
	var err error
	if *ver == "b" {
		bt, err = verify.HashFileHarr(*filePath, 2048)
	} else {
		bt, err = verify.HashFileLargeReadBuffer(*filePath, 2048)
	}
	if err != nil {
		fmt.Println("Error hashing file:", err.Error())
	}
	rootHash := bt.RootHash()
	hexHash := hex.EncodeToString(rootHash[:])
	fmt.Printf("\nhash: %s\nfile: %s\n", hexHash, *filePath)
}
