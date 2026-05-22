package series

import (
	"math"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/scale"
)

// DrawHeatmapArgs Heatmap 渲染参数。
type DrawHeatmapArgs struct {
	Canvas    zcanvas.Canvas
	Series    *option.HeatmapSeries
	XScale    scale.Scale
	YScale    scale.Scale
	GridRect  geom.Rect
	VisualMap *option.VisualMap // 必需：决定颜色映射
	Theme     color.Palette     // visualMap 为空时的兜底配色
}

// DrawHeatmap 绘制热力图。
// 数据格式：每个 DataValue.Values 至少包含 [xIndex, yIndex, value]。
func DrawHeatmap(a DrawHeatmapArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 || a.XScale == nil || a.YScale == nil {
		return
	}
	// 取 X/Y category band 宽度
	xCat, xOK := a.XScale.(*scale.Category)
	yCat, yOK := a.YScale.(*scale.Category)
	if !xOK || !yOK {
		return
	}
	xBand := xCat.BandWidth()
	yBand := math.Abs(yCat.BandWidth())

	// 确定值的范围
	var lo, hi float64
	if a.VisualMap != nil && a.VisualMap.Max > a.VisualMap.Min {
		lo, hi = a.VisualMap.Min, a.VisualMap.Max
	} else {
		lo, hi = math.Inf(1), math.Inf(-1)
		for _, d := range a.Series.Data {
			v := dataValue(d)
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
		}
		if lo > hi {
			lo, hi = 0, 1
		}
	}

	stops := defaultStops()
	if a.VisualMap != nil && len(a.VisualMap.InRange.Color) > 0 {
		stops = make([]color.Color, 0, len(a.VisualMap.InRange.Color))
		for _, hex := range a.VisualMap.InRange.Color {
			if c, err := color.Parse(hex); err == nil {
				stops = append(stops, c)
			}
		}
		if len(stops) == 0 {
			stops = defaultStops()
		}
	}

	a.Canvas.NoStroke()
	for _, d := range a.Series.Data {
		if len(d.Values) < 3 {
			continue
		}
		ix, iy, v := d.Values[0], d.Values[1], d.Values[2]
		px := a.XScale.Pixel(ix)
		py := a.YScale.Pixel(iy)
		t := 0.0
		if hi > lo {
			t = (v - lo) / (hi - lo)
		}
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		c := interpolateStops(stops, t)
		a.Canvas.SetFill(c)
		a.Canvas.DrawRect(px-xBand/2+1, py-yBand/2+1, xBand-2, yBand-2)
	}
}

func dataValue(d option.DataValue) float64 {
	if len(d.Values) >= 3 {
		return d.Values[2]
	}
	return d.Number()
}

func defaultStops() []color.Color {
	return []color.Color{
		color.MustParse("#313695"),
		color.MustParse("#4575b4"),
		color.MustParse("#74add1"),
		color.MustParse("#abd9e9"),
		color.MustParse("#e0f3f8"),
		color.MustParse("#ffffbf"),
		color.MustParse("#fee090"),
		color.MustParse("#fdae61"),
		color.MustParse("#f46d43"),
		color.MustParse("#d73027"),
		color.MustParse("#a50026"),
	}
}

// interpolateStops 在颜色 stop 数组中按 t∈[0,1] 线性插值。
func interpolateStops(stops []color.Color, t float64) color.Color {
	if len(stops) == 0 {
		return color.Color{}
	}
	if len(stops) == 1 || t <= 0 {
		return stops[0]
	}
	if t >= 1 {
		return stops[len(stops)-1]
	}
	pos := t * float64(len(stops)-1)
	i := int(math.Floor(pos))
	frac := pos - float64(i)
	a, b := stops[i], stops[i+1]
	return color.Color{
		R: lerpByte(a.R, b.R, frac),
		G: lerpByte(a.G, b.G, frac),
		B: lerpByte(a.B, b.B, frac),
		A: lerpByte(a.A, b.A, frac),
	}
}

func lerpByte(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*t + 0.5)
}
