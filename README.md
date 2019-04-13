# raycaster-go
Golang conversion of raycaster engine [OwlRaycastEngine](https://github.com/Owlzy/OwlRaycastEngine) using the [Ebiten 2D Game Library](https://github.com/hajimehoshi/ebiten). To see it in action visit the [YouTube playlist](https://www.youtube.com/playlist?list=PLOINtzQqJWIh8OQsvYAahr2yuAF5VLk38).

## Developer setup
To install and run from source the following are required:
1. Download, install, and setup Golang https://golang.org/dl/
2. Use the `go get` command to download the Ebiten 2D Game Library: 
    * `[GOPATH]$ go get github.com/hajimehoshi/ebiten/...`
3. Clone/download this project to your `GOPATH/src` folder, for example: `GOPATH/src/raycaster-go`
4. Now you can use the `go run` command to run `raycast.go`:
    * `[GOPATH/src/raycaster-go]$ go run raycast.go`

## Controls
* Move and turn using WASD and Arrow Keys
* Left/right mouse click currently used for visual/console debugging
