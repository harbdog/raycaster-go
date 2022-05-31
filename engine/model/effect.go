package model

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Effect struct {
	*Sprite
	LoopCount int
}

func NewAnimatedEffect(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int, uSize int, vOffset float64, loopCount int,
) *Effect {
	mapColor := color.RGBA{0, 0, 0, 0}
	e := &Effect{
		Sprite:    NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, uSize, vOffset, 0),
		LoopCount: loopCount,
	}

	return e
}
