package verify

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/Solidsilver/merkle/btree"
	"github.com/Solidsilver/merkle/btree2"
	pb "github.com/schollz/progressbar/v3"
)

const GB_IN_BYTES = 1073741824

func HashBytes(data []byte, splitSize int) *btree.BTree {
	size := len(data)
	bt := btree.New()
	fmt.Printf("Data size is %d\n", size)
	if size < splitSize {
		bt.AddData(data)
		return bt
	}
	for i := 0; i < size-splitSize; i += splitSize {
		part := data[i : i+splitSize]
		fmt.Printf("Part [%d-%d]\n", i, i+splitSize-1)
		bt.AddData(part)
	}
	lastPart := data[size-splitSize:]
	fmt.Printf("Part [%d-%d]\n", size-splitSize, size-1)
	bt.AddData(lastPart)
	return bt
}

func HashFile(path string, splitSize int) (*btree.BTree, error) {
	bt := btree.New()

	openFile, err := os.Open(path)
	defer openFile.Close()
	if err != nil {
		return nil, err
	}

	complete := false
	chunk := make([]byte, splitSize)
	reader := bufio.NewReaderSize(openFile, 8192)
	for !complete {
		// bytesRead, err := openFile.Read(chunk)
		bytesRead, err := io.ReadFull(reader, chunk)
		if err == io.EOF || bytesRead < splitSize {
			// for i := bytesRead; i < splitSize; i++ {
			// 	chunk[i] = 0
			// }
			complete = true
		} else if err != nil && err != io.EOF {
			fmt.Println(err.Error())
			return nil, err
		}
		// fmt.Printf("Read %d bytes\n", bytesRead)
		bt.AddData(chunk)
	}

	return bt, nil
}

func HashFileProgress(path string, splitSize int) (*btree.BTree, error) {
	bt := btree.New()

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
		// bytesRead, err := openFile.Read(chunk)
		bytesRead, err := io.ReadFull(reader, chunk)
		if err == io.EOF || bytesRead < splitSize {
			// for i := bytesRead; i < splitSize; i++ {
			// 	chunk[i] = 0
			// }
			complete = true
		} else if err != nil && err != io.EOF {
			fmt.Println(err.Error())
			return nil, err
		}
		// fmt.Printf("Read %d bytes\n", bytesRead)
		bt.AddData(chunk)
		// Increment Progress
		bar.Add(bytesRead)
	}

	return bt, nil
}

func HashFileProgress2(path string, splitSize int) (*btree.BTree, error) {
	bt := btree.New()

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
		bt.AddData(chunk)
		bar.Add(bytesRead)
	}

	return bt, nil
}

func HashFileProgress3(path string, splitSize int) (*btree.BTree, error) {
	start := time.Now()
	bt := btree.New()

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
		bt.AddDataIter(chunk)
		bar.Add(bytesRead)
	}
	end := time.Since(start)
	fmt.Println("Finished file hash in", end.String())

	return bt, nil
}

func HashFileProgress4(path string, splitSize int) (*btree2.BTree, error) {
	start := time.Now()

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
	bt := btree2.New(int(math.Ceil(float64(fileSize) / float64(splitSize))))
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

	complete := false
	reader := bufio.NewReaderSize(openFile, readSize)
	jobs := make(chan btree2.HashJob, 100)
	var wg sync.WaitGroup

	workers := 3
	wg.Add(workers)
	for range workers {
		go btree2.HashWorker(jobs, bt, &wg)
	}
	for !complete {
		chunk := make([]byte, splitSize)
		bytesRead, err := io.ReadFull(reader, chunk)
		if err == io.EOF || bytesRead < splitSize {
			complete = true
		} else if err != nil && err != io.EOF {
			fmt.Println(err.Error())
			return nil, err
		}
		bt.AddData(chunk, jobs)
		bar.Add(bytesRead)
	}
	close(jobs)
	wg.Wait()
	bt.BuildTree()
	end := time.Since(start)
	fmt.Println("Finished file hash in", end.String())

	return bt, nil
}
