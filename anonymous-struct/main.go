package main

import "fmt"

func main() {
	person := struct {
		Name string
		Age  int
	}{
		Name: "John Thompson",
		Age:  30,
	}

	fmt.Println("Name:", person.Name)
	fmt.Println("Age:", person.Age)
}
