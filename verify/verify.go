package verify

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/Solidsilver/merkle/hash"
	"github.com/Solidsilver/merkle/mtree"
	pb "github.com/schollz/progressbar/v3"
)

const GB_IN_BYTES = 1073741824

// HashFile hashes file using typical tree insertion
// It uses default file read buffer size
func HashFile(path string, splitSize int) (*mtree.Tree, error) {
	bt := mtree.NewEmpty()

	openFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer openFile.Close()
	stat, err := openFile.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()
	bar := pb.NewOptions64(fileSize,
		pb.OptionSetDescription("hashing"),
		pb.OptionShowBytes(true),
		pb.OptionShowElapsedTimeOnFinish(),
		pb.OptionSetPredictTime(true),
	)

	complete := false
	chunk := make([]byte, splitSize)
	reader := bufio.NewReaderSize(openFile, 8192)
	for !complete {
		bytesRead, err := io.ReadFull(reader, chunk)
		if err == io.EOF || bytesRead < splitSize {
			complete = true
		} else if err != nil && err != io.EOF {
			fmt.Println(err.Error())
			return nil, err
		}
		bt.AddData(chunk)
		bar.Add(bytesRead)
	}

	return bt, nil
}

// HashFileLargeReadBuffer hashes file using typical tree insertion
// It uses up to a 1G file read buffer size
func HashFileLargeReadBuffer(path string, splitSize int) (*mtree.Tree, error) {
	bt := mtree.NewEmpty()

	openFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer openFile.Close()
	stat, err := openFile.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()
	bar := pb.NewOptions64(fileSize,
		pb.OptionSetDescription("hashing"),
		pb.OptionShowBytes(true),
		pb.OptionShowElapsedTimeOnFinish(),
		pb.OptionSetPredictTime(true),
	)
	// Read in chunks of 1G at a time
	readSize := GB_IN_BYTES
	if fileSize < int64(readSize) {
		readSize = int(fileSize)
	}

	complete := false
	chunk := make([]byte, splitSize)
	reader := bufio.NewReaderSize(openFile, readSize)
	for !complete {
		bytesRead, err := io.ReadFull(reader, chunk)
		if err == io.EOF || bytesRead < splitSize {
			complete = true
		} else if err != nil && err != io.EOF {
			fmt.Println(err.Error())
			return nil, err
		}
		if bytesRead != 0 {
			bt.AddData(chunk)
			bar.Add(bytesRead)
		}
	}

	return bt, nil
}

// hashes a file by assembling a list of
// leaves, then building the tree from
// the leaves up. This uses up to a 1G
// file read buffer.
func HashFileHarr(path string, splitSize int) (*mtree.Tree, error) {
	openFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer openFile.Close()
	stat, err := openFile.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()
	harr := hash.NewHashArray(int(math.Ceil(float64(fileSize) / float64(splitSize))))
	bar := pb.NewOptions64(fileSize,
		pb.OptionSetDescription("hashing"),
		pb.OptionShowBytes(true),
		pb.OptionShowElapsedTimeOnFinish(),
		pb.OptionSetPredictTime(true),
	)
	readSize := GB_IN_BYTES
	if fileSize < int64(readSize) {
		readSize = int(fileSize)
	}
	reader := bufio.NewReaderSize(openFile, readSize)

	complete := false
	jobs := make(chan hash.HashJob, 100)
	var wg sync.WaitGroup

	workers := 3
	wg.Add(workers)
	for range workers {
		go hash.HashWorker(jobs, harr, &wg)
	}
	for !complete {
		chunk := make([]byte, splitSize)
		bytesRead, err := io.ReadFull(reader, chunk)
		if err == io.EOF || bytesRead < splitSize {
			complete = true
		} else if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		if bytesRead != 0 {
			harr.QueueHashInsert(chunk, jobs)
			bar.Add(bytesRead)
		}
	}
	close(jobs)
	wg.Wait()
	fmt.Println()
	bar = pb.NewOptions(-1,
		pb.OptionSetDescription("Building tree"),
	)
	isLoading := true
	go func() {
		for isLoading {
			bar.Add(1)
			time.Sleep(50 * time.Millisecond)
		}
		fmt.Println("Done Building Tree")
	}()
	bt := harr.BuildTree()
	isLoading = false

	return bt, nil
}

// HashFileCmp is a debug function
// to compare outputs of multiple
// hash techniques.
func HashFileCmp(path string, splitSize int) error {
	openFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer openFile.Close()
	stat, err := openFile.Stat()
	if err != nil {
		return err
	}
	iterBuiltTree := mtree.NewEmpty()
	fileSize := stat.Size()
	fmt.Printf("File size is %d bytes\n", fileSize)
	harrSize := int(math.Ceil(float64(fileSize) / float64(splitSize)))
	fmt.Printf("harrSize is %d\n", harrSize)
	harr := hash.NewHashArray(harrSize)
	bar := pb.NewOptions64(fileSize,
		pb.OptionSetDescription("hashing"),
		pb.OptionShowBytes(true),
		pb.OptionShowElapsedTimeOnFinish(),
		pb.OptionSetPredictTime(true),
	)
	readSize := GB_IN_BYTES
	if fileSize < int64(readSize) {
		readSize = int(fileSize)
	}
	reader := bufio.NewReaderSize(openFile, readSize)

	complete := false
	jobs := make(chan hash.HashJob, 100)
	var wg sync.WaitGroup

	workers := 3
	wg.Add(workers)
	for range workers {
		go hash.HashWorker(jobs, harr, &wg)
	}
	for !complete {
		chunk := make([]byte, splitSize)
		bytesRead, err := io.ReadFull(reader, chunk)
		if err == io.EOF || bytesRead < splitSize {
			complete = true
		} else if err != nil {
			fmt.Println(err.Error())
			return err
		}
		if bytesRead != 0 {
			harr.QueueHashInsert(chunk, jobs)
			iterBuiltTree.AddData(chunk)
			bar.Add(bytesRead)
		}
	}
	close(jobs)
	wg.Wait()
	harrBuiltTree := harr.BuildTree()
	mtree.CompareTrees(harrBuiltTree, iterBuiltTree)

	return nil
}
