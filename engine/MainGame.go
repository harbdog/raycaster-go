package engine

import (
	"fmt"
	"log"
	"runtime"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	screenWidth   = 640
	screenHeight  = 480
	fontSize      = 32
	smallFontSize = fontSize / 2
)

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

type Game struct {
	mode          Mode
	ui            *UI
	gameoverCount int
}

type UI struct {
	// render screen
	screen *ebiten.Image

	// Camera
	cameraX int
	cameraY int
}

type Point struct {
	X, Y float64
}

func NewGame() *Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) init() {
	g.ui = &UI{}
	g.ui.SetCamera(-240, 0)
	g.ui.Init()
}

func (g *Game) Run() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	if err := ebiten.Run(g.Update, screenWidth, screenHeight, 1, "PixelMek-Go"); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update(screen *ebiten.Image) error {

	// Perform logical updates
	g.ui.Update(screen)

	if ebiten.IsDrawingSkipped() {
		// When the game is running slowly, the rendering result
		// will not be adopted.
		return nil
	}

	// Render game to screen
	g.ui.Draw(screen)

	return nil
}

func (ui *UI) Init() {
	// setup test sprite
}

func (ui *UI) SetCamera(x int, y int) {
	ui.cameraX = x
	ui.cameraY = y
}

func (ui *UI) Update(screen *ebiten.Image) error {
	ui.screen = screen

	mx, my := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		fmt.Printf("mouse left clicked: (%v, %v)\n", mx, my)
		// gPoint := Point{float64(mx + ui.cameraX), float64(my + ui.cameraY)}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		fmt.Printf("mouse right clicked: (%v, %v)\n", mx, my)
		// gPoint := Point{float64(mx + ui.cameraX), float64(my + ui.cameraY)}
	}

	// Update sprite positions
	// Update()

	return nil
}

func (ui *UI) Draw(screen *ebiten.Image) error {
	ui.screen = screen

	// Render sprites
	// Draw(ui)

	// TPS counter
	fps := fmt.Sprintf("TPS: %f/%v", ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(screen, fps)

	return nil
}
