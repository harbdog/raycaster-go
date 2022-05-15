package model

import (
	"image/color"

	"raycaster-go/engine/geom"
)

type Player struct {
	*Entity
	Pitch          float64
	Moved          bool
	WeaponCooldown float64
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity: &Entity{
			Pos:      &geom.Vector2{X: x, Y: y},
			Angle:    angle,
			Velocity: 0,
			MapColor: color.RGBA{255, 0, 0, 255},
		},
		Pitch:          pitch,
		Moved:          false,
		WeaponCooldown: 0,
	}

	return p
}
