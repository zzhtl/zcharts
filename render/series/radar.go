package series

import (
	"math"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
)

// DrawRadarArgs 雷达图渲染参数。
type DrawRadarArgs struct {
	Canvas  zcanvas.Canvas
	Radar   *option.Radar // 顶层雷达组件配置
	Series  *option.RadarSeries
	Bounds  geom.Rect
	Palette color.Palette
	Family  string

	GridColor color.Color
	AxisColor color.Color
	TextColor color.Color
}

// DrawRadar 绘制雷达图。Radar 组件 + RadarSeries 数据。
func DrawRadar(a DrawRadarArgs) {
	if a.Radar == nil || a.Series == nil {
		return
	}
	indicators := a.Radar.Indicator
	if len(indicators) < 3 {
		return
	}
	cx, cy := pieCenter(a.Bounds, a.Radar.Center)
	ref := math.Min(a.Bounds.W, a.Bounds.H) / 2
	maxR := ref * 0.7
	if a.Radar.Radius.Set {
		if a.Radar.Radius.IsPercent() {
			maxR = ref * a.Radar.Radius.Value / 100
		} else {
			maxR = a.Radar.Radius.Value
		}
	}

	n := len(indicators)
	// 起始角度（默认 -90° 即 12 点方向），顺时针
	startAngle := -math.Pi / 2
	if a.Radar.StartAngle != 0 {
		startAngle = a.Radar.StartAngle * math.Pi / 180
	}
	step := 2 * math.Pi / float64(n)

	// 网格层数
	layers := a.Radar.SplitNumber
	if layers <= 0 {
		layers = 5
	}

	// 1. 同心多边形 + 放射线
	a.Canvas.SetStroke(a.GridColor)
	a.Canvas.SetStrokeWidth(1)
	a.Canvas.NoFill()
	for li := 1; li <= layers; li++ {
		r := maxR * float64(li) / float64(layers)
		for i := 0; i < n; i++ {
			ang := startAngle + step*float64(i)
			x := cx + math.Cos(ang)*r
			y := cy + math.Sin(ang)*r
			if i == 0 {
				a.Canvas.MoveTo(x, y)
			} else {
				a.Canvas.LineTo(x, y)
			}
		}
		a.Canvas.ClosePath()
		a.Canvas.Stroke()
	}
	// 放射线
	a.Canvas.SetStroke(a.AxisColor)
	for i := 0; i < n; i++ {
		ang := startAngle + step*float64(i)
		x := cx + math.Cos(ang)*maxR
		y := cy + math.Sin(ang)*maxR
		a.Canvas.DrawLine(cx, cy, x, y)
	}

	// 2. 指标名称
	textStyle := zcanvas.TextStyle{
		Family: a.Family,
		Size:   12,
		Color:  a.TextColor,
		VAlign: zcanvas.AlignMiddle,
	}
	for i, ind := range indicators {
		ang := startAngle + step*float64(i)
		x := cx + math.Cos(ang)*(maxR+14)
		y := cy + math.Sin(ang)*(maxR+14)
		style := textStyle
		style.Anchor = anchorForAngle(ang * 180 / math.Pi)
		a.Canvas.DrawText(x, y, ind.Name, style)
	}

	// 3. 数据多边形（一个 RadarSeries 可能有多个 data item，每个对应一条曲线）
	for i, item := range a.Series.Data {
		fillColor := a.Palette.At(i)
		strokeColor := fillColor
		areaAlpha := uint8(80)
		if item.ItemStyle.Color != "" {
			if c, err := color.Parse(string(item.ItemStyle.Color)); err == nil {
				fillColor = c
				strokeColor = c
			}
		}

		// 计算每个顶点
		pts := make([]geom.Point, n)
		for k, ind := range indicators {
			max := ind.Max
			if max <= 0 {
				max = 1
			}
			val := 0.0
			if k < len(item.Value) {
				val = item.Value[k]
			}
			r := maxR * (val / max)
			if r < 0 {
				r = 0
			} else if r > maxR {
				r = maxR
			}
			ang := startAngle + step*float64(k)
			pts[k] = geom.Point{X: cx + math.Cos(ang)*r, Y: cy + math.Sin(ang)*r}
		}

		// 填充
		a.Canvas.SetFill(fillColor.WithAlpha(areaAlpha))
		a.Canvas.NoStroke()
		a.Canvas.MoveTo(pts[0].X, pts[0].Y)
		for j := 1; j < n; j++ {
			a.Canvas.LineTo(pts[j].X, pts[j].Y)
		}
		a.Canvas.ClosePath()
		a.Canvas.Fill()

		// 描边
		a.Canvas.SetStroke(strokeColor)
		a.Canvas.SetStrokeWidth(2)
		a.Canvas.NoFill()
		a.Canvas.MoveTo(pts[0].X, pts[0].Y)
		for j := 1; j < n; j++ {
			a.Canvas.LineTo(pts[j].X, pts[j].Y)
		}
		a.Canvas.ClosePath()
		a.Canvas.Stroke()

		// 顶点
		for _, p := range pts {
			a.Canvas.SetFill(strokeColor)
			a.Canvas.NoStroke()
			a.Canvas.DrawCircle(p.X, p.Y, 3)
		}
	}
}
