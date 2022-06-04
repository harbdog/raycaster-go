package model

import (
	"image/color"
	"math"
	"raycaster-go/engine/geom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type Projectile struct {
	*Sprite
	Ricochets    int
	Lifespan     float64
	ImpactEffect Effect
}

func NewAnimatedProjectile(
	x, y, scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA, columns, rows int,
	uSize int, vOffset, collisionRadius float64,
) *Projectile {
	p := &Projectile{
		Sprite:       NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, uSize, vOffset, collisionRadius),
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	return p
}

func (p *Projectile) SpawnEffect(x, y, z, angle, pitch float64) *Effect {
	e := &Effect{}
	s := &Sprite{}
	copier.Copy(e, p.ImpactEffect)
	copier.Copy(s, p.ImpactEffect.Sprite)

	e.Sprite = s
	e.Pos = &geom.Vector2{X: x, Y: y}
	e.PosZ = z
	e.Angle = angle
	e.Pitch = pitch

	// keep track of what spawned it
	e.Parent = p.Parent

	return e
}
