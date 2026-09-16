package main

import (
	"flag"
	"fmt"
	"os"

	"bagh-bakri/internal/game"
	"bagh-bakri/internal/ui"
)

func main() {
	cli := flag.Bool("cli", false, "run in CLI debug mode")
	flag.Parse()

	if *cli {
		runCLI()
		return
	}

	fmt.Println("Bagh-Bakri: starting GUI...")
	if err := runGUI(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runCLI() {
	state := game.NewGame()
	fmt.Println("Bagh-Bakri - CLI mode")
	fmt.Println("Commands: show, moves, move <from> <to>, place <pos>, undo, restart, status, quit")
	printState(state)

	for {
		fmt.Print("> ")
		var cmd string
		_, err := fmt.Scan(&cmd)
		if err != nil {
			break
		}

		switch cmd {
		case "show":
			printState(state)
		case "moves":
			moves := game.LegalMoves(state)
			for _, m := range moves {
				fmt.Println(m)
			}
		case "move":
			var from, to int
			fmt.Scan(&from, &to)
			moves := game.LegalMoves(state)
			found := false
			for _, m := range moves {
				if m.From == game.PositionID(from) && m.To == game.PositionID(to) {
					state, err = game.ApplyMove(state, m)
					if err != nil {
						fmt.Println("Error:", err)
					} else {
						found = true
					}
					break
				}
			}
			if !found {
				fmt.Println("Illegal move")
			}
			printState(state)
		case "place":
			var pos int
			fmt.Scan(&pos)
			moves := game.LegalMoves(state)
			found := false
			for _, m := range moves {
				if m.To == game.PositionID(pos) {
					state, err = game.ApplyMove(state, m)
					if err != nil {
						fmt.Println("Error:", err)
					} else {
						found = true
					}
					break
				}
			}
			if !found {
				fmt.Println("Illegal placement")
			}
			printState(state)
		case "undo":
			state = game.Undo(state)
			printState(state)
		case "restart":
			state = game.NewGame()
			printState(state)
		case "status":
			fmt.Println("Winner:", game.Winner(state))
			fmt.Println("Game over:", game.IsGameOver(state))
		case "quit":
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func printState(state game.GameState) {
	fmt.Println(state.Board)
	fmt.Printf("Turn: %d, Goats placed: %d, Captured: %d\n", state.Turn, state.GoatsPlaced, state.CapturedGoats)
}

func runGUI() error {
	app := ui.NewFyneApp()
	app.Run()
	return nil
}
