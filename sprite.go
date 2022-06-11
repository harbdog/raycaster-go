package raycaster

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
)

type Sprite interface {
	// Pos returns the X,Y map position
	Pos() *geom.Vector2

	// PosZ returns the Z (vertical) position
	PosZ() float64

	// Scale returns the scale factor (for no scaling, default to 1.0)
	Scale() float64

	// VerticalOffset returns the vertical offset needed (often needed when scaling, default to 0.0)
	VerticalOffset() float64

	// Texture needs to return the current image to render
	Texture() *ebiten.Image

	// TextureRect needs to return the rectangle of the texture coordinates to draw
	TextureRect() image.Rectangle
}
