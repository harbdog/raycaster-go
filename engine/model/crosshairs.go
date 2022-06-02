package model

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type Crosshairs struct {
	*Sprite
	hitTimer     int
	HitIndicator *Crosshairs
}

func NewCrosshairs(
	x, y, scale float64, img *ebiten.Image, columns, rows, crosshairIndex, hitIndex int,
) *Crosshairs {
	mapColor := color.RGBA{0, 0, 0, 0}
	c := &Crosshairs{
		Sprite: NewSpriteFromSheet(x, y, scale, img, mapColor, columns, rows, crosshairIndex, 0, 0, 0),
	}

	hitIndicator := &Crosshairs{}
	copier.Copy(hitIndicator, c)
	hitIndicator.Sprite.SetAnimationFrame(hitIndex)
	c.HitIndicator = hitIndicator

	return c
}

func (c *Crosshairs) ActivateHitIndicator(hitTime int) {
	if c.HitIndicator != nil {
		c.hitTimer = hitTime
	}
}

func (c *Crosshairs) IsHitIndicatorActive() bool {
	return c.HitIndicator != nil && c.hitTimer > 0
}

func (c *Crosshairs) Update() {
	if c.HitIndicator != nil && c.hitTimer > 0 {
		// TODO: prefer to use timer rather than frame update counter?
		c.hitTimer -= 1
	}
}
