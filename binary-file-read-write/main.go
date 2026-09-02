package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	text, _ := inputReader.ReadString('\n')

	file, err := os.Create("datum.bin")

	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	binaryData := []byte(text)
	binary.Write(file, binary.BigEndian, int32(len(text)))
	file.Write(binaryData)
	file.Close()
	fmt.Println("Text written to datum.bin successfully.")

	file, err = os.Open("datum.bin")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	var length int32
	binary.Read(file, binary.BigEndian, &length)
	data := make([]byte, length)
	file.Read(data)
	fmt.Println("Retrieved text:", string(data))

}
