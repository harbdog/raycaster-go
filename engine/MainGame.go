package engine

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"path/filepath"
	"raycaster-go/engine/raycaster"
	"runtime"

	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	// ebiten constants
	screenWidth  = 1024
	screenHeight = 700
	screenScale  = 1.0

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

	//--define camera--//
	camera *raycaster.Camera

	//--graphics manager and sprite batch--//
	spriteBatch *SpriteBatch

	//--test texture--//
	floor *ebiten.Image
	sky   *ebiten.Image

	//--array of levels, levels reffer to "floors" of the world--//
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
	g.width = int(math.Floor(float64(screenWidth) / screenScale))
	g.height = int(math.Floor(float64(screenHeight) / screenScale))

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

	//--init camera--//
	g.camera = raycaster.NewCamera(g.width, g.height, texSize, g.mapObj, g.slices, g.levels, g.floorLvl, g.spriteLvls, g.tex)

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

	// TODO: use loadContent to load your game content here
	g.tex.Textures = make([]*ebiten.Image, 16)

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

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Perform logical updates
	g.camera.Update()

	// handle input
	g.handleInput()

	return nil
}

func (g *Game) handleInput() {
	mx, my := ebiten.CursorPosition()

	forward := false
	backward := false
	rotLeft := false
	rotRight := false

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		fmt.Printf("mouse left clicked: (%v, %v)\n", mx, my)

		// using left click for debugging graphical issues
		if g.DebugX == -1 && g.DebugY == -1 {
			// only allow setting once between clears to debounce
			g.DebugX = mx
			g.DebugY = my
			g.DebugOnce = true
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		fmt.Printf("mouse right clicked: (%v, %v)\n", mx, my)

		// using right click to clear the debugging flag
		g.DebugX = -1
		g.DebugY = -1
		g.DebugOnce = false
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

	if forward {
		g.camera.Move(0.06)
	} else if backward {
		g.camera.Move(-0.06)
	}

	if ebiten.IsKeyPressed(ebiten.KeyAlt) {
		// strafe instead of rotate
		if rotLeft {
			g.camera.Strafe(-0.05)
		} else if rotRight {
			g.camera.Strafe(0.05)
		}
	} else {
		if rotLeft {
			g.camera.Rotate(0.03)
		} else if rotRight {
			g.camera.Rotate(-0.03)
		}
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	g.view = screen
	g.view.Clear()

	//--draw basic sky and floor--//
	texRect := image.Rect(0, 0, texSize, texSize)
	whiteRGBA := &color.RGBA{255, 255, 255, 255}

	// spriteBatch.Draw(floor,
	//    new Rectangle(0, (int)(height * 0.5f), width, (int)(height * 0.5f)),
	//    new Rectangle(0, 0, texSize, texSize),
	//    Color.White);
	floorRect := image.Rect(0, int(float64(g.height)*0.5), g.width, 2*int(float64(g.height)*0.5))
	g.spriteBatch.draw(g.floor, &floorRect, &texRect, whiteRGBA)

	// spriteBatch.Draw(sky,
	//    new Rectangle(0, 0, width, (int)(height * 0.5f)),
	//    new Rectangle(0, 0, texSize, texSize),
	//    Color.White);
	skyRect := image.Rect(0, 0, g.width, int(float64(g.height)*0.5))
	g.spriteBatch.draw(g.sky, &skyRect, &texRect, whiteRGBA)

	//--draw walls--//
	for x := 0; x < g.width; x++ {
		for i := cap(g.levels) - 1; i >= 0; i-- {
			g.spriteBatch.draw(g.levels[i].CurrTex[x], g.levels[i].Sv[x], g.levels[i].Cts[x], g.levels[i].St[x])
		}
	}

	// draw textured floor
	floorImg := ebiten.NewImageFromImage(g.floorLvl.HorBuffer)
	if floorImg == nil {
		log.Fatal("floorImg is nil")
	} else {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		g.view.DrawImage(floorImg, op)
	}

	// draw sprites
	for x := 0; x < g.width; x++ {
		for i := 0; i < cap(g.spriteLvls); i++ {
			spriteLvl := g.spriteLvls[i]
			if spriteLvl == nil {
				continue
			}

			texture := spriteLvl.CurrTex[x]
			if texture != nil {
				g.spriteBatch.draw(texture, spriteLvl.Sv[x], spriteLvl.Cts[x], spriteLvl.St[x])
			}
		}
	}

	if g.DebugOnce {
		// end DebugOnce after one loop
		g.DebugOnce = false
	}

	// draw for debugging
	if g.DebugX >= 0 && g.DebugY >= 0 {
		fX := float64(g.DebugX)
		fY := float64(g.DebugY)
		// draw a red translucent dot at the debug point
		ebitenutil.DrawLine(g.view, fX-0.5, fY-0.5, fX+0.5, fY+0.5, color.RGBA{255, 0, 0, 150})

		// draw two red vertical lines focusing on point
		ebitenutil.DrawLine(g.view, fX-0.5, fY+5, fX+0.5, fY+25, color.RGBA{255, 0, 0, 150})
		ebitenutil.DrawLine(g.view, fX-0.5, fY-25, fX+0.5, fY-5, color.RGBA{255, 0, 0, 150})
	}

	// TPS counter
	fps := fmt.Sprintf("TPS: %f/%v", ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(g.view, fps)
}

//returns initialised Level structs
func (g *Game) createLevels(numLevels int) ([]*raycaster.Level, *raycaster.HorLevel) {
	var levelArr []*raycaster.Level
	levelArr = make([]*raycaster.Level, numLevels)

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

	var spriteArr []*raycaster.Level
	spriteArr = make([]*raycaster.Level, numSprites)

	return spriteArr
}

func (s *SpriteBatch) draw(texture *ebiten.Image, destinationRectangle *image.Rectangle, sourceRectangle *image.Rectangle, color *color.RGBA) {
	if texture == nil || destinationRectangle == nil || sourceRectangle == nil {
		return
	}

	if sourceRectangle.Min.X == 0 {
		// fixes subImage from clipping at edges of textures which can cause gaps
		sourceRectangle.Min.X++
		sourceRectangle.Max.X++
	}

	// if destinationRectangle is not the same size as sourceRectangle, scale to fit
	var scaleX, scaleY float64 = 1.0, 1.0
	if !destinationRectangle.Eq(*sourceRectangle) {
		sSize := sourceRectangle.Size()
		dSize := destinationRectangle.Size()

		scaleX = float64(dSize.X) / float64(sSize.X)
		scaleY = float64(dSize.Y) / float64(sSize.Y)
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(destinationRectangle.Min.X), float64(destinationRectangle.Min.Y))

	var destTexture *ebiten.Image
	destTexture = texture.SubImage(*sourceRectangle).(*ebiten.Image)

	if color != nil {
		// color channel modulation/tinting
		op.ColorM.Scale(float64(color.R)/255, float64(color.G)/255, float64(color.B)/255, float64(color.A)/255)
	}

	if s.g.DebugX > destinationRectangle.Min.X && s.g.DebugX <= destinationRectangle.Max.X &&
		s.g.DebugY > destinationRectangle.Min.Y && s.g.DebugY <= destinationRectangle.Max.Y {

		for texNum, tex := range s.g.tex.Textures {
			if tex == texture {
				s.g.DebugPrintfOnce("[draw@%v,%v]: %v | %v < %v * %v,%v\n", s.g.DebugX, s.g.DebugY, destinationRectangle, texNum, sourceRectangle, scaleX, scaleY)
				return
			}
		}
	}

	view := s.g.view
	view.DrawImage(destTexture, op)
}

// DebugPrintfOnce prints info to screen only one time until g.DebugFlag cleared again
func (g *Game) DebugPrintfOnce(format string, a ...interface{}) {
	if g.DebugOnce {
		fmt.Printf(format, a...)
	}
}
