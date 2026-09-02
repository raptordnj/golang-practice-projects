package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`\b01[1-9]\d{8}\b`)
	result := re.FindAllString("01712345675 01712345674 01912345679 01912345674 00304195282 01304195282", -1)
	fmt.Println(result)
	// [01712345675 01912345679]

}
