package main

import (
	"fmt"
	"net/http"
)

func main() {
	resp, err := http.Get("http://localhost")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body[:n]))
}
