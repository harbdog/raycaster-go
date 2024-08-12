package geom3d

import (
	"fmt"
	"math"
)

func sq(x float64) float64 { return x * x }

// 3-Dimensional point
type Vector3 struct {
	X, Y, Z float64
}

func (v *Vector3) String() string {
	return fmt.Sprintf("{%0.3f,%0.3f,%0.3f}", v.X, v.Y, v.Z)
}

func (v *Vector3) Add(v3 *Vector3) *Vector3 {
	v.X += v3.X
	v.Y += v3.Y
	v.Z += v3.Z
	return v
}

func (v *Vector3) Sub(v3 *Vector3) *Vector3 {
	v.X -= v3.X
	v.Y -= v3.Y
	v.Z -= v3.Z
	return v
}

func (v *Vector3) Copy() *Vector3 {
	return &Vector3{X: v.X, Y: v.Y, Z: v.Z}
}

func (v *Vector3) Equals(v3 *Vector3) bool {
	return v.X == v3.X && v.Y == v3.Y && v.Z == v3.Z
}

// Line implementation for 3-Dimensional Geometry applications
type Line3d struct {
	X1, Y1, Z1, X2, Y2, Z2 float64
}

func (l *Line3d) String() string {
	return fmt.Sprintf("{%0.3f,%0.3f,%0.3f->%0.3f,%0.3f,%0.3f}", l.X1, l.Y1, l.Z1, l.X2, l.Y2, l.Z2)
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

// Line3dFromAngle creates a 3-Dimensional line from starting point at a heading and pitch angle, and hypotenuse length
// based on answer from https://stackoverflow.com/questions/52781607/3d-point-from-two-angles-and-a-distance
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

// Line3dFromBaseAngle creates a 3-Dimensional line from starting point at a heading and pitch angle, and XY axis length
func Line3dFromBaseAngle(x, y, z, heading, pitch, xyLength float64) Line3d {
	return Line3d{
		X1: x,
		Y1: y,
		Z1: z,
		X2: x + (xyLength * math.Cos(heading)),
		Y2: y + (xyLength * math.Sin(heading)),
		Z2: z + (xyLength * math.Tan(pitch)),
	}
}
