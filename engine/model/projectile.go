package model

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Projectile struct {
	*Sprite
	Ricochets int
	Lifespan  float64
}

func NewProjectile(x, y float64, img *ebiten.Image, uSize int, collisionRadius float64) *Projectile {
	p := &Projectile{
		Sprite:    NewSprite(x, y, img, uSize, collisionRadius),
		Ricochets: 0,
		Lifespan:  math.MaxFloat64,
	}

	return p
}

func NewAnimatedProjectile(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int,
	uSize int, collisionRadius float64,
) *Projectile {
	p := &Projectile{
		Sprite:    NewAnimatedSprite(x, y, scale, animationRate, img, columns, rows, uSize, collisionRadius),
		Ricochets: 0,
		Lifespan:  math.MaxFloat64,
	}

	return p
}
