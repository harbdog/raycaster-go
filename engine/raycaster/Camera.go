package raycaster

import (
	"image"
)

const (
	//--move speed--//
	moveSpeed = 0.06

	//--rotate speed--//
	rotSpeed = 0.03
)

// Camera Class that represents a camera in terms of raycasting.
// Contains methods to move the camera, and handles projection to,
// set the rectangle slice position and height,
type Camera struct {
	//--camera position, init to start position--//
	pos *Vector2

	//--current facing direction, init to values coresponding to FOV--//
	dir *Vector2

	//--the 2d raycaster version of camera plane, adjust y component to change FOV (ratio between this and dir x resizes FOV)--//
	plane *Vector2

	//--viewport width and height--//
	w int
	h int

	//--world map--//
	mapObj   *Map
	worldMap [][]int
	upMap    [][]int
	midMap   [][]int

	//--texture width--//
	texWidth int

	//--slices--//
	s []*image.Rectangle

	//--cam x pre calc--//
	camX []float64

	//--structs that contain rects and tints for each level or "floor" renderered--//
	lvls []*Level
}

// Vector2 converted struct from C#
type Vector2 struct {
	X float64
	Y float64
}

// NewCamera initalizes a Camera object
func NewCamera(width int, height int, texWid int, slices []*image.Rectangle, levels []*Level) *Camera {
	c := &Camera{}

	//private static Vector2 pos = new Vector2(22.5f, 11.5f);
	c.pos = &Vector2{X: 22.5, Y: 11.5}
	//private static Vector2 dir = new Vector2(-1.0f, 0.0f);
	c.dir = &Vector2{X: -1.0, Y: 0.0}
	//private static Vector2 plane = new Vector2(0.0f, 0.66f);
	c.plane = &Vector2{X: 0.0, Y: 0.66}

	c.w = width
	c.h = height
	c.texWidth = texWid
	c.s = slices
	c.lvls = levels

	//--init cam pre calc array--//
	c.camX = make([]float64, c.w)
	c.preCalcCamX()

	// map = new Map();
	// worldMap = map.getGrid();
	// upMap = map.getGridUp();
	// midMap = map.getGridMid();

	// raycast();//do an initial raycast

	return c
}

// precalculates camera x coordinate
func (c *Camera) preCalcCamX() {
	for x := 0; x < c.w; x++ {
		c.camX[x] = float64(2)*float64(x)/float64(c.w) - float64(1)
	}
}
