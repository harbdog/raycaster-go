package raycaster

type Map interface {
	// Level returns the 2-dimensional array of texture indices for each level
	Level(levelNum int) [][]int

	// NumLevels returns the number of vertical levels (minimum of 1)
	NumLevels() int
}
