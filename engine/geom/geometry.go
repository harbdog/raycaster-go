package geom

import "math"

const eps = 1e-14

func sq(x float64) float64 { return x * x }

func Degrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func Radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Vector2 converted struct from C#
type Vector2 struct {
	X, Y float64
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
	return sq(x2-x1) + sq(y2-y1)
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

// LineIntersection calculates the intersection of given two lines.
func LineIntersection(l1, l2 Line) (float64, float64, bool) {
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

type Circle struct {
	X, Y   float64
	Radius float64
}

// CircleIntersection gets the intersection points (if any) of a circle,
// and either an infinite line or a line segment.
func CircleIntersection(li Line, ci Circle, isSegment bool) []Vector2 {
	// https://rosettacode.org/wiki/Line_circle_intersection#Go
	var res []Vector2
	x0, y0 := ci.X, ci.Y
	x1, y1 := li.X1, li.Y1
	x2, y2 := li.X2, li.Y2
	A := y2 - y1
	B := x1 - x2
	C := x2*y1 - x1*y2
	a := sq(A) + sq(B)
	var b, c float64
	var bnz = true
	if math.Abs(B) >= eps { // if B isn't zero or close to it
		b = 2 * (A*C + A*B*y0 - sq(B)*x0)
		c = sq(C) + 2*B*C*y0 - sq(B)*(sq(ci.Radius)-sq(x0)-sq(y0))
	} else {
		b = 2 * (B*C + A*B*x0 - sq(A)*y0)
		c = sq(C) + 2*A*C*x0 - sq(A)*(sq(ci.Radius)-sq(x0)-sq(y0))
		bnz = false
	}
	d := sq(b) - 4*a*c // discriminant
	if d < 0 {
		// line & circle don't intersect
		return res
	}

	// checks whether a point is within a segment
	within := func(x, y float64) bool {
		d1 := math.Sqrt(sq(x2-x1) + sq(y2-y1)) // distance between end-points
		d2 := math.Sqrt(sq(x-x1) + sq(y-y1))   // distance from point to one end
		d3 := math.Sqrt(sq(x2-x) + sq(y2-y))   // distance from point to other end
		delta := d1 - d2 - d3
		return math.Abs(delta) < eps // true if delta is less than a small tolerance
	}

	var x, y float64
	fx := func() float64 { return -(A*x + C) / B }
	fy := func() float64 { return -(B*y + C) / A }
	rxy := func() {
		if !isSegment || within(x, y) {
			res = append(res, Vector2{X: x, Y: y})
		}
	}

	if d == 0 {
		// line is tangent to circle, so just one intersect at most
		if bnz {
			x = -b / (2 * a)
			y = fx()
			rxy()
		} else {
			y = -b / (2 * a)
			x = fy()
			rxy()
		}
	} else {
		// two intersects at most
		d = math.Sqrt(d)
		if bnz {
			x = (-b + d) / (2 * a)
			y = fx()
			rxy()
			x = (-b - d) / (2 * a)
			y = fx()
			rxy()
		} else {
			y = (-b + d) / (2 * a)
			x = fy()
			rxy()
			y = (-b - d) / (2 * a)
			x = fy()
			rxy()
		}
	}
	return res
}
