package main

/*
#cgo CFLAGS: -I/opt/miyoomini-toolchain/arm-linux-gnueabihf/libc/usr/include/SDL -O2 -w -D_GNU_SOURCE=1 -D_REENTRANT
#cgo LDFLAGS: -L/opt/miyoomini-toolchain/arm-linux-gnueabihf/libc/usr/lib -lSDL -lpthread
#include "main.c"
*/
import "C"

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"unsafe"
)

func main() {
	C.init()
	println("Init SDL OK")

	file, err := os.Open("./assets/golang.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	pngImage, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	fillColor := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 40; y++ {
		for x := 0; x < 300; x++ {
			img.SetRGBA(x, y, fillColor)
		}
	}

	for y := 40; y < 80; y++ {
		for x := 0; x < 300; x++ {
			img.SetRGBA(x, y, color.RGBA{0, 255, 0, 255})
		}
	}

	for y := 80; y < 120; y++ {
		for x := 0; x < 300; x++ {
			img.SetRGBA(x, y, color.RGBA{0, 0, 255, 255})
		}
	}

	draw.Draw(img, image.Rect(40, 40, 40+pngImage.Bounds().Dx(), 40+pngImage.Bounds().Dy()), pngImage, image.Point{0, 0}, draw.Over)

	println("osdksdd")
	finished := false
	for !finished {
		println("Looping")
		value := C.pollEvents()
		switch value {
		case 0:
			finished = true
		case 1:
			for y := 0; y < 480; y++ {
				for x := 0; x < 640; x++ {
					img.SetRGBA(x, y, color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 255})
				}
			}

			fillColor := color.RGBA{255, 0, 0, 255}
			for y := 0; y < 40; y++ {
				for x := 0; x < 300; x++ {
					img.SetRGBA(x, y, fillColor)
				}
			}

			for y := 40; y < 80; y++ {
				for x := 0; x < 300; x++ {
					img.SetRGBA(x, y, color.RGBA{0, 255, 0, 255})
				}
			}

			for y := 80; y < 120; y++ {
				for x := 0; x < 300; x++ {
					img.SetRGBA(x, y, color.RGBA{0, 0, 255, 255})
				}
			}
		}
		println("Value: ", value)

		incy := rand.Intn(256)
		incx := rand.Intn(420)

		draw.Draw(img, image.Rect(incx, incy, incx+pngImage.Bounds().Dx(), incy+pngImage.Bounds().Dy()), pngImage, image.Point{0, 0}, draw.Over)

		C.refreshScreenPtr((*C.uchar)(unsafe.Pointer(&img.Pix[0])))
	}

	C.quit()
}
