package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/Solidsilver/merkle/btree"
	"github.com/Solidsilver/merkle/verify"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var filePath = flag.String("f", "", "write cpu profile to file")
var ver = flag.String("v", "a", "do the thing")

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
	// go func() {
	// 	http.ListenAndServe(":8080", nil)
	// }()
	//..
	// fileBytes, err := os.ReadFile(path)
	// if err != nil {
	// 	fmt.Println("Error opening file:" + err.Error())
	// 	return
	// }
	// start := time.Now()

	// bt, err := verify.HashFileProgress(*filePath, 2048)
	var bt btree.MTree
	var err error
	if *ver == "b" {
		bt, err = verify.HashFileProgress4(*filePath, 2048)
	} else {
		bt, err = verify.HashFileProgress2(*filePath, 2048)

	}
	if err != nil {
		fmt.Println("Error hashing file:", err.Error())
	}
	// fmt.Println(bt.String())
	rootHash := bt.RootHash()
	fmt.Println("hash is:", base64.StdEncoding.EncodeToString(rootHash[:]))
	fmt.Printf("file: %s\n", *filePath)
	// fmt.Printf("Time to hash: %s\n", opTime.String())

	// time.Sleep(time.Minute * 1)
}
