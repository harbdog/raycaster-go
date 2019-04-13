package main

import (
	"fmt"

	"github.com/harbdog/raycaster-go/engine"
)

func main() {
	fmt.Printf("Hello there\n")

	// run the game
	g := engine.NewGame()
	g.Run()
}
