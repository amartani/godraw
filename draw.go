/* Go draw! */
/* vim: set ts=4 sts=4 sw=4 et: */

package main

import (
    "fmt"
    "exp/draw/x11"
    "exp/draw"
    "image"
    "math"
    "container/list"
)

const (
    HMAX = 600
    WMAX = 800
    SEARCH_RADIUS = 6
)

var matrix =  new([WMAX][HMAX]list.List)

func PushMatrix (point image.Point, draw Drawable) {
    matrix[point.X][point.Y].PushFront(draw)
}

func TopMatrix (point image.Point) Drawable {
    element := matrix[point.X][point.Y].Front()
    if element != nil {
        return element.Value.(Drawable)
    }
    return nil
}

func PopMatrix (point image.Point) Drawable {
    element := matrix[point.X][point.Y].Front()
    if element != nil {
        matrix[point.X][point.Y].Remove(element)
        return element.Value.(Drawable)
    }
    return nil
}

func SearchNearPoint (point image.Point) Drawable {
    for radius := 0; radius < SEARCH_RADIUS; radius++ {
        for x := point.X - radius; x <= point.X + radius; x++ {
            for y := point.Y - radius; y <= point.Y + radius; y++ {
                drawable := TopMatrix(image.Point{x, y})
                if drawable != nil {
                    return drawable
                }
            }
        }
    }
    fmt.Println("Objeto nao encontrado");
    return nil
}

var currentColor = image.RGBAColor{255, 255, 255, 255}
var dottedLine   = true

func abs(n int) int {
    if n>0 { return n }
    return -n
}

type Line struct {
    start image.Point
    end image.Point
    color image.RGBAColor
    dotted bool
}

type Poligon struct {
    points *list.List
    color image.RGBAColor
    dotted bool
}

type ColorPoint struct {
    point image.Point
    color image.RGBAColor
}

func (a Line) length() float64 {
    return math.Sqrt(math.Pow(float64(a.start.X - a.end.X), 2) + math.Pow(float64(a.start.Y - a.end.Y), 2))
}

type Drawable interface {
    PointChan() chan ColorPoint
}

func (colorpoint ColorPoint) Valid() bool {
    if colorpoint.point.X == -1 {
        return false
    }
    return true
}

// Draw line on the surface
// Uses Bresenham's algorithm
func (line Line) PointChan() chan ColorPoint {
    pointchan := make(chan ColorPoint)
    go func() {
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
                pointchan <- ColorPoint{image.Point{y, x}, line.color}
            } else {
                pointchan <- ColorPoint{image.Point{x, y}, line.color}
            }
            error = error - deltay
            if error < 0 {
                y += ystep
                error += deltax
            }
        }
        close(pointchan)
    }()
    return pointchan
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

func EventProcessor (clickchan <-chan image.Point, kbchan chan int) chan chan ColorPoint {
    out := make(chan chan ColorPoint)
    go func() {
        for {
            select {
            case keyevent := <-kbchan:
                fmt.Println("Apertou: ", keyevent)
                switch keyevent {
                case 'l':
                    LineCreator(clickchan, kbchan, out)
                    break
                case 'c':
                    SetColor(kbchan)
                    break
                case 'd':
                    DeleteHandler(clickchan, kbchan, out)
                    break
                case 'p':
                    PoligonCreator(clickchan, kbchan, out)
                }
            case <-clickchan:
               fmt.Println("Outro clique")
            }
        }
    }()

   return out
}

func PoligonCreator (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Desenhar Poligono")
    points := new(list.List)
    i := 0
    for_breaker := false
    var p1 image.Point
    var p2 image.Point
    poligon := Poligon{points, currentColor, dottedLine}
    for i = 0 ; i < 50; i++ {
        select {
        case p := <-clickchan:
            fmt.Println("Ponto para poligono")
            points.PushBack(p)
            if i > 0 {
                p1 = p2
                p2 = p
                line := Line{p1, p2, currentColor, dottedLine}
                out <- RegisterPoints(line.PointChan(), poligon)
            } else {
                p2 = p
            }
        case <- kbchan:
            for_breaker = true
            break
        }
        if for_breaker {
            break
        }
    }
    if i > 0 {
        line := Line{points.Back().Value.(image.Point), points.Front().Value.(image.Point), currentColor, dottedLine}
        out <- RegisterPoints(line.PointChan(), poligon)
    }
}

func (poligon Poligon) PointChan() chan ColorPoint {
    outchan := make(chan ColorPoint)
    go func() {
        points := poligon.points.Iter()
        first := (<-points).(image.Point)
        fmt.Println("Passou :)");
        before := first
        var after image.Point
        for ! closed(points) {
            aftertemp := <-points
            if aftertemp == nil { break }
            after = aftertemp.(image.Point)
            line := Line{before, after, poligon.color, poligon.dotted}
            linechan := line.PointChan()
            for ! closed(linechan) {
                outchan <- <- linechan
            }
            before = after
        }
        // Line to close poligon
        line := Line{before, first, poligon.color, poligon.dotted}
        linechan := line.PointChan()
        for ! closed(linechan) {
            outchan <- <- linechan
        }
        close(outchan)
    }()
    return outchan
}

func DeleteHandler (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Apagar objeto")
    for {
    select {
        case p := <-clickchan:
            drawable := SearchNearPoint(p)
            if drawable != nil {
                go Delete(drawable, out)
                return
            }
        case <-kbchan:
            return
    }
    }
}

func Delete(drawable Drawable, out chan chan ColorPoint) {
    blackpoints := make(chan ColorPoint)
    out <- blackpoints
    colorpoints := drawable.PointChan()
    for ! closed(colorpoints) {
        point := <-colorpoints
        point.color = image.RGBAColor{0, 0, 0, 0}
        blackpoints <- point
    }
    close(blackpoints)
}


func LineCreator (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Desenhar linha")
    pa := [2]image.Point{}
    for i:=0; i<2; i++ {
        select {
        case p := <-clickchan:
            pa[i] = p
        case <-kbchan:
            return
        }
    }
    line := Line{pa[0], pa[1], currentColor, dottedLine}
    out <- RegisterPoints(line.PointChan(), line)
}

func SetColor (kbchan chan int) {
    switch <- kbchan {
    case 'r':
        currentColor = image.RGBAColor{255, 0, 0, 255}
        fmt.Println("Vermelho selecionado")
        break
    case 'g':
        currentColor = image.RGBAColor{0, 255, 0, 255}
        fmt.Println("Verde selecionado")
        break
    case 'b':
        currentColor = image.RGBAColor{0, 0, 255, 255}
        fmt.Println("Azul selecionado")
        break
    case 'w':
        currentColor = image.RGBAColor{255, 255, 255, 255}
        fmt.Println("Branco selecionado")
    }
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

func Draw (surface draw.Image, pointchan chan ColorPoint) {
    for ! closed(pointchan) {
        colorpoint := <-pointchan
        surface.Set(colorpoint.point.X, colorpoint.point.Y, colorpoint.color)
    }
}

func RegisterPoints (pointchan chan ColorPoint, drawable Drawable) chan ColorPoint {
    outchan := make(chan ColorPoint)
    go func() {
        for ! closed(pointchan) {
            colorpoint := <-pointchan
            PushMatrix(colorpoint.point, drawable)
            outchan <- colorpoint
        }
        close (outchan)
    }()
    return outchan
}

func main() {
    context, _ := x11.NewWindow()
    context.FlushImage()
    kbchan := RWKBChan(context.KeyboardChan());
    clickchan := MouseHandler(context.MouseChan())
    colorpointchanchan := EventProcessor(clickchan, kbchan)
    for {
        select {
        case colorpointchan := <-colorpointchanchan:
            Draw(context.Screen(), colorpointchan)
            context.FlushImage()
        case <-context.QuitChan():
            fmt.Println("Quit")
            return
        }
    }
}

