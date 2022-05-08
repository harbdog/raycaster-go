package engine

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"path/filepath"
	"runtime"

	_ "image/png"

	"raycaster-go/engine/geom"
	"raycaster-go/engine/model"
	"raycaster-go/engine/raycaster"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	// ebiten constants
	screenWidth  = 1024
	screenHeight = 768
	renderScale  = 0.75

	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texSize = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

// Game - This is the main type for your game.
type Game struct {
	//--create slicer and declare slices--//
	tex    *raycaster.TextureHandler
	slices []*image.Rectangle

	//--viewport and width / height--//
	view   *ebiten.Image
	width  int
	height int

	player *Player

	//--define camera--//
	camera *raycaster.Camera

	mouseMode      raycaster.MouseMode
	mouseX, mouseY int

	//--graphics manager and sprite batch--//
	spriteBatch *SpriteBatch

	//--test texture--//
	floor *ebiten.Image
	sky   *ebiten.Image

	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *raycaster.Map
	levels       []*raycaster.Level
	spriteLvls   []*raycaster.Level
	floorLvl     *raycaster.HorLevel
	collisionMap []geom.Line

	worldMap            [][]int
	mapWidth, mapHeight int

	// for debugging
	DebugX    int
	DebugY    int
	DebugOnce bool
}

// SpriteBatch - converted C# method Graphics.SpriteBatch
// Enables a group of sprites to be drawn using the same settings.
type SpriteBatch struct {
	g *Game
}

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	fmt.Printf("Initializing Game\n")

	// initialize Game object
	g := new(Game)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Raycaster-Go")

	// use scale to keep the desired window width and height
	g.width = int(math.Floor(float64(screenWidth) * renderScale))
	g.height = int(math.Floor(float64(screenHeight) * renderScale))

	g.tex = raycaster.NewTextureHandler(texSize)

	//--init texture slices--//
	g.slices = g.tex.GetSlices()

	// load map
	g.mapObj = raycaster.NewMap(g.tex)

	//--inits the levels--//
	g.levels, g.floorLvl = g.createLevels(4)

	g.collisionMap = g.mapObj.GetCollisionLines(clipDistance)
	g.worldMap = g.mapObj.GetGrid()
	g.mapWidth = len(g.worldMap)
	g.mapHeight = len(g.worldMap[0])

	// load content once when first run
	g.loadContent()

	// init the sprites
	g.mapObj.LoadSprites()
	g.spriteLvls = g.createSpriteLevels()

	// give sprite a sample velocity for movement
	s := g.mapObj.GetSprite(0)
	s.Vx = -0.02
	// give sprite custom bounds for collision instead of using image bounds
	s.W = int(s.Scale * 85)
	s.H = int(s.Scale * 126)

	// init mouse movement mode
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	g.mouseMode = raycaster.MouseModeMove
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera--//
	g.camera = raycaster.NewCamera(g.width, g.height, texSize, g.mapObj, g.slices, g.levels, g.floorLvl, g.spriteLvls, g.tex)

	// init player model and initialize camera to their position
	angleDegrees := 90.0
	g.player = NewPlayer(10.5, 1.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = clipDistance
	g.updatePlayerCamera(true)

	// for debugging
	g.DebugX = -1
	g.DebugY = -1

	return g
}

// loadContent will be called once per game and is the place to load
// all of your content.
func (g *Game) loadContent() {
	// Create a new SpriteBatch, which can be used to draw textures.
	g.spriteBatch = &SpriteBatch{g: g}

	g.tex.Textures = make([]*ebiten.Image, 16)

	// load wall textures
	g.tex.Textures[0] = getTextureFromFile("stone.png")
	g.tex.Textures[1] = getTextureFromFile("left_bot_house.png")
	g.tex.Textures[2] = getTextureFromFile("right_bot_house.png")
	g.tex.Textures[3] = getTextureFromFile("left_top_house.png")
	g.tex.Textures[4] = getTextureFromFile("right_top_house.png")

	// separating sprites out a bit from wall textures
	g.tex.Textures[9] = getSpriteFromFile("tree_09.png")
	g.tex.Textures[10] = getSpriteFromFile("tree_10.png")
	g.tex.Textures[14] = getSpriteFromFile("tree_14.png")

	g.tex.Textures[15] = getSpriteFromFile("sorcerer_sheet.png")

	g.floor = getTextureFromFile("floor.png")
	g.sky = getTextureFromFile("sky.png")

	// just setting the grass texture apart from the rest since it gets special handling
	g.floorLvl.TexRGBA = make([]*image.RGBA, 1)
	g.floorLvl.TexRGBA[0] = getRGBAFromFile("grass.png")
}

func getRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	resourcePath := filepath.Join("engine", "content", "textures")
	_, tex, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile))
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, texSize, texSize))
		// convert into RGBA format
		for x := 0; x < texSize; x++ {
			for y := 0; y < texSize; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	return rgba
}

func getTextureFromFile(texFile string) *ebiten.Image {
	resourcePath := filepath.Join("engine", "content", "textures")
	eImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile))
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func getSpriteFromFile(sFile string) *ebiten.Image {
	resourcePath := filepath.Join("engine", "content", "sprites")
	eImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, sFile))
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Perform logical updates
	g.updateSprites()

	// handle input
	g.handleInput()

	// handle player camera movement
	g.updatePlayerCamera(false)

	return nil
}

func (g *Game) handleInput() {
	forward := false
	backward := false
	rotLeft := false
	rotRight := false

	moveModifier := 1.0
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveModifier = 2.0
	}

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyControl):
		if g.mouseMode != raycaster.MouseModeCursor {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
			g.mouseMode = raycaster.MouseModeCursor
		}

	case ebiten.IsKeyPressed(ebiten.KeyAlt):
		if g.mouseMode != raycaster.MouseModeMove {
			ebiten.SetCursorMode(ebiten.CursorModeCaptured)
			g.mouseMode = raycaster.MouseModeMove
			g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
		}

	case g.mouseMode != raycaster.MouseModeLook:
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		g.mouseMode = raycaster.MouseModeLook
		g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
	}

	switch g.mouseMode {
	case raycaster.MouseModeCursor:
		g.mouseX, g.mouseY = ebiten.CursorPosition()
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			fmt.Printf("mouse left clicked: (%v, %v)\n", g.mouseX, g.mouseY)

			// using left click for debugging graphical issues
			if g.DebugX == -1 && g.DebugY == -1 {
				// only allow setting once between clears to debounce
				g.DebugX = g.mouseX
				g.DebugY = g.mouseY
				g.DebugOnce = true
			}
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			fmt.Printf("mouse right clicked: (%v, %v)\n", g.mouseX, g.mouseY)

			// using right click to clear the debugging flag
			g.DebugX = -1
			g.DebugY = -1
			g.DebugOnce = false
		}

	case raycaster.MouseModeMove:
		x, y := ebiten.CursorPosition()
		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				g.Rotate(0.005 * float64(dx) * moveModifier)
			}

			if dy != 0 {
				g.Move(0.01 * float64(dy) * moveModifier)
			}
		}
	case raycaster.MouseModeLook:
		x, y := ebiten.CursorPosition()
		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				g.Rotate(0.005 * float64(dx) * moveModifier)
			}

			if dy != 0 {
				g.camera.PitchCamera(dy)
			}
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		rotLeft = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		rotRight = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		forward = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		backward = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.camera.Crouch()
	} else if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.camera.Jump()
	} else {
		g.camera.Stand()
	}

	if forward {
		g.Move(0.06 * moveModifier)
	} else if backward {
		g.Move(-0.06 * moveModifier)
	}

	if g.mouseMode == raycaster.MouseModeLook || g.mouseMode == raycaster.MouseModeMove {
		// strafe instead of rotate
		if rotLeft {
			g.Strafe(-0.05 * moveModifier)
		} else if rotRight {
			g.Strafe(0.05 * moveModifier)
		}
	} else {
		if rotLeft {
			g.Rotate(0.03 * moveModifier)
		} else if rotRight {
			g.Rotate(-0.03 * moveModifier)
		}
	}
}

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	moveLine := geom.LineFromAngle(g.player.Pos.X, g.player.Pos.Y, g.player.Angle, mSpeed)

	newPos := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, true)
	if !newPos.Equals(g.player.Pos) {
		g.player.Pos = newPos
		g.player.Moved = true
	}
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := math.Pi / 2
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	strafeLine := geom.LineFromAngle(g.player.Pos.X, g.player.Pos.Y, g.player.Angle-strafeAngle, math.Abs(sSpeed))

	newPos := g.getValidMove(g.player.Entity, strafeLine.X2, strafeLine.Y2, true)
	if !newPos.Equals(g.player.Pos) {
		g.player.Pos = newPos
		g.player.Moved = true
	}
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	g.player.Angle += rSpeed

	pi2 := 2 * math.Pi
	if g.player.Angle >= pi2 {
		g.player.Angle = pi2 - g.player.Angle
	} else if g.player.Angle <= -pi2 {
		g.player.Angle = g.player.Angle + pi2
	}

	g.player.Moved = true
}

// checks for valid move from current position, returns valid (x, y) position
func (g *Game) getValidMove(entity *model.Entity, moveX, moveY float64, checkAlternate bool) *geom.Vector2 {
	newX, newY := moveX, moveY

	posX, posY := entity.Pos.X, entity.Pos.Y
	if posX == newX && posY == moveY {
		return &geom.Vector2{X: posX, Y: posY}
	}

	ix := int(newX)
	iy := int(newY)

	// prevent index out of bounds errors
	switch {
	case ix < 0 || newX < 0:
		newX = clipDistance
		ix = 0
	case ix >= g.mapWidth:
		newX = float64(g.mapWidth) - clipDistance
		ix = int(newX)
	}

	switch {
	case iy < 0 || newY < 0:
		newY = clipDistance
		iy = 0
	case iy >= g.mapHeight:
		newY = float64(g.mapHeight) - clipDistance
		iy = int(newY)
	}

	moveLine := geom.Line{X1: posX, Y1: posY, X2: newX, Y2: newY}
	entityCircle := geom.Circle{X: newX, Y: newY, Radius: entity.CollisionRadius}

	intersectPoints := []geom.Vector2{}

	// check wall collisions
	for _, borderLine := range g.collisionMap {
		// TODO: only check intersection of nearby wall cells instead of all of them
		if px, py, ok := geom.LineIntersection(moveLine, borderLine); ok {
			intersectPoints = append(intersectPoints, geom.Vector2{X: px, Y: py})
		}
	}

	// check sprite against player collision
	if entity != g.player.Entity {
		// TODO: only check for collision if player is somewhat nearby
		collisionRadius := g.player.CollisionRadius + entity.CollisionRadius
		collisionCircle := geom.Circle{X: g.player.Pos.X, Y: g.player.Pos.Y, Radius: collisionRadius}

		_, isCollision := collisionCircle.CircleCollision(&entityCircle)
		if isCollision {
			// determine new position which would be the center point of the moment of collision
			angle := moveLine.Angle()

			// create line using collision radius to determine center point for intersection
			collisionLine := geom.LineFromAngle(posX, posY, angle, collisionRadius)
			intersectPoints = append(intersectPoints, geom.Vector2{X: collisionLine.X2, Y: collisionLine.Y2})
		}
	}

	// check sprite collisions
	sprites := g.mapObj.GetSprites()
	for _, sprite := range sprites {
		// TODO: only check intersection of nearby sprites instead of all of them
		if entity == sprite.Entity {
			continue
		}

		// FIXME: need some way to let a moving sprite out of the inside of a collision radius without letting it in
		collisionRadius := sprite.CollisionRadius + entity.CollisionRadius
		collisionCircle := geom.Circle{X: sprite.Pos.X, Y: sprite.Pos.Y, Radius: collisionRadius}

		_, isCollision := collisionCircle.CircleCollision(&entityCircle)
		if isCollision {
			// determine new position which would be the center point of the moment of collision
			angle := moveLine.Angle()

			// create line using collision radius to determine center point for intersection
			collisionLine := geom.LineFromAngle(posX, posY, angle, collisionRadius)
			intersectPoints = append(intersectPoints, geom.Vector2{X: collisionLine.X2, Y: collisionLine.Y2})
		}
	}

	if len(intersectPoints) > 0 {
		// find the point closest to the start position
		min := math.Inf(1)
		minI := -1
		for i, p := range intersectPoints {
			d2 := geom.Distance2(posX, posY, p.X, p.Y)
			if d2 < min {
				min = d2
				minI = i
			}
		}

		// use the closest intersecting point to determine a safe distance to make the move
		moveLine = geom.Line{X1: posX, Y1: posY, X2: intersectPoints[minI].X, Y2: intersectPoints[minI].Y}
		dist := math.Sqrt(min)
		angle := moveLine.Angle()

		// generate new move line using calculated angle and safe distance from intersecting point
		moveLine = geom.LineFromAngle(posX, posY, angle, dist-0.01)

		newX, newY = moveLine.X2, moveLine.Y2

		// if either X or Y direction was already intersecting, attempt move only in the adjacent direction
		xDiff := math.Abs(newX - posX)
		yDiff := math.Abs(newY - posY)
		if checkAlternate && (xDiff > 0.001 || yDiff > 0.001) {
			switch {
			case xDiff <= 0.001:
				// no more room to move in X, try to move only Y
				// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
				// 	c.pos.X, c.pos.Y, moveX, moveY, posX, moveY)
				return g.getValidMove(entity, posX, moveY, false)
			case yDiff <= 0.001:
				// no more room to move in Y, try to move only X
				// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
				// 	c.pos.X, c.pos.Y, moveX, moveY, moveX, posY)
				return g.getValidMove(entity, moveX, posY, false)
			default:
				// try the new position
				// TODO: need some way to try a potentially valid shorter move without checkAlternate while also avoiding infinite loop
				return g.getValidMove(entity, newX, newY, false)
			}
		} else {
			// looks like it cannot move
			return &geom.Vector2{X: posX, Y: posY}
		}
	}

	if g.worldMap[ix][iy] <= 0 {
		posX = newX
		posY = newY
	}

	return &geom.Vector2{X: posX, Y: posY}
}

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.Moved = false

	playerPos := g.player.Pos.Copy()
	playerDir := g.camera.GetVecForAngle(g.player.Angle)
	g.camera.SetPosition(playerPos)
	g.camera.SetDirection(playerDir)
	g.camera.SetPlane(g.camera.GetVecForFov(playerDir))
}

func (g *Game) updateSprites() {
	// Testing animated sprite movement
	sprites := g.mapObj.GetSprites()

	for _, s := range sprites {
		if s.Vx != 0 || s.Vy != 0 {
			xCheck := s.Pos.X + s.Vx
			yCheck := s.Pos.Y + s.Vy

			newPos := g.getValidMove(s.Entity, xCheck, yCheck, false)
			if !newPos.NearlyEquals(s.Pos, 0.00001) {
				s.Pos = newPos
			} else {
				// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
				s.Vx = randFloat(-0.03, 0.03)
				s.Vy = randFloat(-0.03, 0.03)
			}
		}
		s.Update()
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

//returns initialised Level structs
func (g *Game) createLevels(numLevels int) ([]*raycaster.Level, *raycaster.HorLevel) {
	levelArr := make([]*raycaster.Level, numLevels)

	for i := 0; i < numLevels; i++ {
		levelArr[i] = new(raycaster.Level)
		levelArr[i].Sv = raycaster.SliceView(g.width, g.height)
		levelArr[i].Cts = make([]*image.Rectangle, g.width)
		levelArr[i].St = make([]*color.RGBA, g.width)
		levelArr[i].CurrTex = make([]*ebiten.Image, g.width)
	}

	horizontalLevel := new(raycaster.HorLevel)
	horizontalLevel.Clear(g.width, g.height)

	return levelArr, horizontalLevel
}

func (g *Game) createSpriteLevels() []*raycaster.Level {
	// create empty "level" for all sprites to render using similar slice methods as walls
	numSprites := g.mapObj.GetNumSprites()

	spriteArr := make([]*raycaster.Level, numSprites)

	return spriteArr
}

// DebugPrintfOnce prints info to screen only one time until g.DebugFlag cleared again
func (g *Game) DebugPrintfOnce(format string, a ...interface{}) {
	if g.DebugOnce {
		fmt.Printf(format, a...)
	}
}
