package geom

import "math"

func Degrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func Radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Vector2 converted struct from C#
type Vector2 struct {
	X float64
	Y float64
}

func (v *Vector2) Add(v2 *Vector2) *Vector2 {
	v.X += v2.X
	v.Y += v2.Y
	return v
}

func (v *Vector2) Sub(v2 *Vector2) *Vector2 {
	v.X -= v2.X
	v.Y -= v2.Y
	return v
}

func (v *Vector2) Copy() *Vector2 {
	return &Vector2{X: v.X, Y: v.Y}
}

func (v *Vector2) Equals(v2 *Vector2) bool {
	return v.X == v2.X && v.Y == v2.Y
}

// Line implementation for Geometry applications
type Line struct {
	X1, Y1, X2, Y2 float64
}

// Angle gets the angle of the line
func (l *Line) Angle() float64 {
	return math.Atan2(l.Y2-l.Y1, l.X2-l.X1)
}

// Distance2 returns the d^2 of the distance between two points
func Distance2(x1, y1, x2, y2 float64) float64 {
	return math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2)
}

// LineFromAngle creates a line from a starting point at a given angle and length
func LineFromAngle(x, y, angle, length float64) Line {
	return Line{
		X1: x,
		Y1: y,
		X2: x + (length * math.Cos(angle)),
		Y2: y + (length * math.Sin(angle)),
	}
}

// Rect implementation for Geometry applications
func Rect(x, y, w, h float64) []Line {
	return []Line{
		{x, y, x, y + h},
		{x, y + h, x + w, y + h},
		{x + w, y + h, x + w, y},
		{x + w, y, x, y},
	}
}

// Intersection calculates the intersection of given two lines.
func Intersection(l1, l2 Line) (float64, float64, bool) {
	// https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_points_on_each_line
	denom := (l1.X1-l1.X2)*(l2.Y1-l2.Y2) - (l1.Y1-l1.Y2)*(l2.X1-l2.X2)
	tNum := (l1.X1-l2.X1)*(l2.Y1-l2.Y2) - (l1.Y1-l2.Y1)*(l2.X1-l2.X2)
	uNum := -((l1.X1-l1.X2)*(l1.Y1-l2.Y1) - (l1.Y1-l1.Y2)*(l1.X1-l2.X1))

	if denom == 0 {
		return 0, 0, false
	}

	t := tNum / denom
	if t > 1 || t < 0 {
		return 0, 0, false
	}

	u := uNum / denom
	if u > 1 || u < 0 {
		return 0, 0, false
	}

	x := l1.X1 + t*(l1.X2-l1.X1)
	y := l1.Y1 + t*(l1.Y2-l1.Y1)
	return x, y, true
}
