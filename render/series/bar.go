package series

import (
	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/scale"
)

// DrawBarArgs Bar 渲染参数。
type DrawBarArgs struct {
	Canvas         zcanvas.Canvas
	Series         *option.BarSeries
	XScale         scale.Scale
	YScale         scale.Scale
	GridRect       geom.Rect
	Color          color.Color
	SeriesIndex    int     // 该 Bar 在所有同坐标 Bar series 中的索引（用于并排）
	TotalBarSeries int     // 同坐标 Bar series 总数
	BandWidth      float64 // 单 category 占据的像素宽度
	StackBases     []float64
}

// DrawBar 绘制一个 Bar series。多 series 时按 SeriesIndex/TotalBarSeries 在每个 category 内并排。
func DrawBar(a DrawBarArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 || a.XScale == nil || a.YScale == nil {
		return
	}
	n := a.TotalBarSeries
	if n <= 0 {
		n = 1
	}
	// 类目内 80% 给所有柱子使用，左右各留 10% gap
	groupWidth := a.BandWidth * 0.8
	barWidth := groupWidth / float64(n)
	if barWidth < 1 {
		barWidth = 1
	}

	borderColor := a.Color
	if a.Series.ItemStyle.BorderColor != "" {
		if c, err := color.Parse(string(a.Series.ItemStyle.BorderColor)); err == nil {
			borderColor = c
		}
	}
	borderWidth := a.Series.ItemStyle.BorderWidth

	fillColor := a.Color
	if a.Series.ItemStyle.Color != "" {
		if c, err := color.Parse(string(a.Series.ItemStyle.Color)); err == nil {
			fillColor = c
		}
	}

	for i, d := range a.Series.Data {
		cx := a.XScale.Pixel(float64(i))
		groupStart := cx - groupWidth/2
		x := groupStart + float64(a.SeriesIndex)*barWidth
		base := 0.0
		if i < len(a.StackBases) {
			base = a.StackBases[i]
		}
		baseY := a.YScale.Pixel(base)
		valueY := a.YScale.Pixel(base + d.Number())
		var top, height float64
		if valueY <= baseY {
			top = valueY
			height = baseY - valueY
		} else {
			top = baseY
			height = valueY - baseY
		}
		a.Canvas.SetFill(fillColor)
		if borderWidth > 0 {
			a.Canvas.SetStroke(borderColor)
			a.Canvas.SetStrokeWidth(borderWidth)
		} else {
			a.Canvas.NoStroke()
		}
		a.Canvas.DrawRect(x, top, barWidth-1, height) // 留 1px 给视觉间隙
	}
}
