package raycaster

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// level --struct to represent rects and tints of vertical level slices --//
type level struct {
	// Sv --texture draw location
	Sv []*image.Rectangle

	// Cts --texture source location
	Cts []*image.Rectangle

	// St --current slice tint (for lighting/shading)--//
	St []*color.RGBA

	// CurrTex --the texture to use as source
	CurrTex []*ebiten.Image
}

// sliceView Creates rectangle slices for each x in width.
func sliceView(width, height int) []*image.Rectangle {
	arr := make([]*image.Rectangle, width)

	for x := 0; x < width; x++ {
		thisRect := image.Rect(x, 0, x+1, height)
		arr[x] = &thisRect
	}

	return arr
}

// horLevel is for handling horizontal renders that cannot use vertical slices (e.g. floor, ceiling)
type horLevel struct {
	// horBuffer is the image representing the pixels to render during the update
	horBuffer *image.RGBA
	// image is the ebitengine image object rendering the horBuffer during draw
	image *ebiten.Image
}

func (h *horLevel) initialize(width, height int) {
	h.horBuffer = image.NewRGBA(image.Rect(0, 0, width, height))
	if h.image == nil {
		h.image = ebiten.NewImage(width, height)
	}
}
