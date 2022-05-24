package model

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Projectile struct {
	*Sprite
	Ricochets    int
	Lifespan     float64
	ImpactEffect Effect
}

func NewProjectile(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	uSize int, anchor SpriteAnchor, collisionRadius float64,
) *Projectile {
	p := &Projectile{
		Sprite:       NewSprite(x, y, scale, img, mapColor, uSize, anchor, collisionRadius),
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	return p
}

func NewAnimatedProjectile(
	x, y, scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA, columns, rows int,
	uSize int, anchor SpriteAnchor, collisionRadius float64,
) *Projectile {
	p := &Projectile{
		Sprite:       NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, uSize, anchor, collisionRadius),
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	return p
}
