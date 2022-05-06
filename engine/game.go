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
	mapObj     *raycaster.Map
	levels     []*raycaster.Level
	spriteLvls []*raycaster.Level
	floorLvl   *raycaster.HorLevel

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
				g.camera.Pitch(dy)
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

	// TODO: collision check using raycast method

	g.player.Pos = &geom.Vector2{X: moveLine.X2, Y: moveLine.Y2}
	g.player.Moved = true
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := math.Pi / 2
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	strafeLine := geom.LineFromAngle(g.player.Pos.X, g.player.Pos.Y, g.player.Angle-strafeAngle, math.Abs(sSpeed))

	// TODO: collision check using raycast method

	g.player.Pos = &geom.Vector2{X: strafeLine.X2, Y: strafeLine.Y2}
	g.player.Moved = true
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
			// TODO: use ebiten.CurrentTPS() to determine actual velicity amount to move sprite per tick

			horBounds := float64(s.W/2) / float64(texSize)

			xCheck := int(s.X)
			yCheck := int(s.Y)
			if s.Vx > 0 {
				xCheck = int(s.X + s.Vx + horBounds)
			} else if s.Vx < 0 {
				xCheck = int(s.X + s.Vx - horBounds)
			}

			if s.Vy > 0 {
				yCheck = int(s.Y + s.Vy + horBounds)
			} else if s.Vy < 0 {
				yCheck = int(s.Y + s.Vy - horBounds)
			}

			if g.mapObj.GetAt(xCheck, yCheck) == 0 {
				// simple collision check to prevent phasing through walls
				s.X += s.Vx
				s.Y += s.Vy
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
