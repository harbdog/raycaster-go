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
	t.slices = MakeSlices(texWidth, texHeight)

	return t
}

func MakeSlices(width, height int) []*image.Rectangle {
	newSlices := make([]*image.Rectangle, width)

	//--loop through creating a "slice" for each texture x--//
	for x := 0; x < width; x++ {
		//tex width and height are always equal so safe to use tex width instead of height here
		thisRect := image.Rect(x, 0, x+1, height)
		newSlices[x] = &thisRect
	}

	return newSlices
}

func (t *TextureHandler) GetSlices() []*image.Rectangle {
	return t.slices
}
