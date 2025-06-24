package sh1107

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
	"github.com/sergeymakinen/go-bmp"
)

type SH1107 struct {
	bus           *i2c.I2C
	fb            []byte
	rot           uint8
	Width, Height int
}

const Black byte = 0x00
const White byte = 0xFF

const (
	Normal            uint8 = 0
	Flipped           uint8 = 1
	UpsideDown        uint8 = 2
	FlippedUpsideDown uint8 = 3
)

// Creates a new SH1107 display connection
func New(address byte, bus_device int, rotation uint8, width, height int) *SH1107 {
	logger.ChangePackageLogLevel("i2c", logger.PanicLevel)
	bus, err := i2c.NewI2C(address, bus_device)
	if err != nil {
		panic(err)
	}
	display := &SH1107{
		bus,
		make([]byte, 128*16),
		rotation,
		width,
		height,
	}
	display.init()
	display.SetRotation(rotation)
	display.Clear(Black)
	return display
}

// TODO: THIS CURRENTLY DOESN'T PROPERLY WORK FOR 90/270 DEGREES
func (d *SH1107) SetRotation(rot uint8) {
	switch rot % 4 {
	case 0: // 0°
		d.writeCommand(0xA0) // Segment remap normal
		d.writeCommand(0xC0) // COM scan flipped
	case 1: // 90°
		d.writeCommand(0xA1) // Segment remap
		d.writeCommand(0xC0) // COM scan flipped
	case 2: // 180°
		d.writeCommand(0xA1) // Segment remap
		d.writeCommand(0xC8) // COM scan normal
	case 3: // 270°
		d.writeCommand(0xA0) // Segment remap normal
		d.writeCommand(0xC8) // COM scan normal
	}
}

func (d *SH1107) multiCommand(cmd ...byte) {
	buf := append([]byte{0x00}, cmd...)
	d.bus.WriteBytes(buf)
}

func (d *SH1107) writeCommand(cmd ...any) {
	d.write(0x00, cmd...)
}

func (d *SH1107) writeData(data ...any) {
	d.write(0x40, data...)
}

func (d *SH1107) write(cmd byte, data ...any) {
	for _, v := range data {
		switch val := v.(type) {
		case byte:
			d.bus.WriteBytes([]byte{cmd, val})
		case int:
			d.bus.WriteBytes([]byte{cmd, byte(val)})
		case []byte:
			d.bus.WriteBytes(append([]byte{cmd}, val...))
		default:
			panic(fmt.Sprintf("Unsupported type %T", v))
		}
	}
}

func (d *SH1107) init() {
	cmds := []byte{
		0xAE,       // display off
		0x00, 0x10, // set column addr low + high
		0xDC, 0x00, // display start line
		0x81, 0x7F, // contrast
		0x20,       // page addressing
		0xA4,       // disable entire display on
		0xA6,       // normal display
		0xA8, 0x7F, // multiplex ratio = 127
		0xD3, 0x00, // display offset
		0xD5, 0x41, // osc
		0xD9, 0x22, // precharge
		0xDB, 0x35, // vcomh
		0xAD, 0x8A, // charge pump enable
		0xAF, // display on
	}
	for _, cmd := range cmds {
		d.writeCommand(cmd)
	}
}

// Writes a raw byte to the framebuffer
func (d *SH1107) Set(x, y int, color byte) {
	if x < 0 || x >= 128 || y < 0 || y >= 128 {
		return
	}
	page := y / 8
	offset := page*128 + x
	bit := uint(y % 8)

	if color != 0 {
		d.fb[offset] |= (1 << bit)
	} else {
		d.fb[offset] &^= (1 << bit)
	}
}

// Closes the display connection
func (d *SH1107) Close() {
	d.bus.Close()
}

// Turns display on
func (d *SH1107) On() {
	d.writeCommand(0xAF)
}

// Turns display off
func (d *SH1107) Off() {
	d.writeCommand(0xAE)
}

// Set brightness from 0.0 to 1.0
func (d *SH1107) SetBrightness(level float64) {
	scaled := byte(uint8(float64(0x7F) * level))
	d.writeCommand(0x81, scaled, level) // 0x7F is the max brightness for SH1107
}

// Clears the display with either all black or all white
func (d *SH1107) Clear(state byte) {
	for i := range d.fb {
		d.fb[i] = state
	}
	d.Render()
}

// Shows a checkerboard pattern on the display
func (d *SH1107) TestPattern() {
	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			if (x+y)%2 == 0 {
				d.Set(x, y, White)
			} else {
				d.Set(x, y, Black)
			}
		}
	}
	d.Render()
}

// Displays whatever is stored in the framebuffer
func (d *SH1107) Render() {
	width := d.Width
	pages := d.Height / 8 // 128px height / 8 pixels per page
	for page := 0; page < pages; page++ {

		// Combine multiple commands into a single transaction
		d.multiCommand(
			0xB0|byte(page), // page address
			0x00,            // low nibble
			0x10,            // high nibble
		)

		// Transmit data as a single transaction
		offset := page * width
		end := offset + width
		d.writeData(d.fb[offset:end])
	}
}

// Display single image to screen
func (d *SH1107) Blit(img image.Image) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width != d.Width || height != d.Height {
		log.Printf("Draw() requires %dx%d image", d.Width, d.Height)
		return
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			pixel := Black
			if gray.Y > 127 { // Maybe adjust this threshold?
				pixel = White
			}
			d.Set(x, y, pixel)
		}
	}
	d.Render()
}

// Plays a sequence of images
func (d *SH1107) PlayAnimation(dir string, frameCount int, fps int) {
	const bufferSize = 30
	frameDuration := time.Second / time.Duration(fps)
	buffer := make([]image.Image, 0, bufferSize)
	startTime := time.Now()
	for i := 1; i <= frameCount; i++ {
		filename := fmt.Sprintf("%s/%d.bmp", dir, i)
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("Failed to open %s: %v\n", filename, err)
			continue
		}
		img, err := bmp.Decode(file)
		file.Close()
		if err != nil {
			fmt.Printf("Failed to decode %s: %v\n", filename, err)
			continue
		}

		buffer = append(buffer, img)

		if len(buffer) == bufferSize || i == frameCount {
			for j, frame := range buffer {
				targetTime := startTime.Add(time.Duration(i-bufferSize+j) * frameDuration)

				d.Blit(frame)

				// This might not be perfectly consistent
				sleepUntil := time.Until(targetTime)
				if sleepUntil > 0 {
					time.Sleep(sleepUntil)
				}
			}
			buffer = buffer[:0]
		}
	}
}
