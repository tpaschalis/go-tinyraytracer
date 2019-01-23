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

type Sphere struct {
    Center r3.Vector
    Radius float64
}



func (s Sphere) ray_intersect(orig, dir r3.Vector, t0 float64) bool{
    L := r3.Vector.Sub(s.Center, orig)
    tca := r3.Vector.Dot(L, dir)
    d2 := r3.Vector.Dot(L, L) - tca*tca

    if d2 > s.Radius*s.Radius {
        return false
    }

    thc := math.Sqrt(s.Radius*s.Radius - d2)
    t0 = tca - thc
    t1 := tca + thc

    if t0 < 0 {
        t0 = t1
    }
    if t0 < 0 {
        return false
    }
    return true
}

func cast_ray(orig, dir r3.Vector, sphere Sphere) color.RGBA {
    sphere_dist := math.MaxFloat64
    if (!sphere.ray_intersect(orig, dir, sphere_dist)) {
        return color.RGBA{50, 180, 205, 255}
    }

    return color.RGBA{100, 100, 75, 255}
}

func render(sphere Sphere) {

    a := r3.Vector{X:1.3, Y:4.5, Z:1.1}
    b := r3.Vector{1.3, -1.5, -1.1}
    fmt.Println(r3.Vector.Add(a, b))
    const w = 1024.0
    const h = 768.0
    const fov = math.Pi/2

    img := image.NewRGBA(image.Rect(0, 0, w, h))
    for j:=0; j<h; j++ {
        for i:=0; i<w; i++ {
            //img.Set(i, j, color.RGBA{uint8(255*j/h), uint8(255*i/w), 0, 255})
            var x, y float64
            x =  (2.0*(float64(i) + 0.5)/w  - 1.0)*math.Tan(fov/2.0)*w/h;
            y = -(2.0*(float64(j) + 0.5)/h - 1.0)*math.Tan(fov/2.0);
            dir := r3.Vector{x, y, -1.0}
            dir = r3.Vector.Normalize(dir)
            img.Set(i, j, cast_ray(r3.Vector{0, 0, 0}, dir, sphere))
        }
    }

    f, _ := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
    defer f.Close()
    png.Encode(f, img)


    fmt.Println("Hey!")
}

func main() {

    sphere := Sphere{r3.Vector{-3, 0, -16}, 2}
    render(sphere)
}
