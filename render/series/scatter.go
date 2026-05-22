package series

import (
	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/scale"
)

// DrawScatterArgs Scatter 渲染参数。
type DrawScatterArgs struct {
	Canvas   zcanvas.Canvas
	Series   *option.ScatterSeries
	XScale   scale.Scale
	YScale   scale.Scale
	GridRect geom.Rect
	Color    color.Color
}

// DrawScatter 绘制散点图。
// 当 XScale 是 category 时，按"数据索引"作为 X 值；否则按数据点的第一维数值。
func DrawScatter(a DrawScatterArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 || a.XScale == nil || a.YScale == nil {
		return
	}
	size := a.Series.SymbolSize
	if size <= 0 {
		size = 10
	}
	fillColor := a.Color
	if a.Series.ItemStyle.Color != "" {
		if c, err := color.Parse(string(a.Series.ItemStyle.Color)); err == nil {
			fillColor = c
		}
	}
	a.Canvas.SetFill(fillColor)
	a.Canvas.NoStroke()
	_, isCat := a.XScale.(*scale.Category)
	for i, d := range a.Series.Data {
		var xVal float64
		if isCat {
			xVal = float64(i)
		} else {
			xVal, _ = d.Pair()
		}
		_, yVal := d.Pair()
		if !isCat && len(d.Values) == 1 {
			// 单维数据：当作 y，x 取索引
			xVal = float64(i)
			yVal = d.Values[0]
		}
		px := a.XScale.Pixel(xVal)
		py := a.YScale.Pixel(yVal)
		a.Canvas.DrawCircle(px, py, size/2)
	}
}
