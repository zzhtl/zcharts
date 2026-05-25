package series

import (
	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/common/number"
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
	Family         string  // 默认字体 family（数据标签用）
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

	stacked := a.Series.Stack != ""
	// 圆角：显式配置优先；否则非堆积柱默认给一个轻微圆角让观感更柔和。
	radius := a.Series.ItemStyle.BorderRadius
	if a.Series.ItemStyle.BorderRadius == 0 && !stacked {
		radius = 3
	}
	// 渐变：顶部略亮、底部本色，营造轻微立体感。
	gradTop := lighten(fillColor, 0.18)

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
		upward := valueY <= baseY
		if upward {
			top = valueY
			height = baseY - valueY
		} else {
			top = baseY
			height = valueY - baseY
		}
		w := barWidth - 1 // 留 1px 给视觉间隙
		if w < 1 {
			w = barWidth
		}

		// 填充：渐变（沿柱高方向）。
		a.Canvas.SetFillLinearGradient(x, top, x, top+height, []zcanvas.GradientStop{
			{Offset: 0, Color: gradTop},
			{Offset: 1, Color: fillColor},
		})
		if borderWidth > 0 {
			a.Canvas.SetStroke(borderColor)
			a.Canvas.SetStrokeWidth(borderWidth)
		} else {
			a.Canvas.NoStroke()
		}

		r := radius
		if r > w/2 {
			r = w / 2
		}
		if r > height {
			r = height
		}
		if r > 0 {
			drawTopRoundedBar(a.Canvas, x, top, w, height, r, upward, borderWidth > 0)
		} else {
			a.Canvas.DrawRect(x, top, w, height)
		}

		// 数据标签
		if a.Series.Label.Show {
			text := number.FormatAuto(d.Number())
			if a.Series.Label.Formatter != "" {
				text = formatLabel(a.Series.Label.Formatter, d, d.Number(), 0)
			}
			ly := top - 4
			valign := zcanvas.AlignBottom
			if !upward {
				ly = top + height + 4
				valign = zcanvas.AlignTop
			}
			drawDataLabel(a.Canvas, a.Series.Label, a.Family, text, x+w/2, ly, zcanvas.AnchorMiddle, valign)
		}
	}
}

// drawTopRoundedBar 绘制顶部两角为 r 圆角的柱子（底部直角，贴合基线）。
// upward=false（向下的负值柱）时改为底部两角圆角。fillStroke 控制是否同时描边。
func drawTopRoundedBar(c zcanvas.Canvas, x, top, w, h, r float64, upward, stroked bool) {
	c.BeginPath()
	if upward {
		c.MoveTo(x, top+h)
		c.LineTo(x, top+r)
		c.QuadTo(x, top, x+r, top)
		c.LineTo(x+w-r, top)
		c.QuadTo(x+w, top, x+w, top+r)
		c.LineTo(x+w, top+h)
	} else {
		c.MoveTo(x, top)
		c.LineTo(x+w, top)
		c.LineTo(x+w, top+h-r)
		c.QuadTo(x+w, top+h, x+w-r, top+h)
		c.LineTo(x+r, top+h)
		c.QuadTo(x, top+h, x, top+h-r)
	}
	c.ClosePath()
	if stroked {
		c.FillStroke()
	} else {
		c.Fill()
	}
}

// lighten 把颜色按 t∈[0,1] 向白色插值。
func lighten(c color.Color, t float64) color.Color {
	if t <= 0 {
		return c
	}
	if t > 1 {
		t = 1
	}
	return color.Color{
		R: uint8(float64(c.R) + (255-float64(c.R))*t),
		G: uint8(float64(c.G) + (255-float64(c.G))*t),
		B: uint8(float64(c.B) + (255-float64(c.B))*t),
		A: c.A,
	}
}
