package model

import (
	"image/color"

	"github.com/harbdog/raycaster-go/geom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type Weapon struct {
	*Sprite
	firing             bool
	cooldown           int
	rateOfFire         float64
	projectileVelocity float64
	projectile         Projectile
}

func NewAnimatedWeapon(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int, projectile Projectile, projectileVelocity, rateOfFire float64,
) *Weapon {
	mapColor := color.RGBA{0, 0, 0, 0}
	w := &Weapon{
		Sprite: NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, 0, 0, 0),
	}
	w.projectile = projectile
	w.projectileVelocity = projectileVelocity
	w.rateOfFire = rateOfFire

	return w
}

func (w *Weapon) Fire() bool {
	if w.cooldown <= 0 {
		// TODO: handle rate of fire greater than 60 per second?
		w.cooldown = int(1 / w.rateOfFire * float64(ebiten.MaxTPS()))

		if !w.firing {
			w.firing = true
			w.Sprite.ResetAnimation()
		}

		return true
	}
	return false
}

func (w *Weapon) SpawnProjectile(x, y, z, angle, pitch float64, spawnedBy *Entity) *Projectile {
	p := &Projectile{}
	s := &Sprite{}
	copier.Copy(p, w.projectile)
	copier.Copy(s, w.projectile.Sprite)

	p.Sprite = s
	p.Pos = &geom.Vector2{X: x, Y: y}
	p.PosZ = z
	p.Angle = angle
	p.Pitch = pitch

	// convert velocity from distance/second to distance per tick
	p.Velocity = w.projectileVelocity / float64(ebiten.MaxTPS())

	// keep track of what spawned it
	p.Parent = spawnedBy

	return p
}

func (w *Weapon) OnCooldown() bool {
	return w.cooldown > 0
}

func (w *Weapon) ResetCooldown() {
	w.cooldown = 0
}

func (w *Weapon) Update() {
	if w.cooldown > 0 {
		w.cooldown -= 1
	}
	if w.firing && w.Sprite.GetLoopCounter() < 1 {
		w.Sprite.Update(nil)
	} else {
		w.firing = false
		w.Sprite.ResetAnimation()
	}
}
