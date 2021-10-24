package main

import (
	"bufio"
	"fmt"
	"github.com/konimarti/kalman"
	"github.com/konimarti/lti"
	"gonum.org/v1/gonum/mat"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

// #cgo LDFLAGS: -framework ApplicationServices
// #include <mouse.c>
import "C"

var ctx kalman.Context
var filter kalman.Filter
var control *mat.VecDense
var control_scalar float64 = 2
var thumb mgl32.Vec3
var finger mgl32.Vec3
var wrist mgl32.Vec3
var index_finger_mcp mgl32.Vec3
var clicked bool
var xmax, ymax float32

//Oh god, golang, and everyone who programs it, is so dumb it hurts
func pf(s string) float32 {
	v, _ := strconv.ParseFloat(s, 32)
	return float32(v)
}

func calcMousePos() (float32, float32) {

	x := wrist[0] * xmax
	y := wrist[1] * ymax

	xout, yout := do_kalman(float64(x), float64(y))

	//x := xmax * (thumb[0] + (thumb[0]-finger[0])/2)
	//y := ymax * (thumb[1] + (thumb[1]-finger[1])/2)

	//log.Printf("Moving mouse to: %v,%v\n", x, y)
	return float32(xout), float32(yout)

}

func handleStr(s []string) {
	landmark, _ := strconv.Atoi(s[1])
	handleVec(landmark, mgl32.Vec3{pf(s[2]), pf(s[3]), pf(s[4])})
}

var smoothing = float32(1.0)

func handleVec(landmark int, v mgl32.Vec3) {
	scale := float32(1.2)
	ss := float32(0.5 / 1.4)
	v = v.Sub(mgl32.Vec3{0.5, 0.5, 0.5})
	v = v.Mul(scale)
	v = v.Add(mgl32.Vec3{ss, ss, ss})
	if landmark == 4 {
		//log.Println("Got line:", s)
		//thumb = mgl32.Vec3{pf(s[2]), pf(s[3]), pf(s[4])}
		thumb = v.Mul(smoothing).Add(thumb.Mul(1 - smoothing))
		//log.Println("Thumb", thumb)
	}
	if landmark == 5 {
		//log.Println("Got line:", s)
		//thumb = mgl32.Vec3{pf(s[2]), pf(s[3]), pf(s[4])}
		index_finger_mcp = v.Mul(smoothing).Add(thumb.Mul(1 - smoothing))
	}
	if landmark == 8 {
		//log.Println("Got line:", s)
		finger = v.Mul(smoothing).Add(finger.Mul(1 - smoothing))
		//log.Println("Finger", finger)

		//x := finger[0] * xmax
		//y := finger[1] * ymax
	}
	if landmark == 0 {
		//log.Println("Got line:", s)
		wrist = v.Mul(smoothing).Add(wrist.Mul(1 - smoothing))
		//log.Println("Finger", finger)

		//x := finger[0] * xmax
		//y := finger[1] * ymax
	}

	//x := xmax * (thumb[0] + (thumb[0]-finger[0])/2)
	//y := ymax * (thumb[1] + (thumb[1]-finger[1])/2)

}

func main() {
	xmax = 2000
	ymax = 1000
	var line string
	setup_kalman()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line = scanner.Text()
		bits := strings.SplitN(line, ",", -1)
		if len(bits) > 2 {
			handleStr(bits)
		}

		v1 := mgl32.Vec2{thumb[0], thumb[1]}
		v2 := mgl32.Vec2{finger[0], finger[1]}
		//z := wrist[2]
		//fmt.Printf("Z: %v\n", z)
		distance := v1.Sub(v2).Len()
		//log.Println("Distance:", distance)
		//x := finger[0] * xmax
		//y := finger[1] * ymax
		x, y := calcMousePos()
		C.MoveMouse(C.int(x), C.int(y))
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

func setup_kalman() {
	// prepare output file
	file, err := os.Create("car.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintln(file, "Measured_v_x,Measured_v_y,Filtered_v_x,Filtered_v_y")

	ctx = kalman.Context{
		// init state: pos_x = 0, pox_y = 0, v_x = 30 km/h, v_y = 10 km/h
		X: mat.NewVecDense(4, []float64{0, 0, 30, 10}),
		// initial covariance matrix
		P: mat.NewDense(4, 4, []float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1}),
	}

	// time step
	dt := 0.1

	lti := lti.Discrete{
		// prediction matrix
		Ad: mat.NewDense(4, 4, []float64{
			1, 0, dt, 0,
			0, 1, 0, dt,
			0, 0, 1, 0,
			0, 0, 0, 1,
		}),
		// no external influence
		Bd: mat.NewDense(4, 4, nil),
		// scaling matrix for measurement
		C: mat.NewDense(2, 4, []float64{
			0, 0, 1, 0,
			0, 0, 0, 1,
		}),
		// scaling matrix for control
		D: mat.NewDense(2, 4, nil),
	}

	// G
	G := mat.NewDense(4, 2, []float64{
		0, 0,
		0, 0,
		1, 0,
		0, 1,
	})
	var Gd mat.Dense
	Gd.Mul(lti.Ad, G)

	// process model covariance matrix
	qk := mat.NewDense(2, 2, []float64{
		0.01, 0,
		0, 0.01,
	})
	var Q mat.Dense
	Q.Product(&Gd, qk, Gd.T())

	// measurement errors
	corr := 0.5
	R := mat.NewDense(2, 2, []float64{1, corr, corr, 1})

	// create noise struct
	nse := kalman.Noise{&Q, R}

	// create Kalman filter
	filter = kalman.NewFilter(lti, nse)

	// no control
	control = mat.NewVecDense(4, []float64{control_scalar, control_scalar, control_scalar, control_scalar})
	//control = mat.NewVecDense(4, nil)

}

func do_kalman(x, y float64) (float64, float64) {

	// measure v_x and v_y with an error which is distributed according to stanard normal
	measurement := mat.NewVecDense(2, []float64{x, y})

	// apply filter
	filtered := filter.Apply(&ctx, measurement, control)

	// print out
	fmt.Sprintf("%3.8f,%3.8f,%3.8f,%3.8f\n", measurement.AtVec(0), measurement.AtVec(1),
		filtered.AtVec(0), filtered.AtVec(1))
	return filtered.AtVec(0), filtered.AtVec(1)

}
