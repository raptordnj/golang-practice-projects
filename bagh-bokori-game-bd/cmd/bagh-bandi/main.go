// Command bagh-bandi plays বাঘবন্দী - the Bangladeshi Tiger-and-Goats board
// game - either as a Fyne desktop application or in the terminal.
//
// The two frontends share one engine. There is no second copy of the rules.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"bagh-bandi/internal/ai"
	"bagh-bandi/internal/cli"
	"bagh-bandi/internal/game"
	"bagh-bandi/internal/ui"
)

// version is stamped by the release build:
//
//	go build -ldflags "-X main.version=1.0.0" ./cmd/bagh-bandi
var version = "dev"

func main() {
	var (
		cliMode    = flag.Bool("cli", false, "play in the terminal instead of opening a window")
		tigerAI    = flag.String("tiger-ai", "", "computer plays the tiger: easy, medium or hard")
		goatAI     = flag.String("goat-ai", "", "computer plays the goats: easy, medium or hard")
		think      = flag.Duration("think", 3*time.Second, "AI thinking budget per move")
		goats      = flag.Int("goats", 0, "number of goats (default 16)")
		winCapture = flag.Int("win-captures", 0, "goats the tiger must eat to win (default 9)")
		forced     = flag.Bool("forced-capture", false, "make capture compulsory for the tiger")
		seed       = flag.Int64("seed", time.Now().UnixNano(), "random seed for the AI")
		showVer    = flag.Bool("version", false, "print the version and exit")
	)
	flag.Usage = usage
	flag.Parse()

	if *showVer {
		fmt.Printf("bagh-bandi %s\n", version)
		return
	}

	rules := game.DefaultRuleSet()
	if *goats > 0 {
		rules.TotalGoats = *goats
	}
	if *winCapture > 0 {
		rules.TigerWinCaptures = *winCapture
	}
	rules.ForcedCapture = *forced

	if rules.TigerWinCaptures > rules.TotalGoats {
		fmt.Fprintf(os.Stderr,
			"bagh-bandi: the tiger cannot need %d captures when there are only %d goats\n",
			rules.TigerWinCaptures, rules.TotalGoats)
		os.Exit(2)
	}

	if *cliMode || *tigerAI != "" || *goatAI != "" {
		os.Exit(runCLI(rules, *tigerAI, *goatAI, *think, *seed))
	}
	ui.NewApp().Run()
}

func runCLI(rules game.RuleSet, tigerAI, goatAI string, think time.Duration, seed int64) int {
	opts := cli.DefaultOptions()
	opts.Rules = rules
	opts.Think = think

	var err error
	if opts.TigerAI, err = buildAI(tigerAI, seed); err != nil {
		fmt.Fprintf(os.Stderr, "bagh-bandi: -tiger-ai: %v\n", err)
		return 2
	}
	if opts.GoatAI, err = buildAI(goatAI, seed+1); err != nil {
		fmt.Fprintf(os.Stderr, "bagh-bandi: -goat-ai: %v\n", err)
		return 2
	}

	if err := cli.NewSession(opts).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "bagh-bandi: %v\n", err)
		return 1
	}
	return 0
}

// buildAI turns a difficulty name into a player. An empty name means a human
// holds that side.
func buildAI(name string, seed int64) (ai.Player, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "":
		return nil, nil
	case "easy", "random":
		return ai.NewPlayer(ai.Easy, seed), nil
	case "medium":
		return ai.NewPlayer(ai.Medium, seed), nil
	case "hard":
		return ai.NewPlayer(ai.Hard, seed), nil
	default:
		return nil, fmt.Errorf("unknown difficulty %q (want easy, medium or hard)", name)
	}
}

func usage() {
	fmt.Fprint(flag.CommandLine.Output(), `বাঘবন্দী - Bagh-Bandi (Tiger & Goats)

Usage:
  bagh-bandi                     open the desktop game
  bagh-bandi --cli               play in the terminal
  bagh-bandi --cli --tiger-ai hard
                                 play the goats against a strong tiger
  bagh-bandi --tiger-ai medium --goat-ai easy
                                 watch two computers play

Options:
`)
	flag.PrintDefaults()
}
