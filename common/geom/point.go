// Package geom 提供 zcharts 内部使用的轻量几何类型。
package geom

import "math"

// Point 表示二维平面上的一个点。坐标单位约定与上层（通常是像素）一致。
type Point struct {
	X, Y float64
}

// Add 返回两点逐分量相加的结果。
func (p Point) Add(q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }

// Sub 返回 p-q。
func (p Point) Sub(q Point) Point { return Point{p.X - q.X, p.Y - q.Y} }

// Scale 等比缩放。
func (p Point) Scale(s float64) Point { return Point{p.X * s, p.Y * s} }

// Distance 返回两点距离。
func (p Point) Distance(q Point) float64 {
	dx, dy := p.X-q.X, p.Y-q.Y
	return math.Sqrt(dx*dx + dy*dy)
}
