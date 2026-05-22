package series

import (
	"math"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/common/number"
	"github.com/zzhtl/zcharts/option"
)

// DrawPieArgs Pie 渲染参数。
type DrawPieArgs struct {
	Canvas  zcanvas.Canvas
	Series  *option.PieSeries
	Bounds  geom.Rect    // 整张画布
	Palette color.Palette // 配色（series.itemStyle.color 覆盖之后兜底）
	Family  string
}

// DrawPie 绘制饼图/环图/玫瑰图。
func DrawPie(a DrawPieArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 {
		return
	}
	cx, cy := pieCenter(a.Bounds, a.Series.Center)
	rIn, rOut := pieRadius(a.Bounds, a.Series.Radius)

	// 求总值
	total := 0.0
	for _, d := range a.Series.Data {
		total += math.Max(0, d.Number())
	}
	if total <= 0 {
		return
	}

	// 玫瑰图：按数据值缩放半径
	maxV := 0.0
	if a.Series.RoseType != "" {
		for _, d := range a.Series.Data {
			if v := d.Number(); v > maxV {
				maxV = v
			}
		}
	}

	startAngle := -90.0 // 从顶部开始（12 点方向）
	for i, d := range a.Series.Data {
		v := math.Max(0, d.Number())
		angle := v / total * 360
		// 玫瑰图：area 模式用 sqrt 缩放半径，radius 模式直接缩放
		segOut := rOut
		if a.Series.RoseType != "" && maxV > 0 {
			scale := v / maxV
			if a.Series.RoseType == "area" {
				scale = math.Sqrt(scale)
			}
			segOut = rIn + (rOut-rIn)*scale + (rOut-rIn)*0.1
			if segOut > rOut {
				segOut = rOut
			}
		}

		// 颜色
		c := a.Palette.At(i)
		if d.Raw != nil {
			// 数据项可指定 itemStyle.color，但我们已经在 Pie 用 Palette.At 兜底，
			// 这里如果需要进一步细化可解析 d.Raw 中的 itemStyle，当前阶段省略
		}
		a.Canvas.SetFill(c)
		a.Canvas.SetStroke(color.RGB(255, 255, 255))
		a.Canvas.SetStrokeWidth(2)
		a.Canvas.DrawSector(cx, cy, rIn, segOut, startAngle, startAngle+angle)

		// label（默认 outside）
		if a.Series.Label.Show {
			midAngle := startAngle + angle/2
			lx := cx + math.Cos(midAngle*math.Pi/180)*(segOut+12)
			ly := cy + math.Sin(midAngle*math.Pi/180)*(segOut+12)
			label := d.Name
			if a.Series.Label.Formatter != "" {
				label = formatLabel(a.Series.Label.Formatter, d, v, v/total)
			} else if label == "" {
				label = number.FormatAuto(v)
			}
			labelColor := color.MustParse("#333")
			if a.Series.Label.Color != "" {
				if cc, err := color.Parse(string(a.Series.Label.Color)); err == nil {
					labelColor = cc
				}
			}
			size := a.Series.Label.FontSize
			if size <= 0 {
				size = 12
			}
			a.Canvas.DrawText(lx, ly, label, zcanvas.TextStyle{
				Family: a.Family,
				Size:   size,
				Color:  labelColor,
				Anchor: anchorForAngle(midAngle),
				VAlign: zcanvas.AlignMiddle,
			})
		}

		startAngle += angle
	}
}

func pieCenter(b geom.Rect, center option.FlexList) (float64, float64) {
	cx := b.X + b.W/2
	cy := b.Y + b.H/2
	if len(center) >= 2 {
		if center[0].Set {
			cx = b.X + center[0].Resolve(b.W, b.W/2)
		}
		if center[1].Set {
			cy = b.Y + center[1].Resolve(b.H, b.H/2)
		}
	}
	return cx, cy
}

func pieRadius(b geom.Rect, radius option.FlexList) (float64, float64) {
	ref := math.Min(b.W, b.H) / 2
	maxR := ref * 0.85
	rIn, rOut := 0.0, maxR
	switch len(radius) {
	case 1:
		if radius[0].IsPercent() {
			rOut = ref * radius[0].Value / 100
		} else {
			rOut = radius[0].Value
		}
	case 2:
		if radius[0].IsPercent() {
			rIn = ref * radius[0].Value / 100
		} else {
			rIn = radius[0].Value
		}
		if radius[1].IsPercent() {
			rOut = ref * radius[1].Value / 100
		} else {
			rOut = radius[1].Value
		}
	}
	return rIn, rOut
}

func anchorForAngle(angleDeg float64) zcanvas.TextAnchor {
	// 屏幕坐标系，0° 右，90° 下，180° 左，270° 上
	a := math.Mod(angleDeg+360, 360)
	if a > 90 && a < 270 {
		return zcanvas.AnchorEnd
	}
	return zcanvas.AnchorStart
}

// formatLabel 实现 ECharts 风格的 label formatter，支持：
//	{a} → series name
//	{b} → data name
//	{c} → data value
//	{d} → 百分比（不带 %）
func formatLabel(format string, d option.DataValue, value, percent float64) string {
	r := format
	r = replaceAll(r, "{b}", d.Name)
	r = replaceAll(r, "{c}", number.FormatAuto(value))
	r = replaceAll(r, "{d}", number.FormatFloat(percent*100, 2))
	return r
}

func replaceAll(s, from, to string) string {
	if from == "" {
		return s
	}
	out := s
	for {
		idx := indexOf(out, from)
		if idx < 0 {
			return out
		}
		out = out[:idx] + to + out[idx+len(from):]
	}
}

func indexOf(s, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
