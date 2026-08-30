package main

import "fmt"

func main() {
	index := 0
	numbers := []int{1, 2, 3, 4, 5}

	// Optimized: Use copy instead of append to avoid variadic unpacking
	// and potential intermediate slice allocation
	copy(numbers[index:], numbers[index+1:])
	numbers = numbers[:len(numbers)-1]

	fmt.Println(numbers)
}
