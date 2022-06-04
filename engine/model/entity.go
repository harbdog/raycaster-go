package model

import (
	"image/color"

	"raycaster-go/engine/geom"
)

type Entity struct {
	Pos             *geom.Vector2
	PosZ            float64
	Angle           float64
	Pitch           float64
	Velocity        float64
	CollisionRadius float64
	MapColor        color.RGBA
	Parent          *Entity
}
