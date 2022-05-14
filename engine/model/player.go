package model

import (
	"raycaster-go/engine/geom"
)

type Player struct {
	*Entity
	Pitch float64
	Moved bool
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity: &Entity{
			Pos:      &geom.Vector2{X: x, Y: y},
			Angle:    angle,
			Velocity: 0,
		},
		Pitch: pitch,
		Moved: false,
	}

	return p
}
