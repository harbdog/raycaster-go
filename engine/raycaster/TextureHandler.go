package raycaster

import (
	"image"
)

type TextureHandler struct {
	slices []*image.Rectangle
}

func NewTextureHandler(texWidth int) *TextureHandler {
	t := &TextureHandler{}

	//--init array--//
	t.slices = make([]*image.Rectangle, texWidth)

	//--for clarity in slice loop--//
	texHeight := texWidth

	//--loop through creating a "slice" for each texture x--//
	for x := 0; x < texWidth; x++ {
		//tex width and height are always equal so safe to use tex width instead of height here
		thisRect := image.Rect(x, 0, x+1, texHeight)
		t.slices[x] = &thisRect
	}

	return t
}

func (t *TextureHandler) getSlices() []*image.Rectangle {
	return t.slices
}
