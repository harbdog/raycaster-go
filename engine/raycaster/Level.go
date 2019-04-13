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

// HorLevel is for handling horizontal slices (e.g. floors, ceiling)
type HorLevel struct {
	// HorBuffer has row or 'y' as first index, col or 'x' as second index
	HorBuffer [][]*HorPixel
}

// HorPixel is for representing individual horizontal buffer pixel requests
type HorPixel struct {
	// TexNum --the texture index to use as source
	TexNum int

	// TexX, TexY --texture source point
	TexX, TexY int

	// St --current slice tint (for lighting/shading)--//
	St *color.RGBA
}
