package main

import (
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"os"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"time"
)

var (
	texture   uint32
	rotationX float32
	rotationY float32
	frameLength float64
	frames            = 0
	second            = time.Tick(time.Second)
	exit	bool
)


func main() {

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	window, err := glfw.CreateWindow(1920, 1080, "Cube", glfw.GetPrimaryMonitor(), nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	window.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		switch {
		case key == glfw.KeyEscape && action == glfw.Press:
			exit = true
		}
	})


	if err := gl.Init(); err != nil {
		panic(err)
	}

	texture = newTexture("square.png")
	defer gl.DeleteTextures(1, &texture)

	setupScene()
	for !window.ShouldClose() && !exit {

		frameStart := time.Now()

		drawScene()
		window.SwapBuffers()
		glfw.PollEvents()

		frames++
		select {
		case <-second:
			frames = 0
		default:
		}

		frameLength = time.Since(frameStart).Seconds()
	}

	window.SetShouldClose(true)

}

func newTexture(file string) uint32 {

	imgFile, err := os.Open(file)
	if err != nil {
		log.Fatalf("texture %q not found on disk: %v\n", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.Enable(gl.TEXTURE_2D)
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture

}

func setupScene() {

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.LIGHTING)

	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	ambient := []float32{0.5, 0.5, 0.5, 1}
	diffuse := []float32{1, 1, 1, 1}
	lightPosition := []float32{-5, 5, 10, 0}
	gl.Lightfv(gl.LIGHT0, gl.AMBIENT, &ambient[0])
	gl.Lightfv(gl.LIGHT0, gl.DIFFUSE, &diffuse[0])
	gl.Lightfv(gl.LIGHT0, gl.POSITION, &lightPosition[0])
	gl.Enable(gl.LIGHT0)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Frustum(-1.92, 1.92, -1.08, 1.08, 1.0, 100.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

}


func drawScene() {

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Translatef(0, 0, -3.0)
	gl.Rotatef(rotationX, 1, 0, 0)
	gl.Rotatef(rotationY, 0, 1, 0)

	rotationX += float32(90*frameLength)
	rotationY += float32(90*frameLength)

	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.Color4f(1, 1, 1, 1)

	gl.Begin(gl.QUADS)

	gl.Normal3f(0, 0, 1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-1, -1, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(1, -1, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(1, 1, 1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-1, 1, 1)

	gl.Normal3f(0, 0, -1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(-1, -1, -1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(-1, 1, -1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(1, 1, -1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(1, -1, -1)

	gl.Normal3f(0, 1, 0)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-1, 1, -1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-1, 1, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(1, 1, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(1, 1, -1)

	gl.Normal3f(0, -1, 0)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(-1, -1, -1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(1, -1, -1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(1, -1, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(-1, -1, 1)

	gl.Normal3f(1, 0, 0)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(1, -1, -1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(1, 1, -1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(1, 1, 1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(1, -1, 1)

	gl.Normal3f(-1, 0, 0)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-1, -1, -1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(-1, -1, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(-1, 1, 1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-1, 1, -1)

	gl.End()
}
