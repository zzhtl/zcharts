// Package scale 把"数据值"与"像素坐标"建立双向映射。
//
// 直角坐标系下 X/Y 轴各有一个 Scale，渲染 series 时把每个数据点用 Scale 转成画布坐标。
package scale

// Tick 是一根刻度线。
type Tick struct {
	Value float64
	Label string
}

// Scale 是坐标轴数据 ↔ 像素的双向映射接口。
type Scale interface {
	// SetPixelRange 设置该轴的像素区间。
	// 对 X 轴通常 lo < hi（左到右），Y 轴可以传 hi < lo（顶到底，常见做法以确保数值大对应屏幕高）。
	SetPixelRange(lo, hi float64)

	// Pixel 把数据值映射为像素坐标。
	Pixel(v float64) float64

	// Ticks 返回该轴上要绘制的刻度。
	Ticks() []Tick

	// Range 返回数据值的有效范围。
	Range() (lo, hi float64)
}
