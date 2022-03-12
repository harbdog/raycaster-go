package raycaster

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// maximum number of concurrent tasks for large task sets (e.g. floor and sprite casting)
	maxConcurrent = 100

	// constant used for movement target framerate to prevent higher framerates from moving too fast
	movementTPS = 60.0

	// unit distance to keep camera away from wall to avoid clipping
	clipDistance = 0.1
)

type MouseMode int

const (
	MouseModeLook MouseMode = iota
	MouseModeMove
	MouseModeCursor
)

// Camera Class that represents a camera in terms of raycasting.
// Contains methods to move the camera, and handles projection to,
// set the rectangle slice position and height,
type Camera struct {
	//--camera position, init to start position--//
	pos *Vector2

	// vertical camera strafing up/down, for jumping/crouching
	posZ float64

	//--current facing direction, init to values coresponding to FOV--//
	dir *Vector2

	//--the 2d raycaster version of camera plane, adjust y component to change FOV (ratio between this and dir x resizes FOV)--//
	plane *Vector2

	//--viewport width and height--//
	w int
	h int

	// camera pitch
	pitch int

	// target framerate reference
	targetTPS int

	//--world map--//
	mapObj       *Map
	worldMap     [][]int
	upMap        [][]int
	midMap       [][]int
	collisionMap []Line

	//--texture width--//
	texWidth int

	//--slices--//
	s []*image.Rectangle

	//--structs that contain rects and tints for each level render--//
	lvls []*Level

	// zbuffer for sprite casting
	zBuffer []float64
	// sprites
	sprite []*Sprite
	//arrays used to sort the sprites
	spriteOrder    []int
	spriteDistance []float64

	spriteLvls []*Level
	tex        *TextureHandler

	horLvl *HorLevel

	// used for concurrency
	semaphore chan struct{}
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

func (v *Vector2) Copy() *Vector2 {
	return &Vector2{X: v.X, Y: v.Y}
}

// NewCamera initalizes a Camera object
func NewCamera(width int, height int, texWid int, mapObj *Map, slices []*image.Rectangle,
	levels []*Level, horizontalLevel *HorLevel, spriteLvls []*Level, tex *TextureHandler) *Camera {

	fmt.Printf("Initializing Camera\n")

	c := &Camera{}

	// set target FPS (TPS)
	// TODO: make target FPS customizable
	c.targetTPS = 60
	ebiten.SetMaxTPS(c.targetTPS)

	//--camera position, init to start position--//
	c.pos = &Vector2{X: 22.5, Y: 11.5}
	c.posZ = 0.0
	//--current facing direction, init to values coresponding to FOV--//
	c.dir = &Vector2{X: -1.0, Y: 0.0}
	//--the 2d raycaster version of camera plane, adjust y component to change FOV (ratio between this and dir x resizes FOV)--//
	c.plane = &Vector2{X: 0.0, Y: 0.66}

	c.w = width
	c.h = height
	c.pitch = 0
	c.texWidth = texWid
	c.s = slices
	c.lvls = levels

	c.horLvl = horizontalLevel
	c.spriteLvls = spriteLvls

	// set zbuffer based on screen width
	c.zBuffer = make([]float64, width)

	c.mapObj = mapObj
	c.worldMap = c.mapObj.getGrid()
	c.upMap = c.mapObj.getGridUp()
	c.midMap = c.mapObj.getGridMid()
	c.collisionMap = c.mapObj.getCollisionLines()

	c.sprite = c.mapObj.GetSprites()
	c.spriteOrder = make([]int, c.mapObj.numSprites)
	c.spriteDistance = make([]float64, c.mapObj.numSprites)

	c.tex = tex

	// initialize a pool of channels to limit concurrent floor and sprite casting
	// from https://pocketgophers.com/limit-concurrent-use/
	c.semaphore = make(chan struct{}, maxConcurrent)

	//do an initial raycast
	c.raycast()

	return c
}

// Update - updates the camera view
func (c *Camera) Update() {
	// clear horizontal buffer by making a new one
	c.horLvl.Clear(c.w, c.h)

	//--do raycast--//
	c.raycast()
}

func (c *Camera) raycast() {
	// cast level
	numLevels := cap(c.lvls)
	var wg sync.WaitGroup
	for i := 0; i < numLevels; i++ {
		wg.Add(1)
		go c.asyncCastLevel(i, &wg)
	}

	wg.Wait()

	//SPRITE CASTING
	//sort sprites from far to close
	numSprites := c.mapObj.numSprites
	for i := 0; i < numSprites; i++ {
		c.spriteOrder[i] = i
		c.spriteDistance[i] = (math.Pow(c.pos.X-c.sprite[i].X, 2) + math.Pow(c.pos.Y-c.sprite[i].Y, 2))
	}
	combSort(c.spriteOrder, c.spriteDistance, numSprites)

	//after sorting the sprites, do the projection and draw them
	for i := 0; i < numSprites; i++ {
		wg.Add(1)
		go c.asyncCastSprite(i, &wg)
	}

	wg.Wait()
}

func (c *Camera) asyncCastLevel(levelNum int, wg *sync.WaitGroup) {
	defer wg.Done()

	var rMap [][]int
	if levelNum == 0 {
		rMap = c.worldMap
	} else if levelNum == 1 {
		rMap = c.midMap
	} else {
		rMap = c.upMap //if above lvl2 just keep extending up
	}

	for x := 0; x < c.w; x++ {
		c.castLevel(x, rMap, c.lvls[levelNum], levelNum, wg)
	}
}

func (c *Camera) asyncCastSprite(spriteNum int, wg *sync.WaitGroup) {
	defer wg.Done()

	c.semaphore <- struct{}{} // Lock
	defer func() {
		<-c.semaphore // Unlock
	}()

	c.castSprite(spriteNum)
}

// credit : Raycast loop and setting up of vectors for matrix calculations
// courtesy - http://lodev.org/cgtutor/raycasting.html
func (c *Camera) castLevel(x int, grid [][]int, lvl *Level, levelNum int, wg *sync.WaitGroup) {
	var _cts, _sv []*image.Rectangle
	var _st []*color.RGBA

	_cts = lvl.Cts
	_sv = lvl.Sv
	_st = lvl.St

	//calculate ray position and direction
	cameraX := 2.0*float64(x)/float64(c.w) - 1.0 //x-coordinate in camera space
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
		mapWidth := len(c.worldMap)
		mapHeight := len(c.worldMap[0])

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
		if mapX < mapWidth && mapY < mapHeight && mapX > 0 && mapY > 0 {
			if grid[mapX][mapY] > 0 {
				hit = 1
			}
		} else {
			//hit grid boundary
			hit = 2

			//prevent out of range errors, needs to be improved
			if mapX < 0 {
				mapX = 0
			} else if mapX > mapWidth-1 {
				mapX = mapWidth - 1
			}

			if mapY < 0 {
				mapY = 0
			} else if mapY > mapHeight-1 {
				mapY = mapHeight - 1
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
	drawStart := (-lineHeight/2 + c.h/2) + c.pitch + int(c.posZ/perpWallDist) - lineHeight*levelNum
	drawEnd := drawStart + lineHeight

	//--due to modern way of drawing using quads this is removed to avoid glitches at the edges--//
	// if drawStart < 0 { drawStart = 0 }
	// if drawEnd >= c.h { drawEnd = c.h - 1 }

	//texturing calculations
	texNum := grid[mapX][mapY] - 1 //1 subtracted from it so that texture 0 can be used
	if texNum < 0 {
		texNum = 0 //why?
	}
	//--some supid hacks to make the houses render correctly--//
	// this corrects textures on two sides of house since the textures are not symmetrical
	if side == 0 {
		if texNum == 3 {
			texNum = 4
		} else if texNum == 4 {
			texNum = 3
		}

		if texNum == 1 {
			texNum = 4
		} else if texNum == 2 {
			texNum = 3
		}
	}

	c.lvls[levelNum].CurrTex[x] = c.tex.Textures[texNum]

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

	//// LIGHTING ////
	//--simulates torch light, as if player was carrying a radial light--//
	var lightFalloff float64 = -100 //decrease value to make torch dimmer

	//--sun brightness, illuminates whole level--//
	var sunLight float64 = 300 //global illumination

	//--distance based dimming of light--//
	var shadowDepth float64
	shadowDepth = math.Sqrt(perpWallDist) * lightFalloff
	_st[x].R = byte(Clamp(int(float64(_st[x].R)+shadowDepth+sunLight), 0, 255))
	_st[x].G = byte(Clamp(int(float64(_st[x].G)+shadowDepth+sunLight), 0, 255))
	_st[x].B = byte(Clamp(int(float64(_st[x].B)+shadowDepth+sunLight), 0, 255))

	//SET THE ZBUFFER FOR THE SPRITE CASTING
	if levelNum == 0 {
		// for now only rendering sprites on first level
		c.zBuffer[x] = perpWallDist //perpendicular distance is used
	}

	//// FLOOR CASTING ////
	if levelNum == 0 {
		// for now only rendering floor on first level
		if drawEnd < 0 {
			drawEnd = c.h //becomes < 0 when the integer overflows
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			c.semaphore <- struct{}{} // Lock
			defer func() {
				<-c.semaphore // Unlock
			}()

			var floorXWall, floorYWall float64

			//4 different wall directions possible
			if side == 0 && rayDirX > 0 {
				floorXWall = float64(mapX)
				floorYWall = float64(mapY) + wallX
			} else if side == 0 && rayDirX < 0 {
				floorXWall = float64(mapX) + 1.0
				floorYWall = float64(mapY) + wallX
			} else if side == 1 && rayDirY > 0 {
				floorXWall = float64(mapX) + wallX
				floorYWall = float64(mapY)
			} else {
				floorXWall = float64(mapX) + wallX
				floorYWall = float64(mapY) + 1.0
			}

			var distWall, distPlayer, currentDist float64

			distWall = perpWallDist
			distPlayer = 0.0

			//draw the floor from drawEnd to the bottom of the screen
			for y := drawEnd + 1; y < c.h; y++ {
				currentDist = (float64(c.h) + (2.0 * c.posZ)) / (2.0*float64(y-c.pitch) - float64(c.h))

				weight := (currentDist - distPlayer) / (distWall - distPlayer)

				currentFloorX := weight*floorXWall + (1.0-weight)*rayPosX
				currentFloorY := weight*floorYWall + (1.0-weight)*rayPosY

				var floorTexX, floorTexY int
				floorTexX = int(currentFloorX*float64(c.texWidth)) % c.texWidth
				floorTexY = int(currentFloorY*float64(c.texWidth)) % c.texWidth

				//floor
				// buffer[y][x] = (texture[3][texWidth * floorTexY + floorTexX] >> 1) & 8355711;
				// the same vertical slice method cannot be used for floor rendering
				floorTexNum := 0
				floorTex := c.horLvl.TexRGBA[floorTexNum]

				//pixel := floorTex.RGBAAt(floorTexX, floorTexY)
				pxOffset := floorTex.PixOffset(floorTexX, floorTexY)
				pixel := color.RGBA{floorTex.Pix[pxOffset],
					floorTex.Pix[pxOffset+1],
					floorTex.Pix[pxOffset+2],
					floorTex.Pix[pxOffset+3]}

				// lighting
				shadowDepth = math.Sqrt(currentDist) * lightFalloff
				pixelSt := &color.RGBA{255, 255, 255, 255}
				pixelSt.R = byte(Clamp(int(float64(pixelSt.R)+shadowDepth+sunLight), 0, 255))
				pixelSt.G = byte(Clamp(int(float64(pixelSt.G)+shadowDepth+sunLight), 0, 255))
				pixelSt.B = byte(Clamp(int(float64(pixelSt.B)+shadowDepth+sunLight), 0, 255))
				pixel.R = uint8(float64(pixel.R) * float64(pixelSt.R) / 256)
				pixel.G = uint8(float64(pixel.G) * float64(pixelSt.G) / 256)
				pixel.B = uint8(float64(pixel.B) * float64(pixelSt.B) / 256)

				//c.horLvl.HorBuffer.SetRGBA(x, y, pixel)
				pxOffset = c.horLvl.HorBuffer.PixOffset(x, y)
				c.horLvl.HorBuffer.Pix[pxOffset] = pixel.R
				c.horLvl.HorBuffer.Pix[pxOffset+1] = pixel.G
				c.horLvl.HorBuffer.Pix[pxOffset+2] = pixel.B
				c.horLvl.HorBuffer.Pix[pxOffset+3] = pixel.A
			}
		}()
	}
}

func (c *Camera) castSprite(spriteOrdIndex int) {
	// track whether the sprite actually needs to draw
	renderSprite := false

	rayPosX := c.pos.X
	rayPosY := c.pos.Y

	//translate sprite position to relative to camera
	spriteX := c.sprite[c.spriteOrder[spriteOrdIndex]].X - rayPosX
	spriteY := c.sprite[c.spriteOrder[spriteOrdIndex]].Y - rayPosY

	spriteTex := c.sprite[c.spriteOrder[spriteOrdIndex]].GetTexture()
	spriteW, spriteH := spriteTex.Size()

	//transform sprite with the inverse camera matrix
	// [ planeX   dirX ] -1                                       [ dirY      -dirX ]
	// [               ]       =  1/(planeX*dirY-dirX*planeY) *   [                 ]
	// [ planeY   dirY ]                                          [ -planeY  planeX ]

	invDet := 1.0 / (c.plane.X*c.dir.Y - c.dir.X*c.plane.Y) //required for correct matrix multiplication

	transformX := invDet * (c.dir.Y*spriteX - c.dir.X*spriteY)
	transformY := invDet * (-c.plane.Y*spriteX + c.plane.X*spriteY) //this is actually the depth inside the screen, that what Z is in 3D

	spriteScreenX := int(float64(c.w) / 2 * (1 + transformX/transformY))

	//parameters for scaling and moving the sprites
	var uDiv = 1
	var vDiv = 1
	var vMove = 0.0
	vMoveScreen := int(vMove/transformY) + c.pitch + int(c.posZ/transformY)

	//calculate height of the sprite on screen
	spriteHeight := int(math.Abs(float64(c.h)/transformY) / float64(vDiv)) //using "transformY" instead of the real distance prevents fisheye
	//calculate lowest and highest pixel to fill in current stripe
	drawStartY := -spriteHeight/2 + c.h/2 + vMoveScreen
	if drawStartY < 0 {
		drawStartY = 0
	}
	drawEndY := spriteHeight/2 + c.h/2 + vMoveScreen
	if drawEndY >= c.h {
		drawEndY = c.h - 1
	}

	//calculate width of the sprite
	spriteWidth := int(math.Abs(float64(c.h)/transformY) / float64(uDiv))
	drawStartX := -spriteWidth/2 + spriteScreenX
	drawEndX := spriteWidth/2 + spriteScreenX

	if drawStartX < 0 {
		drawStartX = 0
	}
	if drawEndX >= c.w {
		drawEndX = c.w - 1
	}

	var spriteSlices []*image.Rectangle

	//// LIGHTING ////
	//--simulates torch light, as if player was carrying a radial light--//
	var lightFalloff float64 = -100 //decrease value to make torch dimmer

	//--sun brightness, illuminates whole level--//
	var sunLight float64 = 300 //global illumination

	//loop through every vertical stripe of the sprite on screen
	for stripe := drawStartX; stripe < drawEndX; stripe++ {
		//the conditions in the if are:
		//1) it's in front of camera plane so you don't see things behind you
		//2) it's on the screen (left)
		//3) it's on the screen (right)
		//4) ZBuffer, with perpendicular distance
		if transformY > 0 && stripe > 0 && stripe < c.w && transformY < c.zBuffer[stripe] {
			var spriteLvl *Level
			if !renderSprite {
				renderSprite = true
				spriteLvl = c.makeSpriteLevel(spriteOrdIndex)
				spriteSlices = MakeSlices(spriteW, spriteH)
			} else {
				spriteLvl = c.spriteLvls[spriteOrdIndex]
			}

			texX := int(256*(stripe-(-spriteWidth/2+spriteScreenX))*c.texWidth/spriteWidth) / 256

			if texX < 0 || texX >= cap(spriteSlices) {
				continue
			}

			// modify tex startY and endY based on distance
			d := (drawStartY-vMoveScreen)*256 - c.h*128 + spriteHeight*128 //256 and 128 factors to avoid floats
			texStartY := ((d * c.texWidth) / spriteHeight) / 256

			d = (drawEndY-1-vMoveScreen)*256 - c.h*128 + spriteHeight*128
			texEndY := ((d * c.texWidth) / spriteHeight) / 256

			if texStartY < 0 || texStartY >= texEndY || texEndY >= c.texWidth {
				continue
			}

			//--set current texture slice--//
			spriteLvl.Cts[stripe] = spriteSlices[texX]
			spriteLvl.Cts[stripe].Min.Y = texStartY + 1
			spriteLvl.Cts[stripe].Max.Y = texEndY

			spriteLvl.CurrTex[stripe] = spriteTex

			//--set height of slice--//
			spriteLvl.Sv[stripe].Min.Y = drawStartY + 1

			//--set draw start of slice--//
			spriteLvl.Sv[stripe].Max.Y = drawEndY

			// distance based lighting/shading
			spriteLvl.St[stripe] = &color.RGBA{255, 255, 255, 255}

			//--distance based dimming of light--//
			shadowDepth := math.Sqrt(transformY) * lightFalloff
			spriteLvl.St[stripe].R = byte(Clamp(int(float64(spriteLvl.St[stripe].R)+shadowDepth+sunLight), 0, 255))
			spriteLvl.St[stripe].G = byte(Clamp(int(float64(spriteLvl.St[stripe].G)+shadowDepth+sunLight), 0, 255))
			spriteLvl.St[stripe].B = byte(Clamp(int(float64(spriteLvl.St[stripe].B)+shadowDepth+sunLight), 0, 255))
		}
	}

	if !renderSprite {
		c.clearSpriteLevel(spriteOrdIndex)
	}
}

func (c *Camera) makeSpriteLevel(spriteOrdIndex int) *Level {
	spriteLvl := new(Level)
	spriteLvl.Sv = SliceView(c.w, c.h)
	spriteLvl.Cts = make([]*image.Rectangle, c.w)
	spriteLvl.St = make([]*color.RGBA, c.w)
	spriteLvl.CurrTex = make([]*ebiten.Image, c.w)

	c.spriteLvls[spriteOrdIndex] = spriteLvl

	return spriteLvl
}

func (c *Camera) clearSpriteLevel(spriteOrdIndex int) {
	c.spriteLvls[spriteOrdIndex] = nil
}

//sort algorithm
func combSort(order []int, dist []float64, amount int) {
	gap := amount
	swapped := false
	for gap > 1 || swapped {
		//shrink factor 1.3
		gap = (gap * 10) / 13
		if gap == 9 || gap == 10 {
			gap = 11
		}
		if gap < 1 {
			gap = 1
		}

		swapped = false
		for i := 0; i < amount-gap; i++ {
			j := i + gap
			if dist[i] < dist[j] {
				// std::swap implementation for go:
				dist[i], dist[j] = dist[j], dist[i]
				order[i], order[j] = order[j], order[i]
				swapped = true
			}
		}
	}
}

// normalize speed based on a constant input rate
func (c *Camera) getNormalSpeed(speed float64) float64 {
	return speed * movementTPS / float64(c.targetTPS)
}

// checks for valid move from current position, returns valid (x, y) position
func (c *Camera) getValidMove(moveX, moveY float64, checkAlternate bool) (float64, float64) {
	posX := c.pos.X
	posY := c.pos.Y
	newX := moveX
	newY := moveY

	if posX == newX && posY == moveY {
		return posX, posY
	}

	ix := int(newX)
	iy := int(newY)

	// prevent index out of bounds errors
	switch {
	case ix < 0 || newX < 0:
		newX = clipDistance
		ix = 0
	case ix >= len(c.worldMap):
		newX = float64(len(c.worldMap)) - clipDistance
		ix = int(newX)
	}

	switch {
	case iy < 0 || newY < 0:
		newY = clipDistance
		iy = 0
	case iy >= len(c.worldMap[0]):
		newY = float64(len(c.worldMap[0])) - clipDistance
		iy = int(newY)
	}

	moveLine := Line{posX, posY, newX, newY}

	intersectPoints := [][2]float64{}
	for _, borderLine := range c.collisionMap {
		// TODO: only check intersection of nearby wall cells instead of all of them
		if px, py, ok := Intersection(moveLine, borderLine); ok {
			intersectPoints = append(intersectPoints, [2]float64{px, py})
		}
	}

	if len(intersectPoints) > 0 {
		// find the point closest to the start position
		min := math.Inf(1)
		minI := -1
		for i, p := range intersectPoints {
			d2 := Distance2(posX, posY, p[0], p[1])
			if d2 < min {
				min = d2
				minI = i
			}
		}

		// use the closest intersecting point to determine a safe distance to make the move
		moveLine = Line{posX, posY, intersectPoints[minI][0], intersectPoints[minI][1]}
		dist := math.Sqrt(min)
		angle := moveLine.Angle()

		// generate new move line using calculated angle and safe distance from intersecting point
		moveLine = LineFromAngle(posX, posY, angle, dist-0.01)

		newX, newY = moveLine.X2, moveLine.Y2
		ix, iy = int(newX), int(newY)

		// if either X or Y direction was already intersecting, attempt move only in the adjacent direction
		if checkAlternate {
			xDiff := math.Abs(newX - posX)
			yDiff := math.Abs(newY - posY)
			switch {
			case xDiff <= 0.01:
				// no more room to move in X, try to move only Y
				// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
				// 	c.pos.X, c.pos.Y, moveX, moveY, posX, moveY)
				return c.getValidMove(posX, moveY, false)
			case yDiff <= 0.01:
				// no more room to move in Y, try to move only X
				// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
				// 	c.pos.X, c.pos.Y, moveX, moveY, moveX, posY)
				return c.getValidMove(moveX, posY, false)
			}
		}

		// fmt.Printf("[@%v,%v] move to (%v,%v) intersects at {%v,%v}\n",
		// 	c.pos.X, c.pos.Y, moveX, moveY, newX, newY)
	}

	if c.worldMap[ix][iy] <= 0 {
		posX = newX
		posY = newY
	}

	return posX, posY
}

// Move camera by move speed
func (c *Camera) Move(mSpeed float64) {
	mSpeed = c.getNormalSpeed(mSpeed)
	mx := c.pos.X + (c.dir.X * mSpeed)
	my := c.pos.Y + (c.dir.Y * mSpeed)
	c.pos.X, c.pos.Y = c.getValidMove(mx, my, true)
}

// Strafe camera by strafe speed
func (c *Camera) Strafe(sSpeed float64) {
	sSpeed = c.getNormalSpeed(sSpeed)
	sx := c.pos.X + (c.plane.X * sSpeed)
	sy := c.pos.Y + (c.plane.Y * sSpeed)
	c.pos.X, c.pos.Y = c.getValidMove(sx, sy, true)
}

// Rotate camera by rotate speed
func (c *Camera) Rotate(rSpeed float64) {
	rSpeed = c.getNormalSpeed(rSpeed)

	//both camera direction and camera plane must be rotated
	oldDirX := c.dir.X
	c.dir.X = (c.dir.X*math.Cos(rSpeed) - c.dir.Y*math.Sin(rSpeed))
	c.dir.Y = (oldDirX*math.Sin(rSpeed) + c.dir.Y*math.Cos(rSpeed))
	oldPlaneX := c.plane.X
	c.plane.X = (c.plane.X*math.Cos(rSpeed) - c.plane.Y*math.Sin(rSpeed))
	c.plane.Y = (oldPlaneX*math.Sin(rSpeed) + c.plane.Y*math.Cos(rSpeed))
}

// Get current pitch value
func (c *Camera) GetPitch() int {
	return c.pitch
}

// Pitch camera by pitch delta
func (c *Camera) Pitch(pDelta int) {
	newPitch := Clamp(c.pitch+pDelta, -c.h/2, c.h/2)
	c.pitch = newPitch
}

// Stand camera position
func (c *Camera) Stand() {
	c.posZ = 0.0
}

// Crouch camera position
func (c *Camera) Crouch() {
	c.posZ = -200.0
}

// Jump camera position
func (c *Camera) Jump() {
	c.posZ = 200.0
}

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
