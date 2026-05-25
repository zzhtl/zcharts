package series

import (
	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/common/number"
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
	Family   string // 默认字体 family（数据标签用）
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
	if a.Series.ItemStyle.Opacity != nil {
		fillColor = fillColor.WithAlpha(uint8(*a.Series.ItemStyle.Opacity * 255))
	}

	// 可选描边
	stroked := a.Series.ItemStyle.BorderWidth > 0 || a.Series.ItemStyle.BorderColor != ""
	if stroked {
		borderColor := color.RGB(255, 255, 255)
		if a.Series.ItemStyle.BorderColor != "" {
			if c, err := color.Parse(string(a.Series.ItemStyle.BorderColor)); err == nil {
				borderColor = c
			}
		}
		bw := a.Series.ItemStyle.BorderWidth
		if bw <= 0 {
			bw = 1
		}
		a.Canvas.SetStroke(borderColor)
		a.Canvas.SetStrokeWidth(bw)
	}

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
		a.Canvas.SetFill(fillColor)
		if !stroked {
			a.Canvas.NoStroke()
		}
		a.Canvas.DrawCircle(px, py, size/2)

		if a.Series.Label.Show {
			text := number.FormatAuto(yVal)
			if a.Series.Label.Formatter != "" {
				text = formatLabel(a.Series.Label.Formatter, d, yVal, 0)
			}
			drawDataLabel(a.Canvas, a.Series.Label, a.Family, text, px, py-size/2-3, zcanvas.AnchorMiddle, zcanvas.AlignBottom)
		}
	}
}
