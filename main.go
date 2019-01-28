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
	Position  r3.Vector
	Intensity float64
}

type Materials struct {
	RefractiveIndex  float64
	Albedo           []float64
	DiffusedColor    color.RGBA
	SpecularExponent float64
}

type Sphere struct {
	Center   r3.Vector
	Radius   float64
	Material Materials
}

func (s Sphere) rayIntersect(orig, dir r3.Vector, t0 *float64) bool {
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

func sceneIntersect(orig, dir r3.Vector, spheres []Sphere, hit, N *r3.Vector, material *Materials) bool {
	spheresDist := math.MaxFloat64
	chessboardDist := math.MaxFloat64

	for i := 0; i < len(spheres); i++ {
		distI := 0.0
		if spheres[i].rayIntersect(orig, dir, &distI) && distI < spheresDist {
			spheresDist = distI
			*hit = r3.Vector.Add(orig, r3.Vector.Mul(dir, distI))
			*N = r3.Vector.Normalize(r3.Vector.Sub(*hit, spheres[i].Center))
			*material = spheres[i].Material
			_ = material
		}
	}
	//return spheresDist < 1000
	if math.Abs(dir.Y) > 0.001 {
		d := -(orig.Y + 4.) / dir.Y // the checkerboard plane has equation y = -4
		pt := r3.Vector.Add(orig, r3.Vector.Mul(dir, d))

		if d > 0 && math.Abs(pt.X) < 10. && pt.Z < -10. && pt.Z > -30. && d < spheresDist {
			chessboardDist = d
			*hit = pt
			*N = r3.Vector{0, 1, 0}
			if (int(.5*hit.X+1000)+int(.5*hit.Z))&1 == 1 {
				// need to re-define 'material', to a valid one, because it is local
				// to the `if` statement above, and we get a null-panic like, eg. when accessing albedo
				*material = spheres[0].Material
				material.DiffusedColor = color.RGBA{255, 255, 255, 255}
			} else {
				*material = spheres[0].Material
				material.DiffusedColor = color.RGBA{255, 177, 65, 255}
			}
		}
	}

	return min(spheresDist, chessboardDist) < 1000

}

func castRay(orig, dir r3.Vector, spheres []Sphere, lights []Light, depth int) color.RGBA {
	var point, N r3.Vector
	var material Materials
	depth += 1
	if depth > 4 || !sceneIntersect(orig, dir, spheres, &point, &N, &material) {
		//return color.RGBA{50, 180, 205, 255}
		return color.RGBA{50, 180, 205, 255}
	}

	var reflectOrig r3.Vector
	var refractOrig r3.Vector

	reflectDir := r3.Vector.Normalize(Reflect(dir, N))
	refractDir := r3.Vector.Normalize(Refract(dir, N, material.RefractiveIndex))

	if r3.Vector.Dot(reflectDir, N) < 0 {
		reflectOrig = r3.Vector.Sub(point, r3.Vector.Mul(N, 0.001))
	} else {
		reflectOrig = r3.Vector.Add(point, r3.Vector.Mul(N, 0.001))
	}

	if r3.Vector.Dot(refractDir, N) < 0 {
		refractOrig = r3.Vector.Sub(point, r3.Vector.Mul(N, 0.001))
	} else {
		refractOrig = r3.Vector.Add(point, r3.Vector.Mul(N, 0.001))
	}

	reflectColor := castRay(reflectOrig, reflectDir, spheres, lights, depth)
	refractColor := castRay(refractOrig, refractDir, spheres, lights, depth)

	diffuseLightIntensity, specularLightIntensity := 0.0, 0.0
	for i := 0; i < len(lights); i++ {
		lightDir := r3.Vector.Normalize(r3.Vector.Sub(lights[i].Position, point))
		lightDistance := r3.Vector.Norm(r3.Vector.Sub(lights[i].Position, point))

		var shadowOrig r3.Vector
		if r3.Vector.Dot(lightDir, N) < 0 {
			shadowOrig = r3.Vector.Sub(point, r3.Vector.Mul(N, 0.001))
		} else {
			shadowOrig = r3.Vector.Add(point, r3.Vector.Mul(N, 0.001))
		}

		var shadowPt, shadowN r3.Vector
		var tmpmaterial Materials

		if sceneIntersect(shadowOrig, lightDir, spheres, &shadowPt, &shadowN, &tmpmaterial) && r3.Vector.Norm(r3.Vector.Sub(shadowPt, shadowOrig)) < lightDistance {
			continue
		}

		diffuseLightIntensity += lights[i].Intensity * max(0, r3.Vector.Dot(lightDir, N))
		mLightDir := r3.Vector.Mul(lightDir, -1.)
		specularLightIntensity += math.Pow(max(0., r3.Vector.Dot(r3.Vector.Mul(Reflect(mLightDir, N), -1), dir)), material.SpecularExponent) * lights[i].Intensity
	}
	res1x := float64(material.DiffusedColor.R) * diffuseLightIntensity * material.Albedo[0]
	res1y := float64(material.DiffusedColor.G) * diffuseLightIntensity * material.Albedo[0]
	res1z := float64(material.DiffusedColor.B) * diffuseLightIntensity * material.Albedo[0]
	black := color.RGBA{255, 255, 255, 255}
	res2x := float64(black.R) * specularLightIntensity * material.Albedo[1]
	res2y := float64(black.G) * specularLightIntensity * material.Albedo[1]
	res2z := float64(black.B) * specularLightIntensity * material.Albedo[1]

	res3x := float64(reflectColor.R) * material.Albedo[2]
	res3y := float64(reflectColor.G) * material.Albedo[2]
	res3z := float64(reflectColor.B) * material.Albedo[2]

	res4x := float64(refractColor.R) * material.Albedo[3]
	res4y := float64(refractColor.G) * material.Albedo[3]
	res4z := float64(refractColor.B) * material.Albedo[3]

	return AddColors(r3.Vector{res1x, res1y, res1z}, r3.Vector{res2x, res2y, res2z}, r3.Vector{res3x, res3y, res3z}, r3.Vector{res4x, res4y, res4z})
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
			img.Set(i, j, castRay(r3.Vector{0, 0, 0}, dir, spheres, lights, 0))
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

func Reflect(I, N r3.Vector) r3.Vector {
	return r3.Vector.Sub(I, r3.Vector.Mul(N, 2.0*r3.Vector.Dot(I, N)))
}

func Refract(I, N r3.Vector, refractiveIdx float64) r3.Vector {
	// Snell's law
	cosi := -max(-1., min(1, r3.Vector.Dot(I, N)))
	etai, etat := 1., refractiveIdx
	n := N
	if cosi < 0. {
		// if the ray is inside the object, swap the indices and invert the normal to get the correct result
		cosi = -cosi
		etai, etat = swap(etai, etat)
		n = r3.Vector.Mul(N, -1)
	}
	eta := etai / etat
	k := 1. - eta*eta*(1.-cosi*cosi)
	if k < 0. {
		return r3.Vector{0., 0., 0.}
	} else {
		return r3.Vector.Add(r3.Vector.Mul(I, eta), r3.Vector.Mul(n, (eta*cosi-math.Sqrt(k))))
	}
}

func AddColors(i, j, k, l r3.Vector) color.RGBA {
	r, g, b := (i.X + j.X + k.X + l.X), (i.Y + j.Y + k.Y + l.Y), (i.Z + j.Z + k.Z + l.Z)
	maxc := float64(max(float64(r), max(float64(g), float64(b))))
	if maxc > 255. {
		return color.RGBA{uint8(float64(r) * 255. / maxc),
			uint8(float64(g) * 255. / maxc),
			uint8(float64(b) * 255. / maxc),
			255}
	}
	return color.RGBA{uint8(r),
		uint8(g),
		uint8(b),
		255}
}

func main() {

	ivory := Materials{RefractiveIndex: 1.0, Albedo: []float64{0.3, 0.6, 0.1, 0.0}, DiffusedColor: color.RGBA{100, 100, 75, 255}, SpecularExponent: 50.}
	glass := Materials{RefractiveIndex: 1.5, Albedo: []float64{0.0, 0.5, 0.1, 0.8}, DiffusedColor: color.RGBA{255, 255, 255, 255}, SpecularExponent: 1425.}
	redRubber := Materials{RefractiveIndex: 1.0, Albedo: []float64{0.9, 0.1, 0.0, 0.0}, DiffusedColor: color.RGBA{76, 25, 25, 255}, SpecularExponent: 10.}
	mirror := Materials{RefractiveIndex: 1.0, Albedo: []float64{0.0, 10.0, 0.8, 0.0}, DiffusedColor: color.RGBA{255, 255, 255, 255}, SpecularExponent: 1425.}

	spheres := make([]Sphere, 0)
	spheres = append(spheres, Sphere{r3.Vector{-3.0, 0.0, -16.0}, 2, ivory})
	spheres = append(spheres, Sphere{r3.Vector{-1.0, -1.5, -12.0}, 2, glass})
	spheres = append(spheres, Sphere{r3.Vector{1.5, -0.5, -18.0}, 3, redRubber})
	spheres = append(spheres, Sphere{r3.Vector{7.0, 5.0, -18.0}, 4, mirror})

	lights := make([]Light, 0)
	lights = append(lights, Light{r3.Vector{-20, 20, 20}, 1.5})
	lights = append(lights, Light{r3.Vector{30, 50, -25}, 1.8})
	lights = append(lights, Light{r3.Vector{30, 20, 30}, 1.7})

	render(spheres, lights)
}

func multiplyColorIntensity(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * f),
		uint8(float64(c.G) * f),
		uint8(float64(c.B) * f),
		255}
}

func swap(a, b float64) (float64, float64) {
	return b, a
}
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
