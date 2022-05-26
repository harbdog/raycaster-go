package model

import (
	"image/color"

	"raycaster-go/engine/geom"
)

type Player struct {
	*Entity
	Moved          bool
	WeaponCooldown float64
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity: &Entity{
			Pos:      &geom.Vector2{X: x, Y: y},
			PosZ:     0.5,
			Angle:    angle,
			Pitch:    pitch,
			Velocity: 0,
			MapColor: color.RGBA{255, 0, 0, 255},
		},
		Moved:          false,
		WeaponCooldown: 0,
	}

	return p
}
