package main

import (
	"fmt"
	"math"
)

type Point struct {
	x float64
	y float64
}

func NewPoint(x, y float64) Point {
	return Point{x: x, y: y}
}

func (p Point) Distance(other Point) float64 {
	return math.Sqrt(math.Pow((p.x - other.x), 2.0) + math.Pow((p.y - other.y), 2.0))
}

func main() {
	a := NewPoint(0.0, 0.0)
	b := NewPoint(3.0, 4.0)
	
	dist := a.Distance(b)
	fmt.Println("Point A:", a)
	fmt.Println("Point B:", b)
	fmt.Println("Distance:", dist)
}
