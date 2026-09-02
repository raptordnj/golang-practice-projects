package main

import (
	"fmt"
)

func main() {
	bytes := []byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd'}
	binBytes := []byte{0b01001000, 0b01100101, 0b01101100, 0b01101100, 0b01101111, 0b00100000, 0b01010111, 0b01101111, 0b01110010, 0b01101100, 0b01100100}
	fmt.Println(string(bytes))
	fmt.Println(string(binBytes))
}
