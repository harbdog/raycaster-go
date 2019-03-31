package raycaster

import (
	"image"
	"image/color"
)

// Level --struct to represent rects and tints of a level--//
type Level struct {
	Sv  []*image.Rectangle
	Cts []*image.Rectangle

	//--current slice tint (for lighting)--//
	St         []*color.RGBA
	CurrTexNum []int
}
