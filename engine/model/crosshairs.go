package model

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type Crosshairs struct {
	*Sprite
	Scale        float64
	HitIndicator *Crosshairs
}

func NewCrosshairs(
	x, y, scale float64, img *ebiten.Image, columns, rows, crosshairIndex, hitIndex int, uSize int,
) *Crosshairs {
	mapColor := color.RGBA{0, 0, 0, 0}
	c := &Crosshairs{
		Sprite:       NewSpriteFromSheet(x, y, 1.0, img, mapColor, columns, rows, crosshairIndex, uSize, 0),
		Scale:        scale,
		HitIndicator: &Crosshairs{},
	}

	hitIndicator := &Crosshairs{}
	copier.Copy(hitIndicator, c)
	c.HitIndicator = hitIndicator

	return c
}
