package main

import (
	"flag"
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"time"
	"unsafe"
	"math/cmplx"
)

type rgb struct {
	r, g, b uint8
}

const (
	WIDTH      = 1280
	HEIGHT     = 1080
	VIEWWIDTH  = 1280
	VIEWHEIGHT = 1024
	THREADS    = 4
	STARTSIZE  = 64
)

var (
	exit, imExit                       bool
	pixel                              [WIDTH][HEIGHT]float64
	processed                          [WIDTH][HEIGHT]bool
	completed                          chan int
	threadFinished                     [THREADS]bool
	xCentre, yCentre, xOffset, yOffset int
	phaser, mono, doJulia, doTricorn   bool
	fracX, fracY, scale, infinity      float64
	iterations, segment, batch, pow    int
	output                             string
	juliaR, juliaI                     float64
)

func init() {
	runtime.LockOSThread()
}

func mandelbrot(x, y int) float64 {

	c := complex(float64(x+xCentre)*scale+fracX, float64(y+yCentre)*scale+fracY)

	z := complex(0, 0)

	var i int

	switch pow {
	case 2:
		for i = 0; i < iterations; i++ {
			z = z*z + c
			if imag(z) > infinity || real(z) > infinity {
				break
			}
		}
	default:
		for i = 0; i < iterations; i++ {
			zz := z
			for iz := 0; iz < pow; iz++ {
				z *= zz
			}
			z += c
			if imag(z) > infinity || real(z) > infinity {
				break
			}
		}

	}

	return float64(i) / float64(iterations)

}


func tricorn(x, y int) float64 {

	c := complex(float64(x+xCentre)*scale+fracX, float64(y+yCentre)*scale+fracY)

	z := complex(0, 0)

	var i int

	switch pow {
	case 2:
		for i = 0; i < iterations; i++ {
			z = cmplx.Conj(z*z) + c
			if imag(z) > infinity || real(z) > infinity {
				break
			}
		}
	default:
		for i = 0; i < iterations; i++ {
			zz := z
			for iz := 0; iz < pow; iz++ {
				z *= zz
			}
			z = cmplx.Conj(z)
			z += c
			if imag(z) > infinity || real(z) > infinity {
				break
			}
		}

	}

	return float64(i) / float64(iterations)

}


func julia(x, y int, jr, ji float64) float64 {

	z := complex(float64(x+xCentre)*scale+fracX, float64(y+yCentre)*scale+fracY)
	c := complex(jr, ji)

	var i int
	for i = 0; i < iterations; i++ {

		switch pow {
		case 2:
			z = z*z + c
		case 3:
			z = z*z*z + c
		case 4:
			z = z*z*z*z + c
		default:
			zz := z
			for iz := 0; iz < pow; iz++ {
				z *= zz
			}
			z += c
		}

		if imag(z) > infinity || real(z) > infinity {
			break
		}
	}

	return float64(i) / float64(iterations)

}

func render(core int) {

	fmt.Println("Thread", core, "started...")

	pixelSize := STARTSIZE
	renderX := 0
	renderY := core * pixelSize

	pixelCount := 0

renderLoop:
	for pixelCount < batch {

		if renderY < HEIGHT {

			if int(renderY/STARTSIZE)%THREADS == core {

				if !processed[renderX][renderY] {

					var m float64

					if doTricorn {
						m = tricorn(renderX, renderY)
					} else if !doJulia {
						m = mandelbrot(renderX, renderY)
					} else {
						m = julia(renderX, renderY, juliaR, juliaI)
					}

					pixelCount++

					for u := 0; u < pixelSize; u++ {
						for v := 0; v < pixelSize; v++ {
							if renderX+u < WIDTH && renderY+v < HEIGHT {
								pixel[renderX+u][renderY+v] = m
							}
						}
					}
					processed[renderX][renderY] = true
				}

			}

			renderX += pixelSize

			if renderX >= WIDTH {
				renderX = 0
				renderY += pixelSize
			}

		} else {

			if pixelSize == 1 {
				fmt.Println("Thread", core, "is finished!!!")
				threadFinished[core] = true
				break renderLoop
			} else {
				renderX = 0
				renderY = core * pixelSize
				pixelSize = pixelSize / 2
				continue renderLoop
			}

		}

	}

	fmt.Println("Sending complete signal for thread", core)
	completed <- core

}

func valueToColor(value float64, phased bool, phase float64) rgb {

	if !phased {

		if mono {

			return rgb{uint8((1 - value) * 255), uint8((1 - value) * 255), uint8((1 - value) * 255)}

		} else {

			value *= 6

			if value < 1 {
				return rgb{uint8(255 * value), 0, 255}
			} else if value < 2 {
				value -= 1
				return rgb{1, uint8(255 * value), uint8(255 * (1.0 - value))}
			} else if value < 3 {
				value -= 2
				return rgb{uint8(255 * (1.0 - value)), 255, 0}
			} else if value < 4 {
				value -= 3
				return rgb{0, 255, uint8(255 * value)}
			} else if value < 5 {
				value -= 4
				return rgb{0, uint8(255 * (1.0 - value)), 255}
			} else {
				value -= 5
				return rgb{0, 0, uint8(255 * (1.0 - value))}
			}

		}

	} else {

		value += phase
		if value >= 1 {
			value -= 1
		}

		if mono {

			value = math.Abs(value*2 - 1)
			return rgb{uint8(value * 255), uint8(value * 255), uint8(value * 255)}

		} else {

			value *= 6

			if value < 1 {
				return rgb{255, 0, uint8(value * 255)}
			} else if value < 2 {
				value -= 1
				return rgb{uint8((1 - value) * 255), 0, 255}
			} else if value < 3 {
				value -= 2
				return rgb{0, uint8(value * 255), 255}
			} else if value < 4 {
				value -= 3
				return rgb{0, 255, uint8((1 - value) * 255)}
			} else if value < 5 {
				value -= 4
				return rgb{uint8(value * 255), 255, 0}
			} else {
				value -= 5
				return rgb{255, uint8((1 - value) * 255), 0}
			}

		}
	}

}

func drawScene(phased bool, phase float64) {

	var data [VIEWHEIGHT][VIEWWIDTH][3]uint8

	for x := 0; x < VIEWWIDTH; x++ {
		for y := 0; y < VIEWHEIGHT; y++ {

			c := valueToColor(pixel[x+xOffset][y+yOffset], phased, phase)

			data[VIEWHEIGHT-y-1][x][0], data[VIEWHEIGHT-y-1][x][1], data[VIEWHEIGHT-y-1][x][2] = c.r, c.g, c.b
		}
	}

	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.DrawPixels(VIEWWIDTH, VIEWHEIGHT, gl.RGB, gl.UNSIGNED_BYTE, unsafe.Pointer(&data))

}

func saveImage() {

	img := image.NewNRGBA(image.Rect(0, 0, WIDTH, HEIGHT))

	for x := 0; x < WIDTH; x++ {
		for y := 0; y < HEIGHT; y++ {
			c := valueToColor(pixel[x][y], false, 0)
			img.Set(x, y, color.RGBA{R: c.r, G: c.g, B: c.b, A: 255})
		}
	}

	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("PNG file saved successfully")

}

func main() {

	fracXPtr := flag.Float64("x", -0.5, "X co-ordinate (floating point)")
	fracYPtr := flag.Float64("y", 0, "Y co-ordinate (floating point)")
	scalePtr := flag.Float64("z", 500, "Zoom (floating point)")
	iterationsPtr := flag.Int("i", 200, "Iterations (integer)")
	segmentPtr := flag.Int("seg", 0, "Segment (0 for none or 1-6 for hex segment")
	phaserPtr := flag.Bool("phase", false, "Phase image after rendering")
	infinityPtr := flag.Float64("inf", 1e+100, "Value for infinity")
	monoPtr := flag.Bool("mono", false, "Use monochrome mode")
	outputPtr := flag.String("output", "lastrender.png", "Filename to save image in .png format")
	imExitPtr := flag.Bool("exit", false, "Immediately exit after rendering")
	batchPtr := flag.Int("batch", 20480, "Batch size between screen updates (default is 16 lines)")
	powPtr := flag.Int("pow", 2, "Power for complex number iteration (2 for standard mandelbrot)")
	tricornPtr := flag.Bool("tricorn", false, "Use Tricorn mode")
	juliaPtr := flag.Bool("julia", false, "Use Julia mode")
	juliaRPtr := flag.Float64("jr", 0, "Julia real part")
	juliaIPtr := flag.Float64("ji", 0, "Julia imaginary part")

	flag.Parse()

	fracX = *fracXPtr
	fracY = *fracYPtr

	zoom := *scalePtr
	if zoom == 0 {
		zoom = 1
	}
	scale = 1.0 / zoom

	iterations = *iterationsPtr

	segment = *segmentPtr
	if segment < 0 || segment > 6 {
		segment = 0
	}

	phaser = *phaserPtr
	infinity = *infinityPtr
	mono = *monoPtr
	output = *outputPtr
	imExit = *imExitPtr
	batch = *batchPtr
	if batch < 1 {
		batch = 1
	}
	if batch > WIDTH*HEIGHT {
		batch = WIDTH * HEIGHT
	}

	pow = *powPtr
	if pow < 2 { pow = 2 }

	doJulia = *juliaPtr
	if doJulia {
		juliaR = *juliaRPtr
		juliaI = *juliaIPtr
	}

	doTricorn = *tricornPtr


	switch segment {
	case 0:
		xCentre = -WIDTH / 2
		yCentre = -HEIGHT / 2
		xOffset = 0
		yOffset = (HEIGHT - VIEWHEIGHT) / 2
	case 1:
		xCentre = -3 * WIDTH / 2
		yCentre = -HEIGHT
		xOffset = 0
		yOffset = HEIGHT - VIEWHEIGHT
	case 2:
		xCentre = -WIDTH / 2
		yCentre = -HEIGHT
		xOffset = 0
		yOffset = HEIGHT - VIEWHEIGHT
	case 3:
		xCentre = WIDTH / 2
		yCentre = -HEIGHT
		xOffset = 0
		yOffset = HEIGHT - VIEWHEIGHT
	case 4:
		xCentre = -3 * WIDTH / 2
		yCentre = 0
		xOffset = 0
		yOffset = 0
	case 5:
		xCentre = -WIDTH / 2
		yCentre = 0
		xOffset = 0
		yOffset = 0
	case 6:
		xCentre = WIDTH / 2
		yCentre = 0
		xOffset = 0
		yOffset = 0
	}

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	window, err := glfw.CreateWindow(VIEWWIDTH, VIEWHEIGHT, "Pixels", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	window.SetPos(0, 0)

	window.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		switch {
		case key == glfw.KeyEscape && action == glfw.Press:
			exit = true
		}
	})

	if err := gl.Init(); err != nil {
		panic(err)
	}

	drawScene(false, 0)
	window.SwapBuffers()

	var done [THREADS]bool
	renderStart := time.Now()
	calculationStart := time.Now()
	allFinished := false
	completed = make(chan int, 1)

	for t := 0; t < THREADS; t++ {
		go render(t)
	}

	for !window.ShouldClose() && !exit {

		select {
		case c := <-completed:

			done[c] = true
			fmt.Println("Thread", c, "done.")
			doneCount := 0

			for t := 0; t < THREADS; t++ {
				if done[t] {
					doneCount++
				}
			}

			if doneCount == THREADS {

				calculationEnd := time.Since(calculationStart).Seconds()

				fmt.Println("All threads done in", calculationEnd, "seconds.")

				sceneStart := time.Now()
				drawScene(false, 0)
				window.SwapBuffers()
				sceneEnd := time.Since(sceneStart).Seconds()

				fmt.Println("Scene draw time", sceneEnd, "seconds.")

				allFinished = true

				calculationStart = time.Now()

				for t := 0; t < THREADS; t++ {
					if !threadFinished[t] {
						done[t] = false
						allFinished = false
						fmt.Println("Starting thread", t)
						go render(t)
					}
				}

				if allFinished {
					renderEnd := time.Since(renderStart).Seconds()

					if output != "" {
						saveImage()
						if imExit {
							exit = true
						}
					}

					fmt.Println("COMPLETED in", renderEnd, "seconds.")
				}

			}

		default:
		}

		if allFinished && phaser {
			_, f := math.Modf(time.Since(calculationStart).Seconds() / 10)
			drawScene(true, f)
			window.SwapBuffers()
		}
		glfw.PollEvents()

	}

	window.SetShouldClose(true)

}
