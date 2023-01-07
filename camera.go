package raycaster

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// maximum number of concurrent tasks for large task sets (e.g. floor and sprite casting)
	maxConcurrent = 100
)

// Camera Class that represents a camera in terms of raycasting.
// Contains methods to move the camera, and handles projection to,
// set the rectangle slice position and height,
type Camera struct {
	//--camera position, init to start position--//
	pos *geom.Vector2

	// vertical camera strafing up/down, for jumping/crouching
	camZ float64
	posZ float64

	//--current facing direction, init to values coresponding to FOV--//
	dir          *geom.Vector2
	headingAngle float64

	//--the 2d raycaster version of camera plane, adjust y component to change FOV (ratio between this and dir x resizes FOV)--//
	plane *geom.Vector2

	//--viewport width and height--//
	w int
	h int

	// camera pitch
	pitch      int
	pitchAngle float64

	// camera fov angle and depth
	fovAngle, fovDepth float64

	//--world map--//
	mapObj    Map
	mapWidth  int
	mapHeight int

	//--floor box, sky box textures--//
	floor *ebiten.Image
	sky   *ebiten.Image

	//--texture width--//
	texSize int

	//--structs that contain rects and tints for each level render--//
	levels   []*level
	floorLvl *horLevel
	slices   []*image.Rectangle

	// zbuffer for sprite casting
	zBuffer []float64
	// sprites
	sprites    []Sprite
	spriteLvls []*level
	//arrays used to sort the sprites
	spriteOrder    []int
	spriteDistance []float64

	tex TextureHandler

	//--simulates torch light, as if player was carrying a radial light--//
	lightFalloff float64

	//--global illumination for whole level (sun brightness)--//
	globalIllumination float64

	// controls the min/max color tinting of the textures when fully shadowed (min) or lighted (max)
	minLightRGB color.NRGBA
	maxLightRGB color.NRGBA

	// maximum distance to render raycasted objects
	renderDistance float64

	// point at which the center of the screen converges (for reticle use)
	convergenceDistance float64
	convergencePoint    *geom3d.Vector3

	// used for concurrency
	semaphore chan struct{}
}

// NewCamera initalizes a Camera object
func NewCamera(width int, height int, texSize int, mapObj Map, tex TextureHandler) *Camera {

	fmt.Printf("Initializing Camera\n")

	c := &Camera{}

	//--map setup
	c.mapObj = mapObj
	firstLevel := mapObj.Level(0)
	c.mapWidth = len(firstLevel)
	c.mapHeight = len(firstLevel[0])

	//--camera position, init to some start position--//
	c.pos = &geom.Vector2{X: 1.0, Y: 1.0}
	c.camZ = 0.0
	c.SetHeadingAngle(0)
	c.SetPitchAngle(0)

	fovDegrees := 70.0
	fovDepth := 1.0
	c.SetFovAngle(fovDegrees, fovDepth)

	// defaults for lighting and distant shadow
	c.SetRenderDistance(-1)
	c.SetLightFalloff(-100)
	c.SetGlobalIllumination(300)
	c.SetLightRGB(color.NRGBA{R: 0, G: 0, B: 0}, color.NRGBA{R: 255, G: 255, B: 255})

	c.texSize = texSize
	c.tex = tex
	c.SetViewSize(width, height)

	c.sprites = []Sprite{}
	c.updateSpriteLevels(16)

	// initialize a pool of channels to limit concurrent floor and sprite casting
	// from https://pocketgophers.com/limit-concurrent-use/
	c.semaphore = make(chan struct{}, maxConcurrent)

	c.convergenceDistance = -1
	c.convergencePoint = nil

	//do an initial raycast
	c.raycast()

	return c
}

// SetViewSize sets the camera resolution
func (c *Camera) SetViewSize(width, height int) {
	c.w = width
	c.h = height

	// creating level slices based on screen size
	c.levels = c.createLevels(c.mapObj.NumLevels())
	c.slices = makeSlices(c.texSize, c.texSize, 0, 0)
	c.floorLvl = c.createFloorLevel()

	// set zbuffer based on screen width
	c.zBuffer = make([]float64, width)
}

func (c *Camera) ViewSize() (int, int) {
	return c.w, c.h
}

// SetFovAngle sets the FOV angle (degrees) and depth
func (c *Camera) SetFovAngle(fovDegrees, fovDepth float64) {
	c.fovAngle = geom.Radians(fovDegrees)
	c.fovDepth = fovDepth

	var headingAngle float64 = 0
	if c.dir != nil {
		headingAngle = c.getAngleFromVec(c.dir)
	}
	c.dir = c.getVecForAngle(headingAngle)
	c.plane = c.getVecForFov(c.dir)
}

func (c *Camera) FovAngle() float64 {
	return geom.Degrees(c.fovAngle)
}

func (c *Camera) FovDepth() float64 {
	return c.fovDepth
}

// SetFloorTexture sets the static floorbox texture
func (c *Camera) SetFloorTexture(floor *ebiten.Image) {
	c.floor = floor
}

// SetSkyTexture sets the static skybox texture
func (c *Camera) SetSkyTexture(sky *ebiten.Image) {
	c.sky = sky
}

// SetRenderDistance sets maximum distance to render raycasted objects (-1 for practically inf)
func (c *Camera) SetRenderDistance(distance float64) {
	if distance < 0 {
		c.renderDistance = math.MaxFloat64
	} else {
		c.renderDistance = distance
	}
}

// SetLightFalloff sets value that simulates torch light, as if player was carrying a radial light.
// Lower values make torch dimmer.
func (c *Camera) SetLightFalloff(falloff float64) {
	c.lightFalloff = falloff
}

// SetGlobalIllumination sets illumination value for whole level (sun brightness)
func (c *Camera) SetGlobalIllumination(illumination float64) {
	c.globalIllumination = illumination
}

// SetLightRGB sets the min/max color tinting of the textures when fully shadowed (min) or lighted (max)
func (c *Camera) SetLightRGB(min, max color.NRGBA) {
	c.minLightRGB = min
	c.maxLightRGB = max
}

// Update - updates the camera view
func (c *Camera) Update(sprites []Sprite) {
	// clear horizontal buffer by making a new one
	c.floorLvl.initialize(c.w, c.h)

	// reset convergence point
	c.convergenceDistance = -1
	c.convergencePoint = nil

	if len(sprites) != len(c.sprites) {
		// sprite buffer may need to be increased in size
		c.updateSpriteLevels(len(sprites))
	} else {
		c.clearAllSpriteLevels()
	}

	//--do raycast--//
	c.sprites = sprites
	c.raycast()
}

func (c *Camera) raycast() {
	// cast level
	numLevels := c.mapObj.NumLevels()
	var wg sync.WaitGroup
	for i := 0; i < numLevels; i++ {
		wg.Add(1)
		go c.asyncCastLevel(i, &wg)
	}

	wg.Wait()

	//SPRITE CASTING
	numSprites := len(c.sprites)
	c.spriteOrder = make([]int, numSprites)
	c.spriteDistance = make([]float64, numSprites)
	//sort sprites from far to close
	for i := 0; i < numSprites; i++ {
		sprite := c.sprites[i]
		c.spriteOrder[i] = i
		c.spriteDistance[i] = math.Sqrt(math.Pow(c.pos.X-sprite.Pos().X, 2) + math.Pow(c.pos.Y-sprite.Pos().Y, 2))
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

	rMap := c.mapObj.Level(levelNum)

	for x := 0; x < c.w; x++ {
		c.castLevel(x, rMap, c.levels[levelNum], levelNum, wg)
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
func (c *Camera) castLevel(x int, grid [][]int, lvl *level, levelNum int, wg *sync.WaitGroup) {
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

		//Calculate distance of perpendicular ray (oblique distance will give fisheye effect!)
		if side == 0 {
			perpWallDist = sideDistX - deltaDistX
		} else {
			perpWallDist = sideDistY - deltaDistY
		}

		//Check if ray has hit a wall
		if mapX >= 0 && mapY >= 0 && mapX < c.mapWidth && mapY < c.mapHeight {
			if perpWallDist > c.renderDistance {
				// hit render distance bounds
				hit = 2
			} else if perpWallDist <= c.renderDistance && grid[mapX][mapY] > 0 {
				// only render walls within render distance
				hit = 1
			}
		} else {
			//hit grid boundary
			hit = 2
		}
	}

	//Calculate height of line to draw on screen
	lineHeight := int(float64(c.h) / perpWallDist)

	//calculate lowest and highest pixel to fill in current stripe
	drawStart := (-lineHeight/2 + c.h/2) + c.pitch + int(c.camZ/perpWallDist) - lineHeight*levelNum
	drawEnd := drawStart + lineHeight

	//--due to modern way of drawing using quads this is removed to avoid glitches at the edges--//
	// if drawStart < 0 { drawStart = 0 }
	// if drawEnd >= c.h { drawEnd = c.h - 1 }

	//calculate value of wallX
	var wallX float64 //where exactly the wall/boundary was hit
	if side == 0 {
		wallX = rayPosY + perpWallDist*rayDirY
	} else {
		wallX = rayPosX + perpWallDist*rayDirX
	}
	wallX -= math.Floor(wallX)

	//texturing calculations
	var texture *ebiten.Image
	if hit == 1 && mapX >= 0 && mapY >= 0 && mapX < c.mapWidth && mapY < c.mapHeight {
		texture = c.tex.TextureAt(mapX, mapY, levelNum, side)
	}

	c.levels[levelNum].CurrTex[x] = texture

	if texture != nil {
		//x coordinate on the texture
		texX := int(wallX * float64(c.texSize))
		if side == 0 && rayDirX > 0 {
			texX = c.texSize - texX - 1
		}

		if side == 1 && rayDirY < 0 {
			texX = c.texSize - texX - 1
		}

		//--set current texture slice to be slice x--//
		_cts[x] = c.slices[texX]

		//--set height of slice--//
		_sv[x].Min.Y = drawStart

		//--set draw start of slice--//
		_sv[x].Max.Y = drawEnd

		//// LIGHTING ////
		//--distance based dimming of light--//
		shadowDepth := math.Sqrt(perpWallDist) * c.lightFalloff
		_st[x] = &color.RGBA{255, 255, 255, 255}
		_st[x].R = byte(geom.ClampInt(int(float64(_st[x].R)+shadowDepth+c.globalIllumination), int(c.minLightRGB.R), int(c.maxLightRGB.R)))
		_st[x].G = byte(geom.ClampInt(int(float64(_st[x].G)+shadowDepth+c.globalIllumination), int(c.minLightRGB.G), int(c.maxLightRGB.G)))
		_st[x].B = byte(geom.ClampInt(int(float64(_st[x].B)+shadowDepth+c.globalIllumination), int(c.minLightRGB.B), int(c.maxLightRGB.B)))

		//--add a bit of tint to differentiate between walls of a corner--//
		if side == 0 {
			wallDiff := 12
			_st[x].R = byte(geom.ClampInt(int(_st[x].R)-wallDiff, 0, 255))
			_st[x].G = byte(geom.ClampInt(int(_st[x].G)-wallDiff, 0, 255))
			_st[x].B = byte(geom.ClampInt(int(_st[x].B)-wallDiff, 0, 255))
		}
	}

	// determine if is convergence point that hit a wall
	convergenceCol, convergenceRow := c.w/2-1, c.h/2-1
	if x == convergenceCol && drawStart <= convergenceRow && convergenceRow <= drawEnd {
		// use pitch angle and perpendicular distance (adjusted for fov zoom) to find Z point of convergence
		convergencePerpDist := perpWallDist * c.fovDepth
		convergenceLine3d := geom3d.Line3dFromBaseAngle(c.pos.X, c.pos.Y, c.posZ, c.headingAngle, c.pitchAngle, convergencePerpDist)
		convergenceDistance := convergenceLine3d.Distance()

		if c.convergenceDistance == -1 || convergenceDistance < c.convergenceDistance {
			c.convergenceDistance = convergenceDistance
			c.convergencePoint = &geom3d.Vector3{X: convergenceLine3d.X2, Y: convergenceLine3d.Y2, Z: convergenceLine3d.Z2}
		}
	}

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
			for y := drawEnd; y < c.h; y++ {
				currentDist = (float64(c.h) + (2.0 * c.camZ)) / (2.0*float64(y-c.pitch) - float64(c.h))
				if currentDist > c.renderDistance {
					continue
				}

				weight := (currentDist - distPlayer) / (distWall - distPlayer)

				currentFloorX := weight*floorXWall + (1.0-weight)*rayPosX
				currentFloorY := weight*floorYWall + (1.0-weight)*rayPosY

				// do not call FloorTextureAt interface if X/Y is outside of map bounds
				if currentFloorX < 0 || currentFloorY < 0 || int(currentFloorX) >= c.mapWidth || int(currentFloorY) >= c.mapHeight {
					continue
				}

				if x == convergenceCol && y == convergenceRow {
					// use pitch angle and perpendicular distance (adjusted for fov zoom) to find Z point of convergence
					convergencePerpDist := currentDist * c.fovDepth
					convergenceLine3d := geom3d.Line3dFromBaseAngle(c.pos.X, c.pos.Y, c.posZ, c.headingAngle, c.pitchAngle, convergencePerpDist)
					convergenceDistance := convergenceLine3d.Distance()

					if c.convergenceDistance == 0 || convergenceDistance < c.convergenceDistance {
						c.convergenceDistance = convergenceDistance
						c.convergencePoint = &geom3d.Vector3{X: convergenceLine3d.X2, Y: convergenceLine3d.Y2, Z: convergenceLine3d.Z2}
					}
				}

				//floor texture for map coordinate being rendered
				floorTex := c.tex.FloorTextureAt(int(currentFloorX), int(currentFloorY))
				if floorTex == nil {
					continue
				}

				floorTexX := int(currentFloorX*float64(c.texSize)) % c.texSize
				floorTexY := int(currentFloorY*float64(c.texSize)) % c.texSize

				// buffer[y][x] = (texture[3][texWidth * floorTexY + floorTexX] >> 1) & 8355711;
				// the same vertical slice method cannot be used for floor rendering
				// floorTexNum := 0
				// floorTex := c.floorLvl.texRGBA[floorTexNum]

				//pixel := floorTex.RGBAAt(floorTexX, floorTexY)
				pxOffset := floorTex.PixOffset(floorTexX, floorTexY)
				if pxOffset < 0 {
					continue
				}
				pixel := color.RGBA{floorTex.Pix[pxOffset],
					floorTex.Pix[pxOffset+1],
					floorTex.Pix[pxOffset+2],
					floorTex.Pix[pxOffset+3]}

				// lighting
				pixelSt := &color.RGBA{255, 255, 255, 255}
				shadowDepth := math.Sqrt(currentDist) * c.lightFalloff
				pixelSt.R = byte(geom.ClampInt(int(float64(pixelSt.R)+shadowDepth+c.globalIllumination), int(c.minLightRGB.R), int(c.maxLightRGB.R)))
				pixelSt.G = byte(geom.ClampInt(int(float64(pixelSt.G)+shadowDepth+c.globalIllumination), int(c.minLightRGB.G), int(c.maxLightRGB.G)))
				pixelSt.B = byte(geom.ClampInt(int(float64(pixelSt.B)+shadowDepth+c.globalIllumination), int(c.minLightRGB.B), int(c.maxLightRGB.B)))
				pixel.R = uint8(float64(pixel.R) * float64(pixelSt.R) / 256)
				pixel.G = uint8(float64(pixel.G) * float64(pixelSt.G) / 256)
				pixel.B = uint8(float64(pixel.B) * float64(pixelSt.B) / 256)

				//c.horLvl.HorBuffer.SetRGBA(x, y, pixel)
				pxOffset = c.floorLvl.horBuffer.PixOffset(x, y)
				c.floorLvl.horBuffer.Pix[pxOffset] = pixel.R
				c.floorLvl.horBuffer.Pix[pxOffset+1] = pixel.G
				c.floorLvl.horBuffer.Pix[pxOffset+2] = pixel.B
				c.floorLvl.horBuffer.Pix[pxOffset+3] = pixel.A
			}
		}()
	}
}

func (c *Camera) castSprite(spriteOrdIndex int) {
	// the sprite
	sprite := c.sprites[c.spriteOrder[spriteOrdIndex]]

	spriteDist := c.spriteDistance[spriteOrdIndex]
	if spriteDist > c.renderDistance {
		sprite.SetScreenRect(nil)
		return
	}

	// track whether the sprite actually needs to draw
	renderSprite := false

	//translate sprite position to relative to camera
	spriteX := sprite.Pos().X - c.pos.X
	spriteY := sprite.Pos().Y - c.pos.Y

	spriteTex := sprite.Texture()
	spriteTexRect := sprite.TextureRect()
	spriteTexWidth, spriteTexHeight := spriteTex.Size()

	//transform sprite with the inverse camera matrix
	// [ planeX   dirX ] -1                                       [ dirY      -dirX ]
	// [               ]       =  1/(planeX*dirY-dirX*planeY) *   [                 ]
	// [ planeY   dirY ]                                          [ -planeY  planeX ]

	invDet := 1.0 / (c.plane.X*c.dir.Y - c.dir.X*c.plane.Y) //required for correct matrix multiplication

	transformX := invDet * (c.dir.Y*spriteX - c.dir.X*spriteY)
	transformY := invDet * (-c.plane.Y*spriteX + c.plane.X*spriteY)

	spriteScreenX := int(float64(c.w) / 2 * (1 + transformX/transformY))

	//parameters for scaling and translating the sprites
	spriteScale := sprite.Scale()
	spriteAnchor := sprite.VerticalAnchor()

	var uDiv float64 = 1 / spriteScale
	var vDiv float64 = 1 / spriteScale
	var vOffset float64 = getAnchorVerticalOffset(spriteAnchor, spriteScale, c.h)

	var vMove float64 = -sprite.PosZ()*float64(c.h) + vOffset

	vMoveScreen := int(vMove/transformY) + c.pitch + int(c.camZ/transformY)

	//calculate height of the sprite on screen
	spriteHeight := int(math.Abs(float64(c.h)/transformY) / vDiv) //using "transformY" instead of the real distance prevents fisheye

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
	spriteWidth := int(math.Abs(float64(c.h)/transformY) / uDiv)

	drawStartX := -spriteWidth/2 + spriteScreenX
	drawEndX := spriteWidth/2 + spriteScreenX

	if drawStartX < 0 {
		drawStartX = 0
	}
	if drawEndX >= c.w {
		drawEndX = c.w - 1
	}

	if spriteWidth == 0 || spriteHeight == 0 {
		sprite.SetScreenRect(nil)
		return
	}

	// used to determine if is convergence point that hit a sprite
	canConverge := sprite.IsFocusable()
	convergenceCol, convergenceRow := c.w/2-1, c.h/2-1

	// modify tex startY and endY based on distance
	d := (drawStartY-vMoveScreen)*256 - c.h*128 + spriteHeight*128 //256 and 128 factors to avoid floats
	texStartY := ((d * spriteTexHeight) / spriteHeight) / 256

	d = (drawEndY-1-vMoveScreen)*256 - c.h*128 + spriteHeight*128
	texEndY := ((d * spriteTexHeight) / spriteHeight) / 256

	var spriteSlices []*image.Rectangle

	//loop through every vertical stripe of the sprite on screen
	for stripe := drawStartX; stripe < drawEndX; stripe++ {
		//the conditions in the if are:
		//1) it's in front of camera plane so you don't see things behind you
		//2) it's on the screen (left)
		//3) it's on the screen (right)
		//4) ZBuffer, with perpendicular distance
		if transformY > 0 && stripe > 0 && stripe < c.w && transformY < c.zBuffer[stripe] {
			var spriteLvl *level
			if !renderSprite {
				renderSprite = true
				spriteLvl = c.makeSpriteLevel(spriteOrdIndex)
				spriteSlices = makeSlices(spriteTexWidth, spriteTexHeight, spriteTexRect.Min.X, spriteTexRect.Min.Y)
			} else {
				spriteLvl = c.spriteLvls[spriteOrdIndex]
			}

			texX := int(256*(stripe-(-spriteWidth/2+spriteScreenX))*spriteTexWidth/spriteWidth) / 256
			if texX < 0 || texX >= cap(spriteSlices) {
				continue
			}

			if canConverge && stripe == convergenceCol && drawStartY <= convergenceRow && convergenceRow <= drawEndY {
				// use pitch angle and perpendicular distance (adjusted for fov zoom) to find Z point of convergence
				convergencePerpDist := spriteDist * c.fovDepth
				convergenceLine3d := geom3d.Line3dFromBaseAngle(c.pos.X, c.pos.Y, c.posZ, c.headingAngle, c.pitchAngle, convergencePerpDist)
				convergenceDistance := convergenceLine3d.Distance()

				if c.convergenceDistance == -1 || convergenceDistance < c.convergenceDistance {
					c.convergenceDistance = convergenceDistance
					c.convergencePoint = &geom3d.Vector3{X: convergenceLine3d.X2, Y: convergenceLine3d.Y2, Z: convergenceLine3d.Z2}
				}
			}

			//--set current texture slice--//
			spriteLvl.Cts[stripe] = spriteSlices[texX]
			spriteLvl.Cts[stripe].Min.Y = spriteTexRect.Min.Y + texStartY
			spriteLvl.Cts[stripe].Max.Y = spriteTexRect.Min.Y + texEndY

			spriteLvl.CurrTex[stripe] = spriteTex

			//--set draw start and height of slice--//
			spriteLvl.Sv[stripe].Min.Y = drawStartY
			spriteLvl.Sv[stripe].Max.Y = drawEndY

			//// LIGHTING ////
			// distance based lighting/shading
			shadowDepth := math.Sqrt(transformY) * c.lightFalloff
			spriteLvl.St[stripe] = &color.RGBA{255, 255, 255, 255}
			spriteLvl.St[stripe].R = byte(geom.ClampInt(int(float64(spriteLvl.St[stripe].R)+shadowDepth+c.globalIllumination), int(c.minLightRGB.R), int(c.maxLightRGB.R)))
			spriteLvl.St[stripe].G = byte(geom.ClampInt(int(float64(spriteLvl.St[stripe].G)+shadowDepth+c.globalIllumination), int(c.minLightRGB.G), int(c.maxLightRGB.G)))
			spriteLvl.St[stripe].B = byte(geom.ClampInt(int(float64(spriteLvl.St[stripe].B)+shadowDepth+c.globalIllumination), int(c.minLightRGB.B), int(c.maxLightRGB.B)))
		}
	}

	if renderSprite {
		// store raycasted sprite x/y view bounds so they can be retrieved by consumers
		spriteCastRect := image.Rect(drawStartX, drawStartY, drawEndX, drawEndY)
		sprite.SetScreenRect(&spriteCastRect)
	} else {
		c.clearSpriteLevel(spriteOrdIndex)
		sprite.SetScreenRect(nil)
	}
}

func makeSlices(width, height, xOffset, yOffset int) []*image.Rectangle {
	newSlices := make([]*image.Rectangle, width)

	//--loop through creating a "slice" for each texture x--//
	for x := 0; x < width; x++ {
		// xOffset/yOffset represent sprite sheet source offsets for texture
		thisRect := image.Rect(xOffset+x, yOffset, xOffset+x+1, yOffset+height)
		newSlices[x] = &thisRect
	}

	return newSlices
}

// creates level slices for raycasting each level
func (c *Camera) createLevels(numLevels int) []*level {
	levelArr := make([]*level, numLevels)

	for i := 0; i < numLevels; i++ {
		levelArr[i] = new(level)
		levelArr[i].Sv = sliceView(c.w, c.h)
		levelArr[i].Cts = make([]*image.Rectangle, c.w)
		levelArr[i].St = make([]*color.RGBA, c.w)
		levelArr[i].CurrTex = make([]*ebiten.Image, c.w)
	}

	return levelArr
}

// creates floor slices for raycasting floor
func (c *Camera) createFloorLevel() *horLevel {
	horizontalLevel := new(horLevel)
	horizontalLevel.initialize(c.w, c.h)
	return horizontalLevel
}

// updates sprite slice array as a level
func (c *Camera) updateSpriteLevels(spriteCapacity int) {
	if c.spriteLvls != nil {
		capacity := len(c.spriteLvls)
		if spriteCapacity <= capacity {
			// no need to grow, just need to clear it out
			c.clearAllSpriteLevels()
			return
		}

		for capacity <= spriteCapacity {
			capacity *= 2
		}

		spriteCapacity = capacity
	}
	c.spriteLvls = make([]*level, spriteCapacity)
}

func (c *Camera) makeSpriteLevel(spriteOrdIndex int) *level {
	spriteLvl := new(level)
	spriteLvl.Sv = sliceView(c.w, c.h)
	spriteLvl.Cts = make([]*image.Rectangle, c.w)
	spriteLvl.St = make([]*color.RGBA, c.w)
	spriteLvl.CurrTex = make([]*ebiten.Image, c.w)

	c.spriteLvls[spriteOrdIndex] = spriteLvl

	return spriteLvl
}

func (c *Camera) clearAllSpriteLevels() {
	for i := 0; i < len(c.spriteLvls); i++ {
		c.clearSpriteLevel(i)
	}
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

// Set camera position vector
func (c *Camera) SetPosition(pos *geom.Vector2) {
	c.pos = pos
}

// Get camera position vector
func (c *Camera) GetPosition() *geom.Vector2 {
	return c.pos
}

// Set camera Z-plane position
func (c *Camera) SetPositionZ(gridPosZ float64) {
	// convert grid position to camera position
	c.posZ = gridPosZ
	c.camZ = (gridPosZ - 0.5) * float64(c.h)
}

// Get camera Z-plane position
func (c *Camera) GetPositionZ() float64 {
	return c.posZ
}

// Set camera direction and plane vectors from given heading angle
func (c *Camera) SetHeadingAngle(headingAngle float64) {
	c.headingAngle = headingAngle
	cameraDir := c.getVecForAngle(headingAngle)
	c.dir = cameraDir
	c.plane = c.getVecForFov(cameraDir)
}

// Set camera pitch view from given pitch angle
func (c *Camera) SetPitchAngle(pitchAngle float64) {
	c.pitchAngle = pitchAngle
	cameraPitch := geom.GetOppositeTriangleLeg(pitchAngle, float64(c.h)*c.fovDepth)
	// clamping it since looking down or up too far causes floor texture glitches and wall warping
	c.pitch = geom.ClampInt(int(cameraPitch), -c.h/2, int(float64(c.h)*c.fovDepth))
}

// Get the angle from the dir vectors
func (c *Camera) getAngleFromVec(dir *geom.Vector2) float64 {
	return math.Atan2(dir.Y, dir.X)
}

// Get the dir vector from angle and fov length
func (c *Camera) getVecForAngleLength(angle, length float64) *geom.Vector2 {
	return &geom.Vector2{X: length * math.Cos(angle), Y: length * math.Sin(angle)}
}

func (c *Camera) getVecForAngle(angle float64) *geom.Vector2 {
	return &geom.Vector2{X: c.fovDepth * math.Cos(angle), Y: c.fovDepth * math.Sin(angle)}
}

// Get the plane vector from FOV based on dir vector
func (c *Camera) getVecForFov(dir *geom.Vector2) *geom.Vector2 {
	// get the hypotenuse of half the FOV triangle to calculate the plane vec points
	angle := c.getAngleFromVec(dir)
	length := math.Sqrt(math.Pow(dir.X, 2) + math.Pow(dir.Y, 2))
	hypotenuse := length / math.Cos(c.fovAngle/2)

	// subtract resulting vector from dir since plane vec is relative to it
	return dir.Copy().Sub(c.getVecForAngleLength(angle+c.fovAngle/2, hypotenuse))
}

// Get the distance to the point of convergence raycasted from the center of the camera view
func (c *Camera) GetConvergenceDistance() float64 {
	return c.convergenceDistance
}

// Get the 3-Dimensional point of convergence raycasted from the center of the camera view
func (c *Camera) GetConvergencePoint() *geom3d.Vector3 {
	return c.convergencePoint
}
