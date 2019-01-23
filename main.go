package main

import "fmt"
import "image"
import "image/color"
import "image/png"
import "os"
import "github.com/golang/geo/r3"
import "math"

const (
	XAxis r3.Axis = iota
	YAxis
	ZAxis
)

type Materials struct {
	DiffusedColor color.RGBA
}

type Sphere struct {
	Center   r3.Vector
	Radius   float64
	Material Materials
}

func (s Sphere) ray_intersect(orig, dir r3.Vector, t0 *float64) bool {
	L := r3.Vector.Sub(s.Center, orig)
	tca := r3.Vector.Dot(L, dir)
	d2 := r3.Vector.Dot(L, L) - tca*tca

	if d2 > s.Radius*s.Radius {
		return false
	}

	thc := math.Sqrt(s.Radius*s.Radius - d2)
	*t0 = tca - thc
	t1 := tca + thc

	if *t0 < 0.0 {
		*t0 = t1
	}
	if *t0 < 0.0 {
		return false
	}
	return true
}

func scene_intersect(orig, dir r3.Vector, spheres []Sphere, hit, N *r3.Vector, material *Materials) bool {
	spheres_dist := math.MaxFloat64
	for i := 0; i < len(spheres); i++ {
        dist_i := 0.0
		if spheres[i].ray_intersect(orig, dir, &dist_i) && dist_i < spheres_dist {
			spheres_dist = dist_i
			*hit = r3.Vector.Add(orig, r3.Vector.Mul(dir, dist_i))
			*N = r3.Vector.Normalize(r3.Vector.Sub(*hit, spheres[i].Center))
			*material = spheres[i].Material // Declared and not used for commit 3
			_ = material
		}
	}
	return spheres_dist < 1000
}

func cast_ray(orig, dir r3.Vector, spheres []Sphere) color.RGBA {
	var point, N r3.Vector
	var material Materials

	if !scene_intersect(orig, dir, spheres, &point, &N, &material) {
		return color.RGBA{50, 180, 205, 255}
	}
	return material.DiffusedColor
}

func render(spheres []Sphere) {

	const w = 1024.0
	const h = 768.0
	const fov = math.Pi / 2

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			//img.Set(i, j, color.RGBA{uint8(255*j/h), uint8(255*i/w), 0, 255})
			var x, y float64
			x = (2.0*(float64(i)+0.5)/w - 1.0) * math.Tan(fov/2.0) * w / h
			y = -(2.0*(float64(j)+0.5)/h - 1.0) * math.Tan(fov/2.0)
			dir := r3.Vector.Normalize(r3.Vector{x, y, -1.0})
			img.Set(i, j, cast_ray(r3.Vector{0, 0, 0}, dir, spheres))
		}
	}

	f, _ := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, img)

	fmt.Println("Hey!")
}

func main() {

	ivory := Materials{DiffusedColor: color.RGBA{100, 100, 75, 255}}
	red_rubber := Materials{DiffusedColor: color.RGBA{75, 25, 25, 255}}

	spheres := make([]Sphere, 0)
	spheres = append(spheres, Sphere{r3.Vector{-3.0, 0.0, -16.0}, 2, ivory})
	spheres = append(spheres, Sphere{r3.Vector{-1.0, -1.5, -12.0}, 2, red_rubber})
	spheres = append(spheres, Sphere{r3.Vector{1.5, -0.5, -18.0}, 3, red_rubber})
	spheres = append(spheres, Sphere{r3.Vector{7.0, 5.0, -18.0}, 4, ivory})



	render(spheres)
}
