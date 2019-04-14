package raycaster

import (
	"image"
	"image/color"
)

// Level --struct to represent rects and tints of vertical level slices --//
type Level struct {
	// Sv --texture draw location
	Sv []*image.Rectangle

	// Cts --texture source location
	Cts []*image.Rectangle

	// St --current slice tint (for lighting/shading)--//
	St []*color.RGBA

	// CurrTexNum --the texture index to use as source
	CurrTexNum []int
}
