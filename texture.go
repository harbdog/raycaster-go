package raycaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextureHandler interface {
	// TextureAt reutrns image used for rendered wall at the given x, y map coordinates and level number
	TextureAt(x, y, levelNum, side int) *ebiten.Image

	// FloorTextureAt returns image used for textured floor at the given x, y map coordinates
	FloorTextureAt(x, y int) *image.RGBA
}
