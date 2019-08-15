package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
)

const (
	width, height = 600, 320            // canvas size in pixels
	cells         = 100                 // number of grid cells
	xyrange       = 30.0                // axis ranges (-xyrange..+xyrange)
	xyscale       = width / 2 / xyrange // pixels per x or y unit
	zscale        = height * 0.4        // pixels per z unit
	angle         = math.Pi / 6         // angle of x, y axes (=30°)
)

var sin30, cos30 = math.Sin(angle), math.Cos(angle) // sin(30°), cos(30°)
var funcName string

func main() {
	defaultFunc := func(w http.ResponseWriter, r *http.Request) {
		funcName = "f"
		svg(w)
	}
	eggbox := func(w http.ResponseWriter, r *http.Request) {
		funcName = "eggbox"
		svg(w)
	}
	saddle := func(w http.ResponseWriter, r *http.Request) {
		funcName = "saddle"
		svg(w)
	}
	http.HandleFunc("/", defaultFunc)
	http.HandleFunc("/eggbox", eggbox)
	http.HandleFunc("/saddle", saddle)

	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func svg(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/svg+xml")

	fmt.Fprintf(w, "<svg xmlns='http://www.w3.org/2000/svg' "+
		"style='stroke: grey; fill: white; stroke-width: 0.7' "+
		"width='%d' height='%d'>", width, height)

	z_min, z_max := min_max()

	for i := 0; i < cells; i++ {
		for j := 0; j < cells; j++ {
			ax, ay := corner(i+1, j)
			bx, by := corner(i, j)
			cx, cy := corner(i, j+1)
			dx, dy := corner(i+1, j+1)
			fmt.Fprintf(w, "<polygon style='stroke: %s;' points='%g,%g %g,%g %g,%g %g,%g'/>\n",
				color(i, j, z_min, z_max), ax, ay, bx, by, cx, cy, dx, dy)
		}
	}
	fmt.Fprintln(w, "</svg>")
}

// minmax返回給定x和y的最小值/最大值並假設爲方域的z的最小值和最大值。
func min_max() (min, max float64) {
	min = math.NaN()
	max = math.NaN()
	for i := 0; i < cells; i++ {
		for j := 0; j < cells; j++ {
			for xoff := 0; xoff <= 1; xoff++ {
				for yoff := 0; yoff <= 1; yoff++ {
					_, _, z := getXyz(i+xoff, j+yoff)

					if math.IsNaN(min) || z < min {
						min = z
					}
					if math.IsNaN(max) || z > max {
						max = z
					}
				}
			}
		}
	}
	return min, max
}

func color(i, j int, zmin, zmax float64) string {
	min := math.NaN()
	max := math.NaN()
	for xoff := 0; xoff <= 1; xoff++ {
		for yoff := 0; yoff <= 1; yoff++ {
			_, _, z := getXyz(i+xoff, j+yoff)

			if math.IsNaN(min) || z < min {
				min = z
			}
			if math.IsNaN(max) || z > max {
				max = z
			}
		}
	}

	color := ""
	if math.Abs(max) > math.Abs(min) {
		red := math.Exp(math.Abs(max)) / math.Exp(math.Abs(zmax)) * 255
		if red > 255 {
			red = 255
		}
		color = fmt.Sprintf("#%02x0000", int(red))
	} else {
		blue := math.Exp(math.Abs(min)) / math.Exp(math.Abs(zmin)) * 255
		if blue > 255 {
			blue = 255
		}
		color = fmt.Sprintf("#0000%02x", int(blue))
	}
	return color
}

func corner(i, j int) (float64, float64) {

	// Project (x,y,z) isometrically onto 2-D SVG canvas (sx,sy).
	x, y, z := getXyz(i, j)
	sx := width/2 + (x-y)*cos30*xyscale
	sy := height/2 + (x+y)*sin30*xyscale - z*zscale
	return sx, sy
}

func getXyz(i, j int) (float64, float64, float64) {
	// Find point (x,y) at corner of cell (i,j).
	x := xyrange * (float64(i)/cells - 0.5)
	y := xyrange * (float64(j)/cells - 0.5)

	var z float64

	// Compute surface height z.
	switch funcName {
	case "f":
		z = f(x, y)
	case "eggbox":
		z = eggbox(x, y)
	case "saddle":
		z = saddle(x, y)
	}
	return x, y, z
}

func f(x, y float64) float64 {
	r := math.Hypot(x, y) // distance from (0,0)

	return safeValue(math.Sin(r) / r)
}

func eggbox(x, y float64) float64 { //雞蛋盒
	return 0.2 * (math.Cos(x) + math.Cos(y))
}

func saddle(x, y float64) float64 { //馬鞍
	a := 25.0
	b := 17.0

	return (y*y)/(a*a) - (x*x)/(b*b)
}

func safeValue(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}

	return v
}
