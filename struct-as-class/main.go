package main

import (
	"fmt"
	"struct-as-class/classes"
)

func main() {
	student := classes.Student{
		Person: classes.Person{
			Name: "Tohidul Islam Touhid",
			Age:  36,
		},
		School: "KBM College Dinajpur",
	}

	fmt.Println(student.Greet())
	fmt.Println(student.Study())
}
