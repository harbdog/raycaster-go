package engine

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"path/filepath"
	"raycaster-go/engine/raycaster"
	"runtime"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	// ebiten constants
	screenWidth  = 1024
	screenHeight = 700

	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texSize = 256
)

// Game - This is the main type for your game.
type Game struct {
	//--create slicer and declare slices--//
	slicer *raycaster.TextureHandler
	slices []*image.Rectangle

	//--viewport and width / height--//
	view   *ebiten.Image
	width  int
	height int

	//--define camera--//
	camera *raycaster.Camera

	//--graphics manager and sprite batch--//
	spriteBatch *SpriteBatch

	textures [5]*ebiten.Image

	//--test texture--//
	floor *ebiten.Image
	sky   *ebiten.Image

	//-test effect--//
	//Effect effect;

	//test sprite
	sprite *ebiten.Image

	//--array of levels, levels reffer to "floors" of the world--//
	levels []*raycaster.Level
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
	// initialize Game object
	g := new(Game)

	g.width = screenWidth
	g.height = screenHeight

	g.slicer = raycaster.NewTextureHandler(texSize)

	//--init texture slices--//
	g.slices = g.slicer.GetSlices()

	//--inits the levels--//
	g.levels = g.createLevels(4)

	//--init camera--//
	g.camera = raycaster.NewCamera(g.width, g.height, texSize, g.slices, g.levels)

	return g
}

// loadContent will be called once per game and is the place to load
// all of your content.
func (g *Game) loadContent() {
	// Create a new SpriteBatch, which can be used to draw textures.
	g.spriteBatch = &SpriteBatch{g: g}

	// TODO: use this.Content to load your game content here
	g.textures[0], _, _ = getTextureFromFile("stone.png")
	g.textures[1], _, _ = getTextureFromFile("left_bot_house.png")
	g.textures[2], _, _ = getTextureFromFile("right_bot_house.png")
	g.textures[3], _, _ = getTextureFromFile("left_top_house.png")
	g.textures[4], _, _ = getTextureFromFile("right_top_house.png")

	g.floor, _, _ = getTextureFromFile("floor.png")
	g.sky, _, _ = getTextureFromFile("sky.png")
}

func getTextureFromFile(texFile string) (*ebiten.Image, image.Image, error) {
	resourcePath := filepath.Join("engine", "content", "textures")
	eImg, iImg, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile), ebiten.FilterNearest)
	if err != nil {
		log.Fatal(err)
	}
	return eImg, iImg, err
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	// load content once when first run
	g.loadContent()

	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	if err := ebiten.Run(g.Update, g.width, g.height, 1, "Raycaster-Go"); err != nil {
		log.Fatal(err)
	}
}

// Update - Allows the game to run logic such as updating the world,
// checking for collisions, gathering input, and playing audio.
func (g *Game) Update(screen *ebiten.Image) error {
	g.view = screen

	// Perform logical updates
	mx, my := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		fmt.Printf("mouse left clicked: (%v, %v)\n", mx, my)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		fmt.Printf("mouse right clicked: (%v, %v)\n", mx, my)
	}

	// TODO: Add your update logic here
	g.camera.Update()

	if ebiten.IsDrawingSkipped() {
		// When the game is running slowly, the rendering result
		// will not be adopted.
		return nil
	}

	// Render game to screen
	g.draw()

	// TPS counter
	fps := fmt.Sprintf("TPS: %f/%v", ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(g.view, fps)

	return nil
}

func (g *Game) draw() {
	g.view.Clear()

	//--draw sky and floor--//
	texRect := image.Rect(0, 0, texSize, texSize)
	whiteRGBA := &color.RGBA{0, 0, 0, 255}

	// spriteBatch.Begin();

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

	// spriteBatch.End();

	// //--draw walls--//
	// spriteBatch.Begin();

	// for (int x = 0; x < width; x++)
	// {
	// 	for (int i = levels.Length - 1; i >= 0; i--)
	// 	{
	// 		spriteBatch.Draw(textures[levels[i].currTexNum[x]], levels[i].sv[x], levels[i].cts[x], levels[i].st[x]);
	// 	}
	// }

	// spriteBatch.End();
}

//returns an initialised Level struct
func (g *Game) createLevels(numLevels int) []*raycaster.Level {
	var arr []*raycaster.Level
	arr = make([]*raycaster.Level, numLevels)

	for i := 0; i < numLevels; i++ {
		arr[i] = new(raycaster.Level)
		arr[i].Sv = g.sliceView()
		arr[i].Cts = make([]*image.Rectangle, g.width)
		arr[i].St = make([]*color.RGBA, g.width)
		arr[i].CurrTexNum = make([]int, g.width)

		for j := 0; j < cap(arr[i].CurrTexNum); j++ {
			arr[i].CurrTexNum[j] = 1
		}
	}

	return arr
}

// Creates rectangle slices for each x in width.
func (g *Game) sliceView() []*image.Rectangle {
	var arr []*image.Rectangle
	arr = make([]*image.Rectangle, g.width)

	for x := 0; x < g.width; x++ {
		thisRect := image.Rect(x, 0, x+1, g.height)
		arr[x] = &thisRect
	}

	return arr
}

func (s *SpriteBatch) draw(texture *ebiten.Image, destinationRectangle *image.Rectangle, sourceRectangle *image.Rectangle, color *color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear

	// if destinationRectangle is not the same size as sourceRectangle, scale to fit
	var scaleX, scaleY float64 = 1.0, 1.0
	if !destinationRectangle.Eq(*sourceRectangle) {
		sSize := sourceRectangle.Size()
		dSize := destinationRectangle.Size()

		scaleX = float64(dSize.X) / float64(sSize.X)
		scaleY = float64(dSize.Y) / float64(sSize.Y)
	}

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(destinationRectangle.Min.X), float64(destinationRectangle.Min.Y))

	var destTexture *ebiten.Image
	destTexture = texture.SubImage(*sourceRectangle).(*ebiten.Image)

	// TODO: color channel modulation/tinting

	view := s.g.view
	view.DrawImage(destTexture, op)
}
