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
	"github.com/jinzhu/copier"
)

const (
	// ebiten constants
	screenWidth  = 1024
	screenHeight = 768
	renderScale  = 0.75

	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

// Game - This is the main type for your game.
type Game struct {
	//--create slicer and declare slices--//
	tex    *raycaster.TextureHandler
	slices []*image.Rectangle

	//--viewport width / height--//
	width  int
	height int

	player *model.Player

	//--define camera and renderer--//
	camera *raycaster.Camera

	mouseMode      raycaster.MouseMode
	mouseX, mouseY int

	crosshairs *model.Crosshairs

	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *model.Map
	levels       []*raycaster.Level
	spriteLvls   []*raycaster.Level
	floorLvl     *raycaster.HorLevel
	collisionMap []geom.Line

	sprites     map[*model.Sprite]struct{}
	projectiles map[*model.Projectile]struct{}
	effects     map[*model.Effect]struct{}

	preloadedSprites map[string]model.Sprite

	worldMap            [][]int
	mapWidth, mapHeight int
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

	g.tex = raycaster.NewTextureHandler(texWidth)

	//--init texture slices--//
	g.slices = g.tex.GetSlices()

	// load map
	g.mapObj = model.NewMap()

	//--inits the levels--//
	g.levels, g.floorLvl = g.createLevels(4)

	g.collisionMap = g.mapObj.GetCollisionLines(clipDistance)
	g.worldMap = g.mapObj.GetGrid()
	g.mapWidth = len(g.worldMap)
	g.mapHeight = len(g.worldMap[0])

	// load content once when first run
	g.loadContent()

	// make crosshairs its own class since it doesn't behave like other sprites
	g.crosshairs = model.NewCrosshairs(1, 1, 2.0, g.tex.Textures[16], 8, 8, 55, 57, 64)

	// init the sprites
	g.loadSprites()
	g.spriteLvls = g.createSpriteLevels()

	// init mouse movement mode
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	g.mouseMode = raycaster.MouseModeMove
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mapObj, g.slices, g.levels, g.floorLvl, g.spriteLvls, g.tex)
	g.camera.SetFloorTexture(getTextureFromFile("floor.png"))
	g.camera.SetSkyTexture(getTextureFromFile("sky.png"))

	// init player model and initialize camera to their position
	angleDegrees := 90.0
	g.player = model.NewPlayer(10.5, 1.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = clipDistance
	g.updatePlayerCamera(true)

	return g
}

// loadContent will be called once per game and is the place to load
// all of your content.
func (g *Game) loadContent() {
	g.tex.Textures = make([]*ebiten.Image, 32)

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

	// load texture sheets
	g.tex.Textures[15] = getSpriteFromFile("sorcerer_sheet.png")
	g.tex.Textures[16] = getSpriteFromFile("crosshairs_sheet.png")
	g.tex.Textures[17] = getSpriteFromFile("charged_bolt_sheet.png")
	g.tex.Textures[18] = getSpriteFromFile("blue_explosion_sheet.png")

	// just setting the grass texture apart from the rest since it gets special handling
	g.floorLvl.TexRGBA = make([]*image.RGBA, 1)
	g.floorLvl.TexRGBA[0] = getRGBAFromFile("grass.png")
}

func getRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	resourcePath := filepath.Join("engine", "resources", "textures")
	_, tex, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile))
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, texWidth, texWidth))
		// convert into RGBA format
		for x := 0; x < texWidth; x++ {
			for y := 0; y < texWidth; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	return rgba
}

func getTextureFromFile(texFile string) *ebiten.Image {
	resourcePath := filepath.Join("engine", "resources", "textures")
	eImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile))
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func getSpriteFromFile(sFile string) *ebiten.Image {
	resourcePath := filepath.Join("engine", "resources", "sprites")
	eImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, sFile))
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func (g *Game) loadSprites() {
	g.projectiles = make(map[*model.Projectile]struct{}, 1024)
	g.effects = make(map[*model.Effect]struct{}, 1024)
	g.sprites = make(map[*model.Sprite]struct{}, 128)
	g.preloadedSprites = make(map[string]model.Sprite, 16)

	// colors for minimap representation
	blueish := color.RGBA{62, 62, 100, 96}
	brown := color.RGBA{47, 40, 30, 196}
	green := color.RGBA{27, 37, 7, 196}
	orange := color.RGBA{69, 30, 5, 196}
	yellow := color.RGBA{255, 200, 0, 196}

	// preload projectile sprite
	projectileCollisionRadius := 20.0 / 256.0
	g.preloadedSprites["charged_bolt"] = *model.NewAnimatedProjectile(
		0, 0, 0.5, 2, g.tex.Textures[17], blueish,
		48, 1, 256, projectileCollisionRadius,
	).Sprite

	// preload explosion sprite
	g.preloadedSprites["blue_explosion"] = *model.NewAnimatedEffect(
		0, 0, 0.75, 3, g.tex.Textures[18], 5, 3, 256, 0,
	).Sprite

	// sorcerer
	sorcScale := 1.4
	sorcCollisionRadius := sorcScale * 24.0 / 256.0
	sorc := model.NewAnimatedSprite(20, 11.5, sorcScale, 5, g.tex.Textures[15], yellow, 10, 1, 256, sorcCollisionRadius) // FIXME: 256 should come from g.texSize
	// give sprite a sample velocity for movement
	sorc.Angle = geom.Radians(180)
	sorc.Velocity = 0.02
	g.addSprite(sorc)

	// testing sprite scaling
	testScale := 0.5
	g.addSprite(model.NewSprite(10.5, 2.5, testScale, g.tex.Textures[9], green, 256, 0))

	// TODO: speed up init by preloading tree Sprites and copying

	// // line of trees for testing in front of initial view
	// Setting CollisionRadius=0 to disable collision against small trees
	g.addSprite(model.NewSprite(19.5, 11.5, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(17.5, 11.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(15.5, 11.5, 1.0, g.tex.Textures[9], green, 256, 0))
	// // // render a forest!
	g.addSprite(model.NewSprite(11.5, 1.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.5, 1.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(132.5, 1.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(11.5, 2, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.5, 2, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(13.5, 2, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(11.5, 2.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.25, 2.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(13.5, 2.25, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(11.5, 3, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.5, 3, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(13.25, 3, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(10.5, 3.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(11.5, 3.25, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.5, 3.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(13.25, 3.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(10.5, 4, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(11.5, 4, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.5, 4, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(13.5, 4, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(10.5, 4.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(11.25, 4.5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.5, 4.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(13.5, 4.5, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(14.5, 4.25, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(10.5, 5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(11.5, 5, 1.0, g.tex.Textures[9], green, 256, 0))
	g.addSprite(model.NewSprite(12.5, 5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(13.25, 5, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(14.5, 5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(11.5, 5.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(12.5, 5.25, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(13.5, 5.25, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(14.5, 5.5, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(15.5, 5.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(11.5, 6, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(12.5, 6, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(13.25, 6, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(14.25, 6, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(15.5, 6, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(12.5, 6.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(13.5, 6.25, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(14.5, 6.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(12.5, 7, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(13.5, 7, 1.0, g.tex.Textures[10], brown, 256, 0))
	g.addSprite(model.NewSprite(14.5, 7, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(13.5, 7.5, 1.0, g.tex.Textures[14], orange, 256, 0))
	g.addSprite(model.NewSprite(13.5, 8, 1.0, g.tex.Textures[14], orange, 256, 0))
}

func (g *Game) addSprite(sprite *model.Sprite) {
	g.sprites[sprite] = struct{}{}
}

func (g *Game) deleteSprite(sprite *model.Sprite) {
	delete(g.sprites, sprite)

	// TODO: refactor the need for this extra update needed when the sprite list expands/contracts
	g.updateSpriteLevels()
}

func (g *Game) addProjectile(projectile *model.Projectile) {
	g.projectiles[projectile] = struct{}{}

	// TODO: refactor the need for this extra update needed when the projectile list expands
	g.updateSpriteLevels()
}

func (g *Game) deleteProjectile(projectile *model.Projectile) {
	delete(g.projectiles, projectile)

	// TODO: refactor the need for this extra update needed when the projectile list contracts
	g.updateSpriteLevels()
}

func (g *Game) addEffect(effect *model.Effect) {
	g.effects[effect] = struct{}{}

	// TODO: refactor the need for this extra update needed when the projectile list expands
	g.updateSpriteLevels()
}

func (g *Game) deleteEffect(effect *model.Effect) {
	delete(g.effects, effect)

	// TODO: refactor the need for this extra update needed when the projectile list contracts
	g.updateSpriteLevels()
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

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth * renderScale, screenHeight * renderScale
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	numSprites, numProjectiles, numEffects := len(g.sprites), len(g.projectiles), len(g.effects)
	raycastSprites := make([]*model.Sprite, numSprites+numProjectiles+numEffects)
	index := 0
	for sprite := range g.sprites {
		raycastSprites[index] = sprite
		index += 1
	}
	for projectile := range g.projectiles {
		raycastSprites[index] = projectile.Sprite
		index += 1
	}
	for effect := range g.effects {
		raycastSprites[index] = effect.Sprite
		index += 1
	}

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render to screen
	g.camera.Draw(screen)

	// draw minimap
	mm := g.miniMap()
	mmImg := ebiten.NewImageFromImage(mm)
	if mmImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		op.GeoM.Scale(5.0, 5.0)
		op.GeoM.Translate(0, 50)
		screen.DrawImage(mmImg, op)
	}

	// draw crosshairs
	if g.crosshairs != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		crosshairScale := g.crosshairs.Scale
		op.GeoM.Scale(crosshairScale, crosshairScale)
		op.GeoM.Translate(
			float64(g.width)/2-float64(g.crosshairs.W)*crosshairScale/2,
			float64(g.height)/2-float64(g.crosshairs.H)*crosshairScale/2,
		)
		screen.DrawImage(g.crosshairs.GetTexture(), op)

		if g.crosshairs.IsHitIndicatorActive() {
			screen.DrawImage(g.crosshairs.HitIndicator.GetTexture(), op)
			g.crosshairs.Update()
		}
	}
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Perform logical updates
	g.updateProjectiles()
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

	if g.player.WeaponCooldown > 0 {
		g.player.WeaponCooldown -= 1 / float64(ebiten.MaxTPS())
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
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			fmt.Printf("mouse right clicked: (%v, %v)\n", g.mouseX, g.mouseY)
		}

	case raycaster.MouseModeMove:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireTestProjectile()
		}

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

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireTestProjectile()
		}

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
	} else if ebiten.IsKeyPressed(ebiten.KeyZ) {
		g.camera.Prone()
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

	newPos, _, _ := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, true)
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

	newPos, _, _ := g.getValidMove(g.player.Entity, strafeLine.X2, strafeLine.Y2, true)
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

// checks for valid move from current position, returns valid (x, y) position, whether a collision
// was encountered, and a list of entity collisions that may have been encountered
func (g *Game) getValidMove(entity *model.Entity, moveX, moveY float64, checkAlternate bool) (*geom.Vector2, bool, []*model.Entity) {
	newX, newY := moveX, moveY

	posX, posY := entity.Pos.X, entity.Pos.Y
	if posX == newX && posY == moveY {
		return &geom.Vector2{X: posX, Y: posY}, false, []*model.Entity{}
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
	collisionEntities := []*model.Entity{}

	// check wall collisions
	for _, borderLine := range g.collisionMap {
		// TODO: only check intersection of nearby wall cells instead of all of them
		if px, py, ok := geom.LineIntersection(moveLine, borderLine); ok {
			intersectPoints = append(intersectPoints, geom.Vector2{X: px, Y: py})
		}
	}

	// check sprite against player collision
	if entity != g.player.Entity && entity.CollisionRadius > 0 {
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
			collisionEntities = append(collisionEntities, g.player.Entity)
		}
	}

	// check sprite collisions
	for sprite := range g.sprites {
		// TODO: only check intersection of nearby sprites instead of all of them
		if entity == sprite.Entity || entity.CollisionRadius <= 0 || sprite.CollisionRadius <= 0 {
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
			collisionEntities = append(collisionEntities, sprite.Entity)
		}
	}

	isCollision := len(intersectPoints) > 0

	if isCollision {
		if checkAlternate {
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
			if xDiff > 0.001 || yDiff > 0.001 {
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
				return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
			}
		} else {
			// looks like it cannot move
			return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
		}
	}

	if g.worldMap[ix][iy] <= 0 {
		posX = newX
		posY = newY
	} else {
		isCollision = true
	}

	return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
}

func (g *Game) fireTestProjectile() {
	if g.player.WeaponCooldown > 0 {
		return
	}

	// TODO: use player's posZ to adjust projectile shoot height
	// TODO: convert player pitch to angle and then allow projectile to move at up/down angles

	g.player.WeaponCooldown = 0.1

	// fire test projectile spawning near but in front of current player position and angle
	spriteTemplate := g.preloadedSprites["charged_bolt"]
	effectTemplate := g.preloadedSprites["blue_explosion"]
	projectileSprite := &model.Sprite{}
	effectSprite := &model.Sprite{}
	copier.Copy(projectileSprite, spriteTemplate)
	copier.Copy(effectSprite, effectTemplate)

	projectileSpawnDistance := 0.4
	projectileSpawn := geom.LineFromAngle(g.player.Pos.X, g.player.Pos.Y, g.player.Angle, projectileSpawnDistance)
	projectile := &model.Projectile{
		Sprite: projectileSprite,
		ImpactEffect: model.Effect{
			Sprite:    effectSprite,
			LoopCount: 1,
		},
	}
	projectile.Pos = &geom.Vector2{X: projectileSpawn.X2, Y: projectileSpawn.Y2}

	// velocity based on distance per tick (1/60sec)
	projectile.Angle = g.player.Angle
	projectile.Velocity = 0.01

	g.addProjectile(projectile)
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

func (g *Game) updateProjectiles() {
	// Testing animated projectile movement
	for p := range g.projectiles {
		if p.Velocity != 0 {
			vLine := geom.LineFromAngle(p.Pos.X, p.Pos.Y, p.Angle, p.Velocity)

			xCheck := vLine.X2
			yCheck := vLine.Y2

			newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, false)
			if isCollision {
				// for testing purposes, projectiles instantly get deleted when collision occurs
				g.deleteProjectile(p)

				// make a sprite/wall getting hit by projectile cause some visual effect
				if p.ImpactEffect.Sprite != nil {
					// TODO: give impact effect optional ability to have some velocity based on the projectile movement upon impact if it didn't hit a wall
					effect := &model.Effect{}
					copier.Copy(effect, p.ImpactEffect)
					effect.Pos = &geom.Vector2{X: newPos.X, Y: newPos.Y}

					g.addEffect(effect)
				}

				for _, entity := range collisions {
					if entity == g.player.Entity {
						println("ouch!")
					} else {
						// show crosshair hit effect
						g.crosshairs.ActivateHitIndicator(30)
					}
				}
			} else {
				p.Pos = newPos
			}
		}
		p.Update()
	}

	// Testing animated effects (explosions)
	for e := range g.effects {
		e.Update()
		if e.GetLoopCounter() >= e.LoopCount {
			g.deleteEffect(e)
		}
	}
}

func (g *Game) updateSprites() {
	// Testing animated sprite movement
	for s := range g.sprites {
		if s.Velocity != 0 {
			vLine := geom.LineFromAngle(s.Pos.X, s.Pos.Y, s.Angle, s.Velocity)

			xCheck := vLine.X2
			yCheck := vLine.Y2

			newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, false)
			if isCollision {
				// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
				s.Angle = randFloat(-180, 180)
				s.Velocity = randFloat(0.01, 0.03)
			} else {
				s.Pos = newPos
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
	numSprites := len(g.sprites)

	spriteArr := make([]*raycaster.Level, numSprites)

	return spriteArr
}

func (g *Game) updateSpriteLevels() {
	// update empty "level" for all sprites used by camera
	// TODO: this should be refactored so to be not necessary
	numSprites, numProjectiles, numEffects := len(g.sprites), len(g.projectiles), len(g.effects)

	g.spriteLvls = make([]*raycaster.Level, numSprites+numProjectiles+numEffects)
	g.camera.UpdateSpriteLevels(g.spriteLvls)
}
