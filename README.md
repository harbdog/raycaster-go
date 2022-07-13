# raycaster-go

Golang raycasting engine using the [Ebitengine 2D Game Library](https://github.com/hajimehoshi/ebiten) to render a
3-Dimensional perspective in a 2-Dimensional view. It was originally converted from the C# based
[OwlRaycastEngine](https://github.com/Owlzy/OwlRaycastEngine), which in turn was created based on a
[raycasting tutorial](https://lodev.org/cgtutor/raycasting.html).

To see it in action visit the [YouTube playlist](https://www.youtube.com/playlist?list=PLOINtzQqJWIh8OQsvYAahr2yuAF5VLk38).

![Screenshot](https://raw.githubusercontent.com/harbdog/raycaster-go-demo/main/docs/images/screenshot.jpg)

## Demo

The [raycaster-go-demo](https://github.com/harbdog/raycaster-go-demo) project is available as an example of how to
use the raycaster-go engine as a module.

## Developer setup

To get started with your own project using the raycaster-go engine as a Go module:

1. Download, install, and setup Golang https://golang.org/dl/
2. Setup your project to use go modules: `go mod init github.com/yourname/yourgame` 
3. Download the raycaster-go module: `go get github.com/harbdog/raycaster-go`

**NOTE**: Depending on the OS, the Ebiten game library may have
[additional dependencies to install](https://ebiten.org/documents/install.html).

## Creating your own raycaster project

You will first want to become familiar with how to use the [Ebitengine 2D Game Library](https://ebiten.org/).
It has all of the APIs needed to render images on a 2D canvas, handle inputs from the player, and even play sounds.
The [raycaster-go-demo](https://github.com/harbdog/raycaster-go-demo) is available for reference.

## Ebitengine interfaces

Just like any other Ebitengine game, there are a few [game interface functions](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#Game) 
required from your game implementation:

- `Update`
- `Draw`
- `Layout`

Refer to the [Ebitengine tour pages](https://ebiten.org/tour/hello_world.html) for detailed explanation
about the usage of these interface functions needed for basic game flow.

## Raycaster-go interfaces

There are additional raycaster-go specific interfaces that will be required to render the map levels,
wall textures, and sprites.

### [Map interfaces](map.go)

Interface functions required to provide layout of wall positions on the 2-dimensional array representing the game map.

`Level(levelNum int) [][]int`
- Needs to return the 2-dimensional array map of a given level index.
- The first order array is used as the X-axis, and the second as the Y-axis.
- The `int` value at each X/Y array index pair represents whether a wall is present at that map coordinate.

  ` > 0 `: indicates presence of a wall at that map coordinate.

  `<= 0 `: indicates absence of walls at that map coordinate.

- Length of X and Y arrays do not need to match within a level, can be square or rectangle map layout.
- Each level of the map must have arrays of the same size.


`NumLevels() int`
- Needs to return the number of vertical/elevation levels.

### [TextureHandler interfaces](texture.go)

Interface functions required for rendering texture images for the walls and floor.

`TextureAt(x, y, levelNum, side int) *ebiten.Image`
- Needs to return an [ebiten.Image](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#Image) as
  the wall texture found at the indicated X/Y map coordinate and level number.
- `side` is currently provided as either `1` or `0` indicating the texture viewed from the X or Y direction,
  respectively. This value can be used, if desired, to have a different texture image representing
  alternate sides of the wall.
- The size of the texture image returned will need to match the texture size (`texSize`) provided
  to the `NewCamera` function. For example `texSize: 256` requires all wall textures to be `256x256` pixels in size.

`FloorTextureAt(x, y int) *image.RGBA`
- Used to return an [image.RGBA](https://pkg.go.dev/image#RGBA) to be used as the repeating floor texture
  at the indicated X/Y map coordinate.
- It can also return `nil` to only render the non-repeating floor texture provided to
  the `camera.SetFloorTexture` function.

### [Sprite interfaces](sprite.go)

Interface functions required to determine sprite images and positions to render in game.

`Pos() *geom.Vector2`
- Needs to return a [geom.Vector2](geom/geometry.go) containing the X/Y map position of the sprite.

`PosZ() float64`
- Needs to return the Z-position of the sprite.
- A value of `0.5` represents positioning the center of the sprite at the center of the first elevation level.

`Scale() float64`
- Needs to return the scale factor of the sprite.
- A value of `1.0` indicates no scaling.
- Note that scaling below or above `1.0 ` will likely require also providing a non-zero value from `VerticalOffset()`.

`VerticalOffset() float64`
- Needs to return the vertical offset to displace the sprite image.
- A value of `0.0` indicates no offset.
- Typically only needs to provide a non-zero value when `Scale()` returns a value other than `1.0`,
  if you find the sprite image is floating above or sinking below the floor when scaled.

`Texture() *ebiten.Image`
- Needs to return the current image to render.
- Called during the `camera.Update` function call.
- Calls to your own update functions will be responsible for changing the image pointer returned
  each tick if animations or other sprite image changes are desired.

`TextureRect() image.Rectangle`
- Needs to return an [image.Rectangle](https://pkg.go.dev/image#Rectangle) representing the texture sheet position
  and area to render for the image currently returned by `Texture()`.
- If the image source only contains a single image, just set to origin position along with
  the width and height of the image:

  ```golang
  return image.Rect(0, 0, imageWidth, imageHeight)
  ```

## Raycaster-go camera

After implementing all required interface functions, the last step is to initialize an instance of `raycaster.Camera`
and make the function calls needed to update and draw during your game loop.

`func NewCamera(width int, height int, texSize int, mapObj Map, tex TextureHandler) *Camera`
- `width`, `height`: the window/viewport size.
- `texSize`: the pixel width and height of all textures.
- `mapObj`: struct implementing all required [Map interfaces](map.go).
- `tex`: struct implementing all required [TextureHandler interfaces](texture.go).

`camera.SetPosition(pos *geom.Vector2)`
- Sets the camera X/Y map position as [geom.Vector2](geom/geometry.go).

`camera.SetPositionZ`
- Sets the camera Z position (where `0.5` represents the middle of the first elevation level).

`camera.SetHeadingAngle`
- Sets the camera heading angle (in radians, where `0.0` is in the positive X-axis with no Y-axis direction).

`camera.SetPitchAngle`
- Sets the camera pitch angle (in radians, where `0.0` is looking straight ahead).

`camera.SetFloorTexture(floor *ebiten.Image)`
- Sets the non-repeating simple floor texture.
- Only shown when `TextureHandler.FloorTexture()` interface returns `nil`, and for areas outside of map bounds.

`camera.SetSkyTexture(sky *ebiten.Image)`
- Sets the non-repeating simple skybox texture.

`camera.Update(sprites []Sprite)`
- `sprites`: an array of structs implementing all required [Sprite interfaces](sprite.go).
- Called during your game's implementation of `Draw(screen *ebiten.Image)` to perform raycasting updates.
- Must be called before `camera.Draw`.

`camera.Draw(screen *ebiten.Image)`
- Called during your game's implementation of `Draw(screen *ebiten.Image)` to render the raycasted levels and sprites.
- Must be called after `camera.Update`.

### Optional camera functions

`camera.SetRenderDistance(distance float64)`
- Sets maximum distance to render raycasted floors, walls, and objects (-1 for practically inf)
- Default: `-1`

`camera.SetLightFalloff(falloff float64)`
- Sets value that simulates "torch" light, lower values make torch dimmer.
- Default: `-100`

`camera.SetGlobalIllumination(illumination float64)`
- Sets illumination value for whole level ("sun" brightness).
- Default: `300`

`camera.SetLightRGB(min, max color.NRGBA)`
- Sets the min/max color tinting of the textures when fully shadowed (min) or lighted (max).
- Default: min=NRGBA{0, 0, 0}, max=NRGBA{255, 255, 255}

## Limitations

- Raycasting is not raytracing.
- Raycasting draws 2D textures and sprites using a semi-3D technique, not using 3D models.
- The raycasting technique used in this project is more like early raycaster games such as Wolfenstein 3D,
  as opposed to later games such as Doom - it does not support stairs, sloped walls,
  or differing heights in elevation levels.
- Multiple elevation levels can be rendered, however camera and sprite positions need to be limited
  to the ground level (Z-position `> 0.0 && <= 1.0`).
- Only a single repeating floor texture can currently be set for the entire map.
- [Ceiling textures](https://lodev.org/cgtutor/raycasting2.html) are not currently implemented.
  Skybox texture is currently the only option, so going indoors from outdoors in the same map is not currently possible.
  Feel free to help figure it out and contribute as a Pull Request!
- [Thin walls](https://lodev.org/cgtutor/raycasting4.html#Thin), [doors]((https://lodev.org/cgtutor/raycasting4.html#Doors)),
  and [secret push walls](https://lodev.org/cgtutor/raycasting4.html#Secrets) are not currently implemented,
  feel free to help figure them out and contribute as a Pull Request!
- [Translucent sprites](https://lodev.org/cgtutor/raycasting3.html#Translucent) are not currently implemented,
  feel free to contribute as a Pull Request!
