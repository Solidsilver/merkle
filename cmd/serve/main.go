package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Solidsilver/merkle/hash"
	"github.com/Solidsilver/merkle/mtree"
)

var port = 8039

func main() {
	pathFlag := flag.String("f", "", "Select file to serve")
	flag.Parse()
	if *pathFlag == "" {
		log.Fatal("You must pass a file to serve `<cmd> -f <file>`")
	}
	openFile, err := os.Open(*pathFlag)
	if err != nil {
		log.Fatal("Failed to open file:", err.Error())
	}
	defer openFile.Close()

	inMemFile, err := io.ReadAll(openFile)
	if err != nil {
		log.Fatal("Failed to read in file", err.Error())
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /getFileChunk", func(respW http.ResponseWriter, req *http.Request) {
		rangeHeader := req.Header.Get("Range")
		if rangeHeader == "" {
			http.Error(respW, "No range header", http.StatusInternalServerError)
		}
		fileRange, err := parseRangeHeader(rangeHeader, len(inMemFile))
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Fetching range %d-%d\n", fileRange.start, fileRange.end)
		retBytes := inMemFile[fileRange.start:fileRange.end]
		fmt.Println("sending bytes")
		respW.Write(retBytes)
	})

	router.HandleFunc("GET /getHash", func(respW http.ResponseWriter, req *http.Request) {
		tree, err := hashBytes(inMemFile)
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}

		tree.TrimLeaves()
		tArr := tree.ToArray()
		fmt.Println("--- tArr ---")
		for i := 0; i < len(tArr); i += 64 {
			fmt.Println(base64.StdEncoding.EncodeToString(tArr[i:i+32]) + "|" + base64.StdEncoding.EncodeToString(tArr[i+32:i+64]))
		}
		fmt.Println("--- tArr end ---")
		// w.Header().Add("Content-Type", "application/octet-stream")

		respW.Write(tArr)
	})

	router.HandleFunc("GET /fileInfo/{id}", func(respW http.ResponseWriter, req *http.Request) {
		fileId := req.PathValue("id")
		fmt.Println("Fetching file info for", fileId)
	})
	fmt.Printf("Serving on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), makeGzipHandler(router))
	fmt.Println("Server closed")
}

type Range struct {
	start int
	end   int
}

func parseRangeHeader(header string, maxLen int) (retRng Range, err error) {
	if header == "" {
		return retRng, errors.New("no range header")
	}
	rngString := strings.Split(header, "-")
	if len(rngString) != 2 {
		return retRng, errors.Join(errors.New("failed to parse range header"), err)
	}
	retRng.start, err = strconv.Atoi(rngString[0])
	if err != nil {
		return retRng, errors.Join(errors.New("failed to parse range header"), err)
	}
	retRng.end, err = strconv.Atoi(rngString[1])
	if err != nil {
		return retRng, errors.Join(errors.New("failed to parse range header"), err)
	}

	if retRng.end > maxLen {
		return retRng, fmt.Errorf("range header out of bounds (requested %d in file of size %d)", retRng.end, maxLen)
	}
	return retRng, nil
}

func hashBytes(data []byte) (*mtree.Tree, error) {
	splitSize := 1024
	harr := hash.NewHashArray(int(math.Ceil(float64(len(data)) / float64(splitSize))))
	jobs := make(chan hash.HashJob, 100)
	var wg sync.WaitGroup

	workers := 3
	wg.Add(workers)
	for range workers {
		go hash.HashWorker(jobs, harr, &wg)
	}
	reader := bytes.NewReader(data)
	complete := false
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
		}
	}
	close(jobs)
	wg.Wait()
	fmt.Println()
	bt := harr.BuildTree()

	return bt, nil
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn.ServeHTTP(gzr, r)
	})
}
