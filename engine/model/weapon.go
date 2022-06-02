package model

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Weapon struct {
	*Sprite
	firing bool
}

func NewAnimatedWeapon(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int,
) *Weapon {
	mapColor := color.RGBA{0, 0, 0, 0}
	w := &Weapon{
		Sprite: NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, 0, 0, 0),
	}

	return w
}

func (w *Weapon) Fire() {
	w.firing = true
	w.Sprite.ResetAnimation()
}

func (w *Weapon) Update() {
	if w.firing && w.Sprite.GetLoopCounter() < 1 {
		w.Sprite.Update(nil)
	} else {
		w.firing = false
		w.Sprite.ResetAnimation()
	}
}
