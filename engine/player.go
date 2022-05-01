package engine

import "raycaster-go/engine/geom"

type Player struct {
	Pos          *geom.Vector2
	Angle, Pitch float64
	Moved        bool
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := new(Player)
	p.Pos = &geom.Vector2{X: x, Y: y}
	p.Angle = angle
	p.Pitch = pitch
	p.Moved = false

	return p
}
