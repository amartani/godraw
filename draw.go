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
    SEARCH_RADIUS = 10
    SIDE_RATIO = 0.4 // Ratio for circle sides per radius
)

var counter_id = 0

var matrix =  new([WMAX][HMAX]list.List)

/* Functions for the Matrix */

func PushMatrix (color_point ColorPoint) {
    matrix[color_point.point.X][color_point.point.Y].PushFront(color_point)
}

func TopMatrix (point image.Point) Drawable {
    element := matrix[point.X][point.Y].Front()
    if element != nil {
        return *(element.Value.(ColorPoint).drawable)
    }
    return nil
}

func TopMatrixColorPoint (point image.Point) ColorPoint {
    element := matrix[point.X][point.Y].Front()
    if element != nil {
        return element.Value.(ColorPoint)
    }
    return ColorPoint{point, image.RGBAColor{0, 0, 0, 255}, nil}
}

func PopMatrix (point image.Point) Drawable {
    element := matrix[point.X][point.Y].Front()
    if element != nil {
        matrix[point.X][point.Y].Remove(element)
        return *(element.Value.(ColorPoint).drawable)
    }
    return nil
}

func RemoveFromMatrix(point image.Point, drawable interface{}) {
    list := &matrix[point.X][point.Y]
    elem := SearchList(list, drawable)
    if (elem != nil) {
        list.Remove(elem)
    }
}

func ListMatrix(point image.Point) *list.List {
    return &matrix[point.X][point.Y]
}

/* End funcions for the Matrix */

/* Helper functions for list.List */

func SearchList(list *list.List, subject interface{}) *list.Element {
    for elem := list.Front(); elem != nil; elem = elem.Next() {
        // if reflect.Typeof(elem.Value) != reflect.Typeof(subject) {
        //     continue
        // }
        if elem.Value.(ColorPoint).drawable.GetId() == subject.(Drawable).GetId() {
        //    fmt.Println("Found ", elem.Value.(ColorPoint).drawable.GetId(), ":", subject.(Drawable).GetId())
            return elem
        }
    }
    fmt.Println("Not found")
    return nil
}

// func MergeLists(list *list.List, list2 *list.List) {
//     for elem := list2.Front(); elem != nil; elem = elem.Next() {
//         if SearchList(list, elem.Value) == nil {
//             list.PushFront(elem.Value)
//         }
//     }
// }

/* End helper funcions for list.List */

/* Helper functions for image.Point */

func PointsDistance(p1 image.Point, p2 image.Point) float64 {
    return math.Sqrt(math.Pow(float64(p1.X - p2.X), 2) + math.Pow(float64(p1.Y - p2.Y), 2))
}

/* End helper functions for image.Point */

func SearchNearPoint (point image.Point) (Drawable, image.Point) {
    for radius := 0; radius < SEARCH_RADIUS; radius++ {
        for x := point.X - radius; x <= point.X + radius; x++ {
            for y := point.Y - radius; y <= point.Y + radius; y++ {
                drawable := TopMatrix(image.Point{x, y})
                if drawable != nil {
                    fmt.Println("Objeto encontrado")
                    return drawable, image.Point{x, y}
                }
            }
        }
    }
    fmt.Println("Objeto nao encontrado")
    return nil, point
}

var currentCounter = 0
var currentColor     = image.RGBAColor{255, 255, 255, 255}
var currentDashStyle = 0
var currentThick     = false

const (
    SOLID = 0
    DOTTED = 1
    DASHED = 2
)


func abs(n int) int {
    if n>0 { return n }
    return -n
}

type Line struct {
    start image.Point
    end image.Point
    color image.RGBAColor
    dotted int
    thick bool
    Id
}

type Poligon struct {
    points *list.List
    color image.RGBAColor
    dotted int
    thick bool
    Id
}

type RegularPoligon struct {
    origin  image.Point
    start   image.Point
    sides   int
    Id
}

/* Circle definition */

type Circle struct {
    center  image.Point
    start   image.Point
    Id
}

/* Circle methods */

func (circle *Circle) PointChan() chan ColorPoint {
    radius := PointsDistance(circle.start, circle.center)
    sides := int(SIDE_RATIO * radius)
    fmt.Println("DEBUG: Circle sides: %d", sides)
    regpol := RegularPoligon{circle.center, circle.start, sides, Id{0}}
    return regpol.PointChan()
}

func (circle *Circle) Move(delta image.Point) {
    circle.center   = circle.center.Add(delta)
    circle.start    = circle.start.Add(delta)
}

type ColorPoint struct {
    point image.Point
    color image.RGBAColor
    drawable *Drawable
}

type Drawable interface {
    PointChan() chan ColorPoint
    GetId() int
    SetId(int)
    Move(image.Point)
    RotatePoints(image.Point, float64)
}

type Id struct {
    id int
}

func (w *Id) GetId() int {
    return w.id
}

func (w *Id) SetId(id int) {
    w.id = id
}

func (line *Line) RotatePoints(origin image.Point, angle float64){
    line.start = RotatePoint(line.start, origin, angle)
    line.end   = RotatePoint(line.end, origin, angle)
}

func (poligon *Poligon) RotatePoints(origin image.Point, angle float64){
    for elem := poligon.points.Front(); elem != nil; elem = elem.Next() {
        point := elem.Value.(image.Point)
        elem.Value = RotatePoint(point, origin, angle)
    }
}

func (reg *RegularPoligon) RotatePoints(origin image.Point, angle float64){
    reg.start  = RotatePoint(reg.start, origin, angle)
    reg.origin = RotatePoint(reg.origin, origin, angle)
}

func (colorpoint ColorPoint) Valid() bool {
    point := colorpoint.point
    if point.X < 0 {
        return false
    }
    if point.Y < 0 {
        return false
    }
    if point.X >= WMAX {
        return false
    }
    if point.Y >= HMAX {
        return false
    }
    return true
}

// Draw line on the surface
// Uses Bresenham's algorithm
func (line *Line) PointChan() chan ColorPoint {
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
        // For dashed line
        segments := deltax/10
        if segments % 2 == 0 { segments++ }
        for x := start.X; x<end.X; x++ {
            progress := x - start.X
            showpoint := true
            if line.dotted == DOTTED {
                if progress % 4 >= 2 {
                    showpoint = false
                }
            }
            if line.dotted == DASHED {
                if (progress * segments) % (2 * deltax) > 10 * segments {
                    showpoint = false
                }
            }
            if showpoint {
                if steep {
                    if line.thick {
                        pointchan <- ColorPoint{image.Point{y+1, x}, line.color, nil}
                    }
                    pointchan <- ColorPoint{image.Point{y, x}, line.color, nil}
                } else {
                    if line.thick {
                        pointchan <- ColorPoint{image.Point{x, y+1}, line.color, nil}
                    }
                    pointchan <- ColorPoint{image.Point{x, y}, line.color, nil}
                }
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

func (line *Line) Move (dest image.Point) {
    line.start = line.start.Add(dest)
    line.end   = line.end.Add(dest)
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

func FilterInvalidPoints(in chan ColorPoint) (out chan ColorPoint) {
    out = make(chan ColorPoint)
    go func() {
        for ! closed(in) {
            point := <-in
            if point.Valid() {
                out <- point
            } else {
            }
        }
        close(out)
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
                case 'r':
                    RegularPoligonCreator(clickchan, kbchan, out, currentCounter)
                case 'o':
                    CircleCreator(clickchan, kbchan, out)
                case '+':
                    currentCounter++
                    fmt.Println("Contador Generico: ", currentCounter)
                case '-':
                    currentCounter--
                    fmt.Println("Contador Generico: ", currentCounter)
                case 'm':
                    MoveHandler(clickchan, kbchan, out)
                case 't':
                    DashHandler()
                case 'b':
                    ThickHandler()
                case 'g':
                    RotateHandler(clickchan, kbchan, out)
                case 'x':
                    origin := image.Point{10, 10}
                    point := image.Point{20, 10}
                    angle := math.Pi/float64(int(2))
                    fmt.Println("Origem:", origin, " Ponto:", point, " Angulo:", angle*180/math.Pi, " Rotacionado:", RotatePoint(point, origin, angle))
                }
            case <-clickchan:
               fmt.Println("Outro clique")
            }
        }
    }()

   return out
}

func DashHandler() {
    currentDashStyle ++
    currentDashStyle %= 3
    switch currentDashStyle {
    case SOLID:
        fmt.Println("Style: solid")
    case DOTTED:
        fmt.Println("Style: dotted")
    case DASHED:
        fmt.Println("Style: dashed")
    }
}

func ThickHandler() {
    if currentThick {
        currentThick = false
        fmt.Println("Thick: no")
    } else {
        currentThick = true
        fmt.Println("Thick: yes")
    }
}

func CircleCreator (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Desenhar Circulo")
    points := [2]image.Point{}
    for_breaker := false
    for i := 0; i < 2; i++ {
        select {
        case p := <-clickchan:
            fmt.Println("Ponto para circulo")
            points[i] = p
        case <-kbchan:
            for_breaker = true
            break
        }
        if for_breaker {
            break
        }
    }
    counter_id++
    circle := Circle{points[0], points[1], Id{counter_id}}
    out <- RegisterPoints(FilterInvalidPoints(circle.PointChan()), &circle)
}

func RegularPoligonCreator (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint, sides int) {
    if sides < 3 {
        fmt.Println("Numero de lados invalido, lados:", sides)
		return
    }
    fmt.Println("Desenhar Poligono Regular")
    points := [2]image.Point{}
    for_breaker := false
    for i := 0; i < 2; i++ {
        select {
        case p := <-clickchan:
            fmt.Println("Ponto para poligono regular")
            points[i] = p
        case <-kbchan:
            for_breaker = true
            break
        }
        if for_breaker {
            break
        }
    }
    counter_id++
    regpol := RegularPoligon{points[0], points[1], sides, Id{counter_id}}
    out <- RegisterPoints(FilterInvalidPoints(regpol.PointChan()), &regpol)
}

func (regpol *RegularPoligon) PointChan() chan ColorPoint {
    start := regpol.start
    origin := regpol.origin
    sides := regpol.sides
    radius := start.Sub(origin)
    start_ang := math.Atan(float64(int(radius.Y))/float64(int(radius.X)))
    if radius.X < 0 { start_ang -= math.Pi }
    fmt.Println("Angulo inicial: ", start_ang*180/math.Pi, " Origem: ", origin, " Inicio:", start, " Vetor Inicial:", radius)
    module := math.Sqrt(math.Pow(float64(int(radius.X)), 2)+math.Pow(float64(int(radius.Y)), 2))
    theta := 2*math.Pi/float64(int(sides))
    poli_points := new(list.List)
    for i := 0; i < sides; i++ {
        p := origin.Add(image.Point{int(float64(module*math.Cos(float64(int(i))*theta+start_ang))), int(float64(module*math.Sin(float64(int(i))*theta+start_ang)))})
        poli_points.PushBack(p)
//        fmt.Println("Ponto: ", p)
    }
    poligon := Poligon{poli_points, currentColor, currentDashStyle, currentThick, Id{0}}
    return poligon.PointChan()
}

func (regpol *RegularPoligon) Move(delta image.Point) {
    regpol.origin = regpol.origin.Add(delta)
    regpol.start  = regpol.start .Add(delta)
}

func PoligonCreator (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Desenhar Poligono")
    points := new(list.List)
    i := 0
    for_breaker := false
    var p1 image.Point
    var p2 image.Point
    poligon := Poligon{points, currentColor, currentDashStyle, currentThick, Id{0}}
    for i = 0 ; i < 50; i++ {
        select {
        case p := <-clickchan:
            fmt.Println("Ponto para poligono")
            points.PushBack(p)
            if i > 0 {
                p1 = p2
                p2 = p
                line := Line{p1, p2, currentColor, currentDashStyle, currentThick, Id{0}}
                out <- RegisterPoints(FilterInvalidPoints(line.PointChan()), &poligon)
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
        line := Line{points.Back().Value.(image.Point), points.Front().Value.(image.Point), currentColor, currentDashStyle, currentThick, Id{0}}
        out <- RegisterPoints(FilterInvalidPoints(line.PointChan()), &poligon)
    }
    counter_id++
    (&poligon).SetId(counter_id)
}

func RotatePoint(point image.Point, origin image.Point, angle float64) image.Point {
    delta := point.Sub(origin)
    x := float64(int(delta.X))*math.Cos(angle) - float64(int(delta.Y))*math.Sin(angle)
    y := float64(int(delta.X))*math.Sin(angle) + float64(int(delta.Y))*math.Cos(angle)
    return image.Point{int(float64(x)), int(float64(y))}.Add(origin)
}

func (poligon *Poligon) PointChan() chan ColorPoint {
    outchan := make(chan ColorPoint)
    go func() {
        points := poligon.points.Iter()
        first := (<-points).(image.Point)
        before := first
        var after image.Point
        for ! closed(points) {
            aftertemp := <-points
            if aftertemp == nil { break }
            after = aftertemp.(image.Point)
            line := Line{before, after, poligon.color, poligon.dotted, poligon.thick, Id{0}}
            linechan := line.PointChan()
            for ! closed(linechan) {
                outchan <- <- linechan
            }
            before = after
        }
        // Line to close poligon
        line := Line{before, first, poligon.color, poligon.dotted, poligon.thick, Id{0}}
        linechan := line.PointChan()
        for ! closed(linechan) {
            outchan <- <- linechan
        }
        close(outchan)
    }()
    return outchan
}

func (poligon *Poligon) Move(delta image.Point) {
    for elem := poligon.points.Front(); elem != nil; elem = elem.Next() {
        point := elem.Value.(image.Point)
        point = point.Add(delta)
        elem.Value = point
    }
}

func RotateHandler (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Rotacionar objeto")
    state := 0
    var drawable Drawable
    var origin image.Point
    var point1 image.Point
    var point2 image.Point
    for {
        select{
        case p := <-clickchan:
            switch(state){
            case 0:
                drawable, _ = SearchNearPoint(p)
                if drawable != nil { state = 1 }
                break
            case 1:
                origin = p
                fmt.Println("Origem:", origin)
                state = 2
                break
            case 2:
                point1 = p
                fmt.Println("Ponto 1:", point1)
                state = 3
                break
            case 3:
                point2 = p
                fmt.Println("Ponto 2:", point2)
                state = 4
            }
        case <-kbchan:
            return
        }
        if state == 4 { break }
    }
    Delete(drawable, out)
    drawable.RotatePoints(origin, Angle(origin, point1, point2))
    out <- RegisterPoints(FilterInvalidPoints(drawable.PointChan()), drawable)
}

func Angle (origin image.Point, point1 image.Point, point2 image.Point) float64 {
    radius1 := point1.Sub(origin)
    ang1    := math.Atan(float64(int(radius1.Y))/float64(int(radius1.X)))
    if radius1.X < 0 { ang1 -= math.Pi }
    radius2 := point2.Sub(origin)
    ang2    := math.Atan(float64(int(radius2.Y))/float64(int(radius2.X)))
    if radius2.X < 0 { ang2 -= math.Pi }
    return ang2-ang1
}

func DeleteHandler (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Apagar objeto")
    for {
    select {
        case p := <-clickchan:
            drawable, _ := SearchNearPoint(p)
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
    colorpoints := FilterInvalidPoints(drawable.PointChan())
    for ! closed(colorpoints) {
        point := <-colorpoints
        RemoveFromMatrix(point.point, drawable);
        color_point := TopMatrixColorPoint(point.point)
        blackpoints <- color_point
        //MergeLists(redraw, ListMatrix(point.point))
        //RedrawList(redraw, out)
    }
    close(blackpoints)
}

// func RedrawList(list *list.List, out chan chan ColorPoint) {
//     for elem := list.Front(); elem != nil; elem = elem.Next() {
//         drawable := elem.Value.(ColorPoint).drawable
//         out <- drawable.PointChan()
//     }
// }

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
    line := Line{pa[0], pa[1], currentColor, currentDashStyle, currentThick, Id{0}}
    out <- RegisterPoints(FilterInvalidPoints(line.PointChan()), &line)
    counter_id++
    (&line).SetId(counter_id)
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

func MoveHandler (clickchan <-chan image.Point, kbchan chan int, out chan chan ColorPoint) {
    fmt.Println("Mover objeto")
    has_origin := false
    var drawable Drawable
    var origin image.Point
    for {
    select {
        case p := <-clickchan:
            if has_origin == false {
                drawable, origin = SearchNearPoint(p)
                if drawable != nil {
                    has_origin = true
                }
            } else {
                dest := p
                moviment := dest.Sub(origin)
                fmt.Println("Move (%d, %d)", moviment.X, moviment.Y)
                Delete(drawable, out)
                drawable.Move(moviment)
                out <- RegisterPoints(FilterInvalidPoints(drawable.PointChan()), drawable)
                return
            }
        case <-kbchan:
            return
    }
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
            colorpoint.drawable = &drawable
            PushMatrix(colorpoint)
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

