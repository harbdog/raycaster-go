package raycaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextureHandler struct {
	Textures []*ebiten.Image

	// TODO: an interface would better suit texture handling
	FloorTex *image.RGBA
}

func NewTextureHandler(texSize int) *TextureHandler {
	t := &TextureHandler{}

	return t
}

func makeSlices(width, height, xOffset, yOffset int) []*image.Rectangle {
	newSlices := make([]*image.Rectangle, width)

	//--loop through creating a "slice" for each texture x--//
	for x := 0; x < width; x++ {
		// xOffset/yOffset represent sprite sheet source offsets for texture
		thisRect := image.Rect(xOffset+x, yOffset, xOffset+x+1, yOffset+height)
		newSlices[x] = &thisRect
	}

	return newSlices
}
