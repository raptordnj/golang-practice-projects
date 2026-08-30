package main

import "fmt"

func main() {

	employees := map[string]map[string]string{
		"John": {
			"position":   "Software Engineer",
			"department": "Engineering",
		},
		"Jane": {
			"position":   "Product Manager",
			"department": "Product",
		},
	}

	for name, details := range employees {
		fmt.Printf("Name: %s, Position: %s, Department: %s\n", name, details["position"], details["department"])
	}

}
