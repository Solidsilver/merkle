package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/Solidsilver/merkle/mtree"
)

func main() {
	client := new(http.Client)

	request, err := http.NewRequest("GET", "http://localhost:8039/getHash", nil)
	if err != nil {
		fmt.Println("Got err:", err.Error())
		return
	}
	// request.Header.Add("Accept-Encoding", "gzip")

	response, err := client.Do(request)
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Got err:", err.Error())
		return
	}
	fmt.Println("Got hash:\n", base64.RawStdEncoding.EncodeToString(bytes))
	mt, err := mtree.DeserializeFromArray(bytes)
	if err != nil {
		fmt.Println("Error deserializing array:", err.Error())
		return
	}
	fmt.Println("Constructed tree:\n", mt.String())
}
