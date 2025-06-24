package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"time"

	SH1107 "github.com/mikedev101/sh1107-i2c-go/sh1107"
)

func drawSmiley(img *image.Gray) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	cx, cy := w/2, h/2

	// Face (outer circle)
	drawCircle(img, cx, cy, 40, color.Gray{Y: 255})

	// Eyes
	drawCircle(img, cx-15, cy-10, 5, color.Gray{Y: 0})
	drawCircle(img, cx+15, cy-10, 5, color.Gray{Y: 0})

	// Smile (arc)
	for x := -20; x <= 20; x++ {

		// Calculate y for a half ellipse: y = b * sqrt(1 - (x^2 / a^2))
		a, b := 20.0, 10.0
		y := int(b * math.Sqrt(1-(float64(x*x)/(a*a))))
		for dy := -1; dy <= 1; dy++ {
			px, py := cx+x, cy+8+y+dy
			if px >= 0 && px < w && py >= 0 && py < h {
				img.SetGray(px, py, color.Gray{Y: 0})
			}
		}
	}
}

func drawCircle(img *image.Gray, cx, cy, r int, col color.Gray) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				px, py := cx+x, cy+y
				if px >= 0 && px < w && py >= 0 && py < h {
					img.SetGray(px, py, col)
				}
			}
		}
	}
}

func main() {

	// Create a framebuffer
	img := image.NewGray(image.Rect(0, 0, 128, 128))

	// Initialize display
	log.Print("Starting display...")
	display := SH1107.New(0x3C, 0, SH1107.Normal, 128, 128)
	defer display.Close()

	log.Print("Turning ON display...")
	display.On()

	log.Print("Clearing display...")
	display.Clear(SH1107.White)
	time.Sleep(time.Second)

	log.Print("Adjusting brightness...")
	for i := 0; i < 100; i++ {
		display.SetBrightness(float64(i) / 100.0)
		time.Sleep(5 * time.Millisecond)
	}

	log.Print("Test pattern")
	display.TestPattern()
	time.Sleep(time.Second)

	log.Print("Drawing a smiley :)")
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Gray{0}}, image.Point{}, draw.Src)
	drawSmiley(img)
	display.Blit(img)
	time.Sleep(time.Second)

	log.Print("Bad apple demo!")
	display.PlayAnimation("frame128", 6574, 30)
}
