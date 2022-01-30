# raycaster-go
Golang conversion of raycaster engine [OwlRaycastEngine](https://github.com/Owlzy/OwlRaycastEngine)
using the [Ebiten 2D Game Library](https://github.com/hajimehoshi/ebiten).
To see it in action visit the [YouTube playlist](https://www.youtube.com/playlist?list=PLOINtzQqJWIh8OQsvYAahr2yuAF5VLk38).

![Screenshot](screenshot.jpg?raw=true)

## Developer setup
To install and run from source the following are required:
1. Download, install, and setup Golang https://golang.org/dl/
2. Clone/download this project locally.
3. From the project folder use the following command to download the Go module dependencies of this project:
    * `[path/to/raycaster-go]$ go mod download`
4. Now you can use the `go run` command to run `raycast.go`:
    * `[path/to/raycaster-go]$ go run raycast.go`

## Controls
* Move and rotate using WASD or Arrow Keys
* Strafe by holding Alt with the rotate keys
* Left/right mouse click currently used for visual/console debugging
