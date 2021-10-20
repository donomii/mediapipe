package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

// #cgo LDFLAGS: -framework ApplicationServices
// #include <mouse.c>
import "C"

var thumb mgl32.Vec3
var finger mgl32.Vec3
var clicked bool
var xmax, ymax float32

//Oh god, golang, and everyone who programs it, is so dumb it hurts
func pf(s string) float32 {
	v, _ := strconv.ParseFloat(s, 32)
	return float32(v)
}

func handle(s []string) {
	if s[1] == "4" {
		//log.Println("Got line:", s)
		//thumb = mgl32.Vec3{pf(s[2]), pf(s[3]), pf(s[4])}
		newThumb := mgl32.Vec3{pf(s[2]), pf(s[3]), pf(s[4])}
		thumb = newThumb.Mul(0.6).Add(thumb.Mul(0.4))
		//log.Println("Thumb", thumb)
		x := thumb[0] * xmax
		y := thumb[1] * ymax
		//log.Printf("Moving mouse to: %v,%v\n", x, y)
		C.MoveMouse(C.int(x), C.int(y))
	}
	if s[1] == "8" {
		//log.Println("Got line:", s)
		newfinger := mgl32.Vec3{pf(s[2]), pf(s[3]), pf(s[4])}
		finger = newfinger.Mul(0.6).Add(finger.Mul(0.4))
		//log.Println("Finger", finger)

		//x := finger[0] * xmax
		//y := finger[1] * ymax
	}
	x := xmax * (thumb[0] + (thumb[0]-finger[0])/2)
	y := ymax * (thumb[1] + (thumb[1]-finger[1])/2)
	//log.Printf("Moving mouse to: %v,%v\n", x, y)
	C.MoveMouse(C.int(x), C.int(y))

}

func main() {
	xmax = 2000
	ymax = 1000
	var line string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line = scanner.Text()
		bits := strings.SplitN(line, ",", -1)
		if len(bits) > 2 {
			handle(bits)
		}

		v1 := mgl32.Vec2{thumb[0], thumb[1]}
		v2 := mgl32.Vec2{finger[0], finger[1]}
		distance := v1.Sub(v2).Len()
		//log.Println("Distance:", distance)
		//x := finger[0] * xmax
		//y := finger[1] * ymax
		x := xmax * (thumb[0] + (thumb[0]-finger[0])/2)
		y := ymax * (thumb[1] + (thumb[1]-finger[1])/2)
		if distance < 0.04 {
			if !clicked {
				log.Println("Click down")
				C.ClickMouseDown(C.int(x), C.int(y))
				C.ClickMouseUp(C.int(x), C.int(y))
				clicked = true
			}
		} else if distance > 0.07 {
			if clicked {
				C.ClickMouseUp(C.int(x), C.int(y))
				clicked = false
				log.Println("Click up")
			}

		}
		/*for _, x := range strings.SplitN(line, ",", -1) {
		  fmt.Println("next field: ", x)
		}*/
	}
}
