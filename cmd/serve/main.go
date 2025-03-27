package main

import (
	"compress/gzip"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Solidsilver/merkle/verify"
)

var port = 8039

func main() {
	pathFlag := flag.String("f", "", "Select file to serve")
	flag.Parse()
	if *pathFlag == "" {
		log.Fatal("You must pass a file to serve `<cmd> -f <file>`")
	}
	dir, err := os.ReadDir(*pathFlag)
	if err != nil {
		log.Fatal("Failed to open dir:", err.Error())
	}
	fileDirName := path.Dir(*pathFlag)
	fileList := []string{}
	for _, entry := range dir {
		if !entry.IsDir() {
			fileList = append(fileList, entry.Name())
		}
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /getFile/{fname}", func(respW http.ResponseWriter, req *http.Request) {
		rangeHeader := req.Header.Get("Range")
		reqFileName := req.PathValue("fname")
		if !slices.Contains(fileList, reqFileName) {
			http.Error(respW, "File does not exist: "+reqFileName, http.StatusNotFound)
			return
		}
		file, err := os.Open(path.Join(fileDirName, reqFileName))
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		fs, err := file.Stat()
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		fileRange, err := parseRangeHeader(rangeHeader, int(fs.Size()))
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		fileBytes := make([]byte, (fileRange.end+1)-fileRange.start)
		_, err = file.Seek(int64(fileRange.start), 0)
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = file.Read(fileBytes)
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Fetching range %d-%d\n", fileRange.start, fileRange.end)
		fmt.Println("sending bytes")
		respW.Write(fileBytes)
	})

	router.HandleFunc("GET /getMerkle/{id}", func(respW http.ResponseWriter, req *http.Request) {
		reqFileName := req.PathValue("id")
		if !slices.Contains(fileList, reqFileName) {
			http.Error(respW, "File does not exist: "+reqFileName, http.StatusNotFound)
			return
		}
		file, err := os.Open(path.Join(fileDirName, reqFileName))
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		tree, err := verify.HashFileHarr(path.Join(fileDirName, reqFileName), 1024)
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
		reqFileName := req.PathValue("id")
		fmt.Println("Fetching file info for", reqFileName)
		fmt.Printf("HEAD request for file %s\n", reqFileName)
		if !slices.Contains(fileList, reqFileName) {
			http.Error(respW, "File does not exist: "+reqFileName, http.StatusNotFound)
			return
		}
		file, err := os.Open(path.Join(fileDirName, reqFileName))
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		fs, err := file.Stat()
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(respW, `{ "size": "%d" }`, fs.Size())
	})

	router.HandleFunc("HEAD /getFile/{fname}", func(respW http.ResponseWriter, req *http.Request) {
		reqFileName := req.PathValue("fname")
		fmt.Printf("HEAD request for file %s\n", reqFileName)
		if !slices.Contains(fileList, reqFileName) {
			http.Error(respW, "File does not exist: "+reqFileName, http.StatusNotFound)
			return
		}
		file, err := os.Open(path.Join(fileDirName, reqFileName))
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		fs, err := file.Stat()
		if err != nil {
			http.Error(respW, err.Error(), http.StatusInternalServerError)
			return
		}
		respW.Header().Set("Content-Type", "blob")
		respW.Header().Set("Date", time.Now().String())
		respW.Header().Set("Content-Length", fmt.Sprintf("%d", fs.Size()))
		respW.WriteHeader(200)
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
		retRng.start = 0
		retRng.end = maxLen - 1
		return retRng, nil
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

// func hashBytes(data []byte) (*mtree.Tree, error) {
// 	splitSize := 1024
// 	harr := hash.NewHashArray(int(math.Ceil(float64(len(data)) / float64(splitSize))))
// 	jobs := make(chan hash.HashJob, 100)
// 	var wg sync.WaitGroup
//
// 	workers := 3
// 	wg.Add(workers)
// 	for range workers {
// 		go hash.HashWorker(jobs, harr, &wg)
// 	}
// 	reader := bytes.NewReader(data)
// 	complete := false
// 	for !complete {
// 		chunk := make([]byte, splitSize)
// 		bytesRead, err := io.ReadFull(reader, chunk)
// 		if err == io.EOF || bytesRead < splitSize {
// 			complete = true
// 		} else if err != nil {
// 			fmt.Println(err.Error())
// 			return nil, err
// 		}
// 		if bytesRead != 0 {
// 			harr.QueueHashInsert(chunk, jobs)
// 		}
// 	}
// 	close(jobs)
// 	wg.Wait()
// 	fmt.Println()
// 	bt := harr.BuildTree()
//
// 	return bt, nil
// }

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
