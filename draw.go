// vim: tabstop=4 expandtab


package main

import (
    "fmt"
    "exp/draw/x11"
    "exp/draw"
    "image"
    "math"
)

func abs(n int) int {
    if n>0 { return n }
    return -n
}

type Line struct {
    start image.Point
    end image.Point
}

func (a Line) length() float64 {
    return math.Sqrt(math.Pow(float64(a.start.X - a.end.X), 2) + math.Pow(float64(a.start.Y - a.end.Y), 2))
}

type Drawable interface {
    Draw(draw.Image)
}

// Draw line on the surface
// Uses Bresenham's algorithm
func (line Line) Draw(surface draw.Image) {
    fmt.Println("Line: (", line.start.X, ", ", line.start.Y, ") - (", line.end.X, ", ", line.end.Y, ")")
    start := line.start
    end := line.end
    steep := abs(end.Y - start.Y) > abs(end.X - start.X)
    if steep {
        start.X, start.Y = start.Y, start.X
        end.X, end.Y = end.Y, end.X
    }
    if start.X > end.X {
        start, end = end, start
    }
    deltax := end.X - start.X
    deltay := abs(end.Y - start.Y)
    error := deltax/2
    y := start.Y
    ystep := 1
    if start.Y > end.Y {
        ystep = -1
    }
    for x := start.X; x<end.X; x++ {
        if steep {
            surface.Set(y, x, image.RGBAColor{255, 255, 255, 255})
        } else {
            surface.Set(x, y, image.RGBAColor{255, 255, 255, 255})
        }
        error = error - deltay
        if error < 0 {
            y += ystep
            error += deltax
        }
    }
}

func MouseHandler(mousechan <-chan draw.Mouse) chan image.Point {
    out := make(chan image.Point)
    go func() {
        clicked := false
        for {
            mouse := <-mousechan
            if clicked == false && mouse.Buttons & 1<<0 == 1<<0 { // botao esquerdo
                clicked = true
                fmt.Println("Click: ", mouse.X, ", ", mouse.Y)
                out <- mouse.Point
            }
            if clicked == true && mouse.Buttons & 1<<0 == 0 {
                clicked = false
            }
        }
    }()
    return out
}

func EventProcessor (clickchan <-chan image.Point, kbchan chan int) chan Drawable {
    out := make(chan Drawable)
    go func() {
        for {
            select {
            case keyevent := <-kbchan:
                fmt.Println("Apertou: ", keyevent)
                switch keyevent {
                case 'l':
                    LineCreator(clickchan, kbchan, out)
                    break
                }
            case <-clickchan:
               fmt.Println("Outro clique")
            }
        }
    }()

    return out
}

func LineCreator (clickchan <-chan image.Point, kbchan chan int, out chan<- Drawable) {
    fmt.Println("Desenhar linha")
    pa := [2]image.Point{}
    for i:=0; i<2; i++ {
        fmt.Println("Ponto inicial")
        select {
        case p := <-clickchan:
            fmt.Println("clique para linha")
            pa[i] = p
        case <-kbchan:
            return
        }
    }
    out <- Line{pa[0], pa[1]}
}

// Turns kbchan into a read and writable chan
func RWKBChan (kbchan <-chan int) chan int {
    rwchan := make(chan int);
    go func() {
        for {
            key := <-kbchan;
            if key > 0 {
                rwchan <-key;
            }
        }
    }()
    return rwchan;
}


func main() {
    context, _ := x11.NewWindow()
    context.FlushImage()
    kbchan := RWKBChan(context.KeyboardChan());
    clickchan := MouseHandler(context.MouseChan())
    drawablechan := EventProcessor(clickchan, kbchan)
    for {
        select {
        case object := <-drawablechan:
            object.Draw(context.Screen())
            context.FlushImage()
        case <-context.QuitChan():
            fmt.Println("Quit")
            return
        }
    }
}

