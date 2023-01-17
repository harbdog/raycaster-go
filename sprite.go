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

	// Illumination needs to return sprite specific illumination offset (for normal illumination, default to 0)
	Illumination() float64

	// SetScreenRect accepts the raycasted rectangle of the screen coordinates to be rendered (nil if not on screen)
	SetScreenRect(rect *image.Rectangle)

	// IsFocusable should return true only if the convergence point can focus on the sprite
	IsFocusable() bool
}

type SpriteAnchor int

const (
	// AnchorBottom anchors the bottom of the sprite to its Z-position
	AnchorBottom SpriteAnchor = iota
	// AnchorCenter anchors the center of the sprite to its Z-position
	AnchorCenter
	// AnchorTop anchors the top of the sprite to its Z-position
	AnchorTop
)

func getAnchorVerticalOffset(anchor SpriteAnchor, spriteScale float64, cameraHeight int) float64 {
	halfHeight := float64(cameraHeight) / 2

	switch anchor {
	case AnchorBottom:
		return halfHeight - (spriteScale * halfHeight)
	case AnchorCenter:
		return halfHeight
	case AnchorTop:
		return halfHeight + (spriteScale * halfHeight)
	}

	return 0
}
