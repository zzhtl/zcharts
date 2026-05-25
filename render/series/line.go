// Package series 各 series 类型的渲染器实现。
//
// 每个 series 渲染函数都是无状态的，由 render 包的 engine 在合适时机调用。
package series

import (
	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/common/number"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/scale"
)

// DrawLineArgs Line 渲染参数集合。
type DrawLineArgs struct {
	Canvas   zcanvas.Canvas
	Series   *option.LineSeries
	XScale   scale.Scale
	YScale   scale.Scale
	GridRect geom.Rect
	Color    color.Color // 主色（来自调色板或 series.color）
	BaseY    float64     // areaStyle 填充时的基准 y（一般是 grid 底边）
	Family   string      // 默认字体 family
	UseDataX bool        // true 时 data 数组的前两个数值分别作为 x/y
}

// DrawLine 绘制一条折线 series（包含可选的面积填充和数据点 symbol）。
func DrawLine(a DrawLineArgs) {
	c := a.Canvas
	data := a.Series.Data
	if len(data) == 0 || a.XScale == nil || a.YScale == nil {
		return
	}

	// 计算每个点的像素位置
	points := make([]geom.Point, 0, len(data))
	for i, d := range data {
		xVal, yVal := linePointValue(d, i, a.UseDataX)
		x := a.XScale.Pixel(xVal)
		y := a.YScale.Pixel(yVal)
		points = append(points, geom.Point{X: x, Y: y})
	}

	// 面积填充（默认用从线到基线的竖直渐变，更有层次）
	if a.Series.AreaStyle != nil {
		areaColor := a.Color.WithAlpha(150)
		if a.Series.AreaStyle.Color != "" {
			if cc, err := color.Parse(string(a.Series.AreaStyle.Color)); err == nil {
				areaColor = cc
				if a.Series.AreaStyle.Opacity != nil {
					areaColor.A = uint8(*a.Series.AreaStyle.Opacity * 255)
				}
			}
		}
		topY := points[0].Y
		for _, p := range points {
			if p.Y < topY {
				topY = p.Y
			}
		}
		c.SetFillLinearGradient(0, topY, 0, a.BaseY, []zcanvas.GradientStop{
			{Offset: 0, Color: areaColor},
			{Offset: 1, Color: areaColor.WithAlpha(areaColor.A / 6)},
		})
		c.NoStroke()
		c.MoveTo(points[0].X, a.BaseY)
		c.LineTo(points[0].X, points[0].Y)
		if a.Series.Smooth {
			drawSmoothPath(c, points)
		} else {
			for _, p := range points[1:] {
				c.LineTo(p.X, p.Y)
			}
		}
		c.LineTo(points[len(points)-1].X, a.BaseY)
		c.ClosePath()
		c.Fill()
	}

	// 折线
	lineColor := a.Color
	if a.Series.LineStyle.Color != "" {
		if cc, err := color.Parse(string(a.Series.LineStyle.Color)); err == nil {
			lineColor = cc
		}
	}
	lineWidth := a.Series.LineStyle.Width
	if lineWidth <= 0 {
		lineWidth = 2
	}
	c.SetStroke(lineColor)
	c.SetStrokeWidth(lineWidth)
	c.SetLineCap(zcanvas.CapRound)
	// 线型：dashed / dotted（虚线/点线），其余为实线。
	switch a.Series.LineStyle.Type {
	case "dashed":
		c.SetDash(0, []float64{lineWidth * 4, lineWidth * 3})
	case "dotted":
		c.SetDash(0, []float64{lineWidth, lineWidth * 2})
	}
	c.NoFill()
	c.MoveTo(points[0].X, points[0].Y)
	if a.Series.Smooth {
		drawSmoothPath(c, points)
	} else {
		for _, p := range points[1:] {
			c.LineTo(p.X, p.Y)
		}
	}
	c.Stroke()
	c.SetDash(0, nil) // 复位虚线，避免影响后续绘制

	// 数据点 symbol
	showSymbol := true
	if a.Series.ShowSymbol != nil {
		showSymbol = *a.Series.ShowSymbol
	}
	if showSymbol && a.Series.Symbol != "none" {
		size := a.Series.SymbolSize
		if size <= 0 {
			size = 6
		}
		for _, p := range points {
			c.SetFill(color.RGB(255, 255, 255))
			c.SetStroke(lineColor)
			c.SetStrokeWidth(2)
			c.DrawCircle(p.X, p.Y, size/2)
		}
	}

	// 数据标签
	if a.Series.Label.Show {
		for i, p := range points {
			_, yVal := linePointValue(data[i], i, a.UseDataX)
			text := number.FormatAuto(yVal)
			if a.Series.Label.Formatter != "" {
				text = formatLabel(a.Series.Label.Formatter, data[i], yVal, 0)
			}
			drawDataLabel(c, a.Series.Label, a.Family, text, p.X, p.Y-8, zcanvas.AnchorMiddle, zcanvas.AlignBottom)
		}
	}
}

func linePointValue(d option.DataValue, index int, useDataX bool) (float64, float64) {
	if useDataX && len(d.Values) >= 2 {
		return d.Values[0], d.Values[1]
	}
	return float64(index), d.Number()
}

// drawSmoothPath 用 Catmull-Rom → Cubic Bezier 转换，绘制平滑曲线（pts[0] 已 MoveTo）。
func drawSmoothPath(c zcanvas.Canvas, pts []geom.Point) {
	n := len(pts)
	if n < 2 {
		return
	}
	if n == 2 {
		c.LineTo(pts[1].X, pts[1].Y)
		return
	}
	for i := 0; i < n-1; i++ {
		p0 := pts[max(i-1, 0)]
		p1 := pts[i]
		p2 := pts[i+1]
		p3 := pts[min(i+2, n-1)]

		const tension = 0.2
		c1x := p1.X + (p2.X-p0.X)*tension
		c1y := p1.Y + (p2.Y-p0.Y)*tension
		c2x := p2.X - (p3.X-p1.X)*tension
		c2y := p2.Y - (p3.Y-p1.Y)*tension
		c.CubicTo(c1x, c1y, c2x, c2y, p2.X, p2.Y)
	}
}
