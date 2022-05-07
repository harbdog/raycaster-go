package engine

import (
	"raycaster-go/engine/geom"
	"raycaster-go/engine/model"
)

type Player struct {
	*model.Entity
	Pitch float64
	Moved bool
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity: &model.Entity{
			Pos:   &geom.Vector2{X: x, Y: y},
			Angle: angle,
		},
		Pitch: pitch,
		Moved: false,
	}

	return p
}
