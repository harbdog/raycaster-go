package raycaster

import (
	"image"
	"image/color"
	"math"
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

	//--camera position, init to start position--//
	c.pos = &Vector2{X: 22.5, Y: 11.5}
	//--current facing direction, init to values coresponding to FOV--//
	c.dir = &Vector2{X: -1.0, Y: 0.0}
	//--the 2d raycaster version of camera plane, adjust y component to change FOV (ratio between this and dir x resizes FOV)--//
	c.plane = &Vector2{X: 0.0, Y: 0.66}

	c.w = width
	c.h = height
	c.texWidth = texWid
	c.s = slices
	c.lvls = levels

	//--init cam pre calc array--//
	c.camX = make([]float64, c.w)
	c.preCalcCamX()

	c.mapObj = NewMap()
	c.worldMap = c.mapObj.getGrid()
	c.upMap = c.mapObj.getGridUp()
	c.midMap = c.mapObj.getGridMid()

	//do an initial raycast
	c.raycast()

	return c
}

// Update - updates the camera view
func (c *Camera) Update() {
	//--do raycast--//
	c.raycast()
}

// precalculates camera x coordinate
func (c *Camera) preCalcCamX() {
	for x := 0; x < c.w; x++ {
		c.camX[x] = 2.0*float64(x)/float64(c.w) - 1.0
	}
}

func (c *Camera) raycast() {
	for x := 0; x < c.w; x++ {
		for i := 0; i < cap(c.lvls); i++ {
			var rMap [][]int
			if i == 0 {
				rMap = c.worldMap
			} else if i == 1 {
				rMap = c.midMap
			} else {
				rMap = c.upMap //if above lvl2 just keep extending up
			}

			lvl := c.lvls[i]
			c.castLevel(x, rMap, lvl.Cts, lvl.Sv, lvl.St, i)
		}
	}
}

// credit : Raycast loop and setting up of vectors for matrix calculations
// courtesy - http://lodev.org/cgtutor/raycasting.html
func (c *Camera) castLevel(x int, grid [][]int, _cts []*image.Rectangle, _sv []*image.Rectangle, _st []*color.RGBA, levelNum int) {
	//calculate ray position and direction
	cameraX := c.camX[x] //x-coordinate in camera space
	rayDirX := c.dir.X + c.plane.X*cameraX
	rayDirY := c.dir.Y + c.plane.Y*cameraX

	//--rays start at camera position--//
	rayPosX := c.pos.X
	rayPosY := c.pos.Y

	//which box of the map we're in
	mapX := int(rayPosX)
	mapY := int(rayPosY)

	//length of ray from current position to next x or y-side
	var sideDistX float64
	var sideDistY float64

	//length of ray from one x or y-side to next x or y-side
	deltaDistX := math.Abs(1 / rayDirX)
	deltaDistY := math.Abs(1 / rayDirY)
	var perpWallDist float64

	//what direction to step in x or y-direction (either +1 or -1)
	var stepX int
	var stepY int

	hit := 0   //was there a wall hit?
	side := -1 //was a NS or a EW wall hit?

	//calculate step and initial sideDist
	if rayDirX < 0 {
		stepX = -1
		sideDistX = (rayPosX - float64(mapX)) * deltaDistX
	} else {
		stepX = 1
		sideDistX = (float64(mapX) + 1.0 - rayPosX) * deltaDistX
	}

	if rayDirY < 0 {
		stepY = -1
		sideDistY = (rayPosY - float64(mapY)) * deltaDistY
	} else {
		stepY = 1
		sideDistY = (float64(mapY) + 1.0 - rayPosY) * deltaDistY
	}

	//perform DDA
	for hit == 0 {
		//jump to next map square, OR in x-direction, OR in y-direction
		if sideDistX < sideDistY {
			sideDistX += deltaDistX
			mapX += stepX
			side = 0
		} else {
			sideDistY += deltaDistY
			mapY += stepY
			side = 1
		}

		//Check if ray has hit a wall
		if mapX < 24 && mapY < 24 && mapX > 0 && mapY > 0 {
			if grid[mapX][mapY] > 0 {
				hit = 1
			}
		} else {
			//hit grid boundary
			hit = 2

			//prevent out of range errors, needs to be improved
			if mapX < 0 {
				mapX = 0
			} else if mapX > 23 {
				mapX = 23
			}

			if mapY < 0 {
				mapY = 0
			} else if mapY > 23 {
				mapY = 23
			}
		}
	}

	//Calculate distance of perpendicular ray (oblique distance will give fisheye effect!)
	if side == 0 {
		perpWallDist = (float64(mapX) - rayPosX + (1.0-float64(stepX))/2.0) / rayDirX
	} else {
		perpWallDist = (float64(mapY) - rayPosY + (1.0-float64(stepY))/2.0) / rayDirY
	}

	//Calculate height of line to draw on screen
	lineHeight := int(float64(c.h) / perpWallDist)

	//calculate lowest and highest pixel to fill in current stripe
	drawStart := (-lineHeight/2 + c.h/2) - lineHeight*levelNum
	drawEnd := drawStart + lineHeight

	//--due to modern way of drawing using quads this is removed to avoid glitches at the edges--//
	// if drawStart < 0 { drawStart = 0 }
	// if drawEnd >= c.h { drawEnd = c.h - 1 }

	//texturing calculations
	texNum := grid[mapX][mapY] - 1 //1 subtracted from it so that texture 0 can be used
	if texNum < 0 {
		texNum = 0 //why?
	}
	c.lvls[levelNum].CurrTexNum[x] = texNum

	//calculate value of wallX
	var wallX float64 //where exactly the wall was hit
	if side == 0 {
		wallX = rayPosY + perpWallDist*rayDirY
	} else {
		wallX = rayPosX + perpWallDist*rayDirX
	}
	wallX -= math.Floor(wallX)

	//x coordinate on the texture
	texX := int(wallX * float64(c.texWidth))
	if side == 0 && rayDirX > 0 {
		texX = c.texWidth - texX - 1
	}

	if side == 1 && rayDirY < 0 {
		texX = c.texWidth - texX - 1
	}

	//--some supid hacks to make the houses render correctly--//
	// this corrects textures on two sides of house since the textures are not symmetrical
	if side == 0 {
		if texNum == 3 {
			c.lvls[levelNum].CurrTexNum[x] = 4
		} else if texNum == 4 {
			c.lvls[levelNum].CurrTexNum[x] = 3
		}

		if texNum == 1 {
			c.lvls[levelNum].CurrTexNum[x] = 4
		} else if texNum == 2 {
			c.lvls[levelNum].CurrTexNum[x] = 3
		}
	}

	//--set current texture slice to be slice x--//
	_cts[x] = c.s[texX]

	//--set height of slice--//
	_sv[x].Min.Y = drawStart

	//--set draw start of slice--//
	_sv[x].Max.Y = drawEnd

	//--add a bit of tint to differentiate between walls of a corner--//
	_st[x] = &color.RGBA{255, 255, 255, 255}
	if side == 1 {
		wallDiff := 12
		_st[x].R -= byte(wallDiff)
		_st[x].G -= byte(wallDiff)
		_st[x].B -= byte(wallDiff)
	}

	//--simulates torch light, as if player was carrying a radial light--//
	var lightFalloff float64 = -100 //decrease value to make torch dimmer

	//--sun brightness, illuminates whole level--//
	var sunLight float64 = 300 //global illuminaion

	//--distance based dimming of light--//
	var shadowDepth float64
	shadowDepth = math.Sqrt(perpWallDist) * lightFalloff
	_st[x].R = byte(Clamp(int(float64(_st[x].R)+shadowDepth+sunLight), 0, 255))
	_st[x].G = byte(Clamp(int(float64(_st[x].G)+shadowDepth+sunLight), 0, 255))
	_st[x].B = byte(Clamp(int(float64(_st[x].B)+shadowDepth+sunLight), 0, 255))
}

// Moves camera by move speed
func (c *Camera) Move(mSpeed float64) {
	if c.worldMap[int(c.pos.X+c.dir.X*mSpeed*12)][int(c.pos.Y)] <= 0 {
		c.pos.X += (c.dir.X * mSpeed)
	}
	if c.worldMap[int(c.pos.X)][int(c.pos.Y+c.dir.Y*mSpeed*12)] <= 0 {
		c.pos.Y += (c.dir.Y * mSpeed)
	}
}

// Rotates camera by rotate speed
func (c *Camera) Rotate(rSpeed float64) {
	//both camera direction and camera plane must be rotated
	oldDirX := c.dir.X
	c.dir.X = (c.dir.X*math.Cos(rSpeed) - c.dir.Y*math.Sin(rSpeed))
	c.dir.Y = (oldDirX*math.Sin(rSpeed) + c.dir.Y*math.Cos(rSpeed))
	oldPlaneX := c.plane.X
	c.plane.X = (c.plane.X*math.Cos(rSpeed) - c.plane.Y*math.Sin(rSpeed))
	c.plane.Y = (oldPlaneX*math.Sin(rSpeed) + c.plane.Y*math.Cos(rSpeed))
}

// public void rotate(double rSpeed)
// {
// 	//both camera direction and camera plane must be rotated
// 	double oldDirX = dir.X;
// 	dir.X = (float)(dir.X * Math.Cos(rSpeed) - dir.Y * Math.Sin(rSpeed));
// 	dir.Y = (float)(oldDirX * Math.Sin(rSpeed) + dir.Y * Math.Cos(rSpeed));
// 	double oldPlaneX = plane.X;
// 	plane.X = (float)(plane.X * Math.Cos(rSpeed) - plane.Y * Math.Sin(rSpeed));
// 	plane.Y = (float)(oldPlaneX * Math.Sin(rSpeed) + plane.Y * Math.Cos(rSpeed));
// }

// Clamp - converted C# method MathHelper.Clamp
// Restricts a value to be within a specified range.
func Clamp(value int, min int, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}

	return value
}
