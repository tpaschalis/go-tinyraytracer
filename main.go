package main

import "fmt"
import "image"
import "image/color"
import "image/png"
import "os"

func render() {
    const w = 1024
    const h = 768

    img := image.NewRGBA(image.Rect(0, 0, w, h))
    for j:=0; j<h; j++ {
        for i:=0; i<w; i++ {
            img.Set(i, j, color.RGBA{uint8(255*j/h), uint8(255*i/w), 0, 255})
        }
    }

    f, _ := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
    defer f.Close()
    png.Encode(f, img)


    fmt.Println("Hey!")
}

func main() {

    render()

}
