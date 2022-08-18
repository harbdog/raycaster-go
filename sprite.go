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

	// VerticalAnchor returns the vertical anchor position (only used when scaling image)
	VerticalAnchor() SpriteAnchor

	// Texture needs to return the current image to render
	Texture() *ebiten.Image

	// TextureRect needs to return the rectangle of the texture coordinates to draw
	TextureRect() image.Rectangle

	// SetScreenRect accepts the raycasted rectangle of the screen coordinates to be rendered (nil if not on screen)
	SetScreenRect(rect *image.Rectangle)
}

type SpriteAnchor int

const (
	AnchorBottom SpriteAnchor = iota
	AnchorCenter
	AnchorTop
)

func getAnchorVerticalOffset(anchor SpriteAnchor, spriteScale float64, cameraHeight int) float64 {
	if spriteScale == 1.0 {
		return 0
	}

	switch anchor {
	case AnchorBottom:
		halfHeight := float64(cameraHeight) / 2
		return halfHeight - (spriteScale * halfHeight)
	case AnchorCenter:
		return 0
	case AnchorTop:
		halfHeight := float64(cameraHeight) / 2
		return -(spriteScale * halfHeight)
	}

	return 0
}
