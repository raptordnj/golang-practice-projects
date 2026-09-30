package main

import (
	"constructor-function/classes"
	"fmt"
)

func main() {
	p1 := classes.NewPerson("Tohidul Islam Touhid", 36)
	fmt.Println(p1.Name)
	fmt.Println(p1.Age)

	p2 := classes.NewPerson("Arman Hossain", 28)
	fmt.Println(p2.Name)
	fmt.Println(p2.Age)
}
