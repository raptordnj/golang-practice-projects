package main

import "fmt"

func main() {
	var currentYear int
	var yearOfBirth int

BEGIN:
	fmt.Print("Enter the current year: ")
	fmt.Scan(&currentYear)
	fmt.Print("Enter your year of birth:")
	fmt.Scan(&yearOfBirth)
	if yearOfBirth < 0 || yearOfBirth > currentYear {
		fmt.Println("Invalid year of birth. Please try again.")
		goto BEGIN
	}
	age := currentYear - yearOfBirth

	fmt.Printf("You are %d years old.\n", age)
	goto BEGIN
}
