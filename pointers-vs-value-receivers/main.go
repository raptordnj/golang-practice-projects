package main

import (
	"fmt"
	"pointers-vs-value-receivers/classes"
)

func main() {
	counter := classes.Counter{}
	counter.Increment()
	counter.Increment()
	counter.Increment()
	counter.Increment()
	counter.Increment()
	counter.Increment()
	counter.Increment()
	counter.Increment()

	fmt.Println(counter.GetCount())

	counter.Reset()
	fmt.Println("After reset:")
	fmt.Println(counter.GetCount())
}
