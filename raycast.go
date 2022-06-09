package main

import (
	"fmt"
	"runtime"

	"github.com/harbdog/raycaster-go/engine"
)

func main() {
	numCPU := runtime.NumCPU()
	// only way to see maxprocs is to set it and see the return value, then set it back
	maxProcs := runtime.GOMAXPROCS(numCPU)
	runtime.GOMAXPROCS(maxProcs)
	fmt.Printf("Hello there, you have %v cores and %v GOMAXPROCS\n", numCPU, maxProcs)

	// run the game
	g := engine.NewGame()
	g.Run()
}
