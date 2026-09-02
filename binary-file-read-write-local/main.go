package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your text:")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	file, err := os.Create("data.bin")

	if err != nil {
		panic(err)
	}

	length := int32(len(text))
	binary.Write(file, binary.LittleEndian, length)
	file.Write([]byte(text))
	defer file.Close()
	fmt.Println("Write to data.bin as real binary file")

	file, err = os.Open("data.bin")

	if err != nil {
		panic(err)
	}

	var l int32
	binary.Read(file, binary.BigEndian, &l)
	buf := make([]byte, l)
	io.ReadFull(file, buf)
	fmt.Println(string(buf))
	file.Close()
}
