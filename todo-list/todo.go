package main

import (
	"fmt"
	"os"
)

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

var todos []Todo
var nextID int = 1

func main() {
	showMenu()

}

func showMenu() {
	fmt.Println("\n===== TODO APP =====")
	fmt.Println("1. Add Todo")
	fmt.Println("2. List Todos")
	fmt.Println("3. Complete Todo")
	fmt.Println("4. Delete Todo")
	fmt.Println("5. Exit")
	var choice int
	fmt.Print("Choose: ")
	fmt.Scanln(&choice)
	switch choice {
	case 1:
		addTodo("")
	case 2:
		listTodos()
		showMenu()
	case 3:
		completeTodo()
		showMenu()
	case 4:
		deleteTodo()
		showMenu()
	case 5:
		fmt.Println("Exiting...")
		os.Exit(0)
	default:
		fmt.Println("Invalid choice. Please try again.")
		showMenu()
	}
}

func addTodo(title string) {
	fmt.Print("Todo title: ")
	fmt.Scanln(&title)
	todo := Todo{
		ID:        nextID,
		Title:     title,
		Completed: false,
	}
	todos = append(todos, todo)
	nextID++
	fmt.Println("Todo added successfully!")
	showMenu()
}

func listTodos() {
	if len(todos) == 0 {
		fmt.Println("No todos found.")
	} else {
		fmt.Println("Todos:")
		for _, todo := range todos {
			status := "❌"
			if todo.Completed {
				status = "✅"
			}
			fmt.Printf("%d. [%s] %s\n", todo.ID, status, todo.Title)
		}
	}
}

func completeTodo() {
	id := readID("Enter the ID of the todo to complete: ")

	for i, todo := range todos {
		if todo.ID == id {
			todos[i].Completed = true
			fmt.Println("Todo marked as completed!")
			showMenu()
			return
		}
	}
	fmt.Println("Todo not found!")
	showMenu()
}

func deleteTodo() {
	id := readID("Enter the ID of the todo to delete: ")

	for i, todo := range todos {
		if todo.ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			fmt.Println("Todo deleted successfully!")
			showMenu()
			return
		}
	}
	fmt.Println("Todo not found!")
	showMenu()
}

func readID(message string) int {
	var id int
	fmt.Print(message)
	fmt.Scanln(&id)
	return id
}
