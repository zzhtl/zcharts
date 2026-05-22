package series

import (
	"math"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/common/number"
	"github.com/zzhtl/zcharts/option"
)

// DrawGaugeArgs 仪表盘渲染参数。
type DrawGaugeArgs struct {
	Canvas  zcanvas.Canvas
	Series  *option.GaugeSeries
	Bounds  geom.Rect
	Palette color.Palette
	Family  string

	TrackColor color.Color // 背景轨道色
	TextColor  color.Color
}

// DrawGauge 绘制仪表盘（仅取 data 第 0 个值）。
func DrawGauge(a DrawGaugeArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 {
		return
	}
	min := a.Series.Min
	max := a.Series.Max
	if max <= min {
		max = 100
	}
	startAngle := 225.0
	endAngle := -45.0
	if a.Series.StartAngle != nil {
		startAngle = *a.Series.StartAngle
	}
	if a.Series.EndAngle != nil {
		endAngle = *a.Series.EndAngle
	}

	cx, cy := pieCenter(a.Bounds, a.Series.Center)
	ref := math.Min(a.Bounds.W, a.Bounds.H) / 2
	maxR := ref * 0.7
	if a.Series.Radius.Set {
		if a.Series.Radius.IsPercent() {
			maxR = ref * a.Series.Radius.Value / 100
		} else {
			maxR = a.Series.Radius.Value
		}
	}
	thickness := maxR * 0.12

	val := a.Series.Data[0].Number()
	frac := (val - min) / (max - min)
	if frac < 0 {
		frac = 0
	} else if frac > 1 {
		frac = 1
	}
	// ECharts 仪表盘的角度通常表示数学坐标系（CCW、X 轴正向 0°）
	// 我们的 DrawSector 在屏幕坐标系下顺时针（与 SVG 一致），把 ECharts 角度转换：
	// screenAngle = -echartsAngle（X 轴方向不变，Y 翻转）
	sStart := -startAngle
	sEnd := -endAngle
	if sEnd < sStart {
		sEnd, sStart = sStart, sEnd
	}
	curEnd := sStart + (sEnd-sStart)*frac

	// 轨道
	a.Canvas.SetFill(a.TrackColor)
	a.Canvas.NoStroke()
	a.Canvas.DrawSector(cx, cy, maxR-thickness, maxR, sStart, sEnd)

	// 当前值弧
	activeColor := a.Palette.At(0)
	if a.Series.ItemStyle.Color != "" {
		if c, err := color.Parse(string(a.Series.ItemStyle.Color)); err == nil {
			activeColor = c
		}
	}
	a.Canvas.SetFill(activeColor)
	a.Canvas.DrawSector(cx, cy, maxR-thickness, maxR, sStart, curEnd)

	// 指针：从中心指向当前弧端点
	pointerAng := curEnd * math.Pi / 180
	px := cx + math.Cos(pointerAng)*(maxR-thickness-6)
	py := cy + math.Sin(pointerAng)*(maxR-thickness-6)
	a.Canvas.SetStroke(activeColor)
	a.Canvas.SetStrokeWidth(3)
	a.Canvas.SetLineCap(zcanvas.CapRound)
	a.Canvas.DrawLine(cx, cy, px, py)

	// 中心圆点
	a.Canvas.SetFill(activeColor)
	a.Canvas.NoStroke()
	a.Canvas.DrawCircle(cx, cy, 6)

	// 数字
	label := number.FormatAuto(val)
	if a.Series.Label.Formatter != "" {
		label = formatLabel(a.Series.Label.Formatter, a.Series.Data[0], val, frac)
	}
	a.Canvas.DrawText(cx, cy+thickness*2, label, zcanvas.TextStyle{
		Family: a.Family,
		Size:   24,
		Weight: "bold",
		Color:  a.TextColor,
		Anchor: zcanvas.AnchorMiddle,
		VAlign: zcanvas.AlignTop,
	})

	// 名称（如果有）
	if name := a.Series.Data[0].Name; name != "" {
		a.Canvas.DrawText(cx, cy+thickness*2+30, name, zcanvas.TextStyle{
			Family: a.Family,
			Size:   12,
			Color:  a.TextColor,
			Anchor: zcanvas.AnchorMiddle,
			VAlign: zcanvas.AlignTop,
		})
	}
	// 抑制 geom 仅在文件中没被使用时的 import 错误（保留以备扩展）
	_ = geom.Point{}
}
