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

// HorLevel is for handling horizontal renders that cannot use vertical slices (e.g. floor, ceiling)
type HorLevel struct {
	// HorBuffer is the image representing the pixels to render during the update
	HorBuffer *image.RGBA

	// TexRGBA contains image.RGBA textures used as sources for the HorBuffer
	TexRGBA []*image.RGBA
}

func (h *HorLevel) Clear(width, height int) {
	h.HorBuffer = image.NewRGBA(image.Rect(0, 0, width, height))
}

func (h *HorLevel) Set(x, y int, c color.Color) {
	h.HorBuffer.Set(x, y, c)
}
