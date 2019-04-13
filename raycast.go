package main

import (
	"fmt"
	"runtime"

	"raycaster-go/engine"
)

func main() {
	fmt.Printf("Hello there, you have %v cores\n", runtime.NumCPU())

	// run the game
	g := engine.NewGame()
	g.Run()
}
