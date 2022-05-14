package model

import "raycaster-go/engine/geom"

type Entity struct {
	Pos             *geom.Vector2
	Angle           float64
	Velocity        float64
	CollisionRadius float64
}
