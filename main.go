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

type Light struct {
    Position r3.Vector
    Intensity float64
}

type Materials struct {
    Albedo []float64
	DiffusedColor color.RGBA
    SpecularExponent float64
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

func cast_ray(orig, dir r3.Vector, spheres []Sphere, lights []Light) color.RGBA {
	var point, N r3.Vector
	var material Materials

	if !scene_intersect(orig, dir, spheres, &point, &N, &material) {
		return color.RGBA{50, 180, 205, 255}
	}
	//return material.DiffusedColor
    diffuse_light_intensity, specular_light_intensity := 0.0, 0.0
    for i:=0; i<len(lights); i++ {
        light_dir := r3.Vector.Normalize(r3.Vector.Sub(lights[i].Position, point))

        diffuse_light_intensity += lights[i].Intensity * max(0, r3.Vector.Dot(light_dir, N))
        m_light_dir := r3.Vector.Mul(light_dir, -1.)
        specular_light_intensity += math.Pow( max(0., r3.Vector.Dot(r3.Vector.Mul(Reflect(m_light_dir, N), -1), dir)), material.SpecularExponent) * lights[i].Intensity
    }
    //return multiplyColorIntensity(material.DiffusedColor, diffuse_light_intensity)
    res1x := float64(material.DiffusedColor.R) * diffuse_light_intensity * material.Albedo[0]
    res1y := float64(material.DiffusedColor.G) * diffuse_light_intensity * material.Albedo[0]
    res1z := float64(material.DiffusedColor.B) * diffuse_light_intensity * material.Albedo[0]
    black := color.RGBA{255, 255, 255, 255}
    res2x := float64(black.R) * specular_light_intensity * material.Albedo[1]
    res2y := float64(black.G) * specular_light_intensity * material.Albedo[1]
    res2z := float64(black.B) * specular_light_intensity * material.Albedo[1]
    return AddColors(r3.Vector{res1x, res1y, res1z}, r3.Vector{res2x, res2y, res2z})
}

func render(spheres []Sphere, lights []Light) {

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
			img.Set(i, j, cast_ray(r3.Vector{0, 0, 0}, dir, spheres, lights))
		}
	}

	f, _ := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, img)

	fmt.Println("Hey!")
}


func max(a, b float64) float64 {
    if a >= b {
        return a
    }
    return b
}

func Reflect (I, N r3.Vector) r3.Vector {
    return r3.Vector.Sub(I, r3.Vector.Mul(N,2.0*r3.Vector.Dot(I, N)))
}

func AddColors(i, j r3.Vector) color.RGBA {
    r, g, b := i.X+j.X, i.Y+j.Y, i.Z+j.Z
    maxc := float64(max(float64(r), max(float64(g), float64(b))))
    if maxc > 255. {
        return color.RGBA{uint8(float64(r)*255./maxc),
                        uint8(float64(g)*255./maxc),
                        uint8(float64(b)*255./maxc),
                        255}
    }
    return color.RGBA{uint8(r),
                    uint8(g),
                    uint8(b),
                    255}
}

func main() {

	ivory := Materials{Albedo: []float64{0.3, 0.6}, DiffusedColor: color.RGBA{100, 100, 75, 255}, SpecularExponent: 50.}
	red_rubber := Materials{Albedo: []float64{0.9, 0.1}, DiffusedColor: color.RGBA{75, 25, 25, 255}, SpecularExponent: 10.}

	spheres := make([]Sphere, 0)
	spheres = append(spheres, Sphere{r3.Vector{-3.0, 0.0, -16.0}, 2, ivory})
	spheres = append(spheres, Sphere{r3.Vector{-1.0, -1.5, -12.0}, 2, red_rubber})
	spheres = append(spheres, Sphere{r3.Vector{1.5, -0.5, -18.0}, 3, red_rubber})
	spheres = append(spheres, Sphere{r3.Vector{7.0, 5.0, -18.0}, 4, ivory})

    lights := make([]Light, 0)
    lights = append(lights, Light{r3.Vector{-20, 20, 20}, 1.5})
    lights = append(lights, Light{r3.Vector{30, 50, -25}, 1.8})
    lights = append(lights, Light{r3.Vector{30, 20, 30}, 1.7})

	render(spheres, lights)
}

func multiplyColorIntensity(c color.RGBA, f float64) color.RGBA {
    return color.RGBA{uint8(float64(c.R)*f),
                      uint8(float64(c.G)*f),
                      uint8(float64(c.B)*f),
                      255}
}

