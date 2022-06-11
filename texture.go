package raycaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextureHandler struct {
	slices   []*image.Rectangle
	Textures []*ebiten.Image
}

func NewTextureHandler(texWidth int) *TextureHandler {
	t := &TextureHandler{}

	//--for clarity in slice loop--//
	texHeight := texWidth

	//--init array--//
	t.slices = MakeSlices(texWidth, texHeight, 0, 0)

	return t
}

func MakeSlices(width, height, xOffset, yOffset int) []*image.Rectangle {
	newSlices := make([]*image.Rectangle, width)

	//--loop through creating a "slice" for each texture x--//
	for x := 0; x < width; x++ {
		// xOffset/yOffset represent sprite sheet source offsets for texture
		thisRect := image.Rect(xOffset+x, yOffset, xOffset+x+1, yOffset+height)
		newSlices[x] = &thisRect
	}

	return newSlices
}

func (t *TextureHandler) GetSlices() []*image.Rectangle {
	return t.slices
}
