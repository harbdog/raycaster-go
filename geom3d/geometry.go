package geom3d

import (
	"math"
)

func sq(x float64) float64 { return x * x }

//  Line implementation for 3-Dimensional Geometry applications
type Line3d struct {
	X1, Y1, Z1, X2, Y2, Z2 float64
}

// Heading gets the XY axis angle of the 3-dimensional line
func (l *Line3d) Heading() float64 {
	return math.Atan2(l.Y2-l.Y1, l.X2-l.X1)
}

// Pitch gets the Z axis angle of the 3-dimensional line
func (l *Line3d) Pitch() float64 {
	return math.Atan2(l.Z2-l.Z1, math.Sqrt(sq(l.X2-l.X1)+sq(l.Y2-l.Y1)))
}

// Distance gets the distance between the two endpoints of the 3-dimensional line
func (l *Line3d) Distance() float64 {
	return math.Sqrt(sq(l.X2-l.X1) + sq(l.Y2-l.Y1) + sq(l.Z2-l.Z1))
}

// Line3dFromAngle creates a 3-Dimensional line from a starting point at a given heading and pitch angle, and length
func Line3dFromAngle(x, y, z, heading, pitch, length float64) Line3d {
	return Line3d{
		X1: x,
		Y1: y,
		Z1: z,
		X2: x + (length * math.Cos(heading) * math.Cos(pitch)),
		Y2: y + (length * math.Sin(heading) * math.Cos(pitch)),
		Z2: z + (length * math.Sin(pitch)),
	}
}
