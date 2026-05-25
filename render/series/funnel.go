package series

import (
	"math"
	"sort"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/common/number"
	"github.com/zzhtl/zcharts/option"
)

// DrawFunnelArgs Funnel 渲染参数。
type DrawFunnelArgs struct {
	Canvas  zcanvas.Canvas
	Series  *option.FunnelSeries
	Bounds  geom.Rect
	Palette color.Palette
	Family  string
}

// DrawFunnel 绘制漏斗图。
func DrawFunnel(a DrawFunnelArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 {
		return
	}
	rect := funnelRect(a.Bounds, a.Series)
	if rect.IsZero() {
		return
	}

	items := make([]option.DataValue, len(a.Series.Data))
	copy(items, a.Series.Data)
	sortFunnelItems(items, a.Series.Sort)

	minValue, maxValue := funnelValueRange(items, a.Series.Min, a.Series.Max)
	n := len(items)
	gap := a.Series.Gap
	if gap <= 0 {
		gap = 4
	}
	segH := (rect.H - gap*float64(n-1)) / float64(n)
	if segH <= 0 {
		return
	}

	widths := make([]float64, n)
	total := 0.0
	for i, d := range items {
		widths[i] = funnelWidth(rect.W, d.Number(), minValue, maxValue)
		total += math.Max(0, d.Number())
	}

	for i, d := range items {
		topW := widths[i]
		bottomW := topW * 0.72
		if i+1 < n {
			bottomW = widths[i+1]
		}
		y0 := rect.Y + float64(i)*(segH+gap)
		y1 := y0 + segH
		fill := a.Palette.At(i)
		drawFunnelSegment(a.Canvas, rect.CenterX(), y0, y1, topW, bottomW, fill)

		// 百分比（基于各段值占总和的比例）。
		pct := 0.0
		if total > 0 {
			pct = math.Max(0, d.Number()) / total
		}
		var label string
		if a.Series.Label.Formatter != "" {
			label = formatLabel(a.Series.Label.Formatter, d, d.Number(), pct)
		} else {
			// 默认显示：名称:数量(百分比%)（漏斗标签默认显示，与 Word 漏斗一致）。
			valueText := number.FormatAuto(d.Number())
			pctText := number.FormatFloat(pct*100, 1) + "%"
			if d.Name == "" {
				label = valueText + "(" + pctText + ")"
			} else {
				label = d.Name + ":" + valueText + "(" + pctText + ")"
			}
		}
		if label != "" {
			size := a.Series.Label.FontSize
			if size <= 0 {
				size = 13
			}
			customColor, hasCustom := color.Color{}, false
			if a.Series.Label.Color != "" {
				if c, err := color.Parse(string(a.Series.Label.Color)); err == nil {
					customColor, hasCustom = c, true
				}
			}

			midY := (y0 + y1) / 2
			midW := (topW + bottomW) / 2 // 标签所在高度（段中线）处的梯形宽度
			measure := zcanvas.TextStyle{Family: a.Family, Size: size}
			tw, _, _ := a.Canvas.MeasureText(label, measure)

			if tw <= midW-12 {
				// 段内放得下：居中，按底色选对比文字色。
				textColor := contrastText(fill)
				if hasCustom {
					textColor = customColor
				}
				a.Canvas.DrawText(rect.CenterX(), midY, label, zcanvas.TextStyle{
					Family: a.Family, Size: size, Color: textColor,
					Anchor: zcanvas.AnchorMiddle, VAlign: zcanvas.AlignMiddle,
				})
			} else {
				// 段内放不下：移到该段右侧外部并加引导线，保证完整显示。
				edgeX := rect.CenterX() + midW/2
				lx := edgeX + 8
				textColor := color.MustParse("#333")
				if hasCustom {
					textColor = customColor
				}
				a.Canvas.SetStroke(fill)
				a.Canvas.SetStrokeWidth(1)
				a.Canvas.NoFill()
				a.Canvas.DrawLine(edgeX, midY, lx-2, midY)

				anchor := zcanvas.AnchorStart
				if lx+tw > a.Bounds.Right()-4 {
					// 超出画布右缘：改为贴右对齐，避免被截断。
					anchor = zcanvas.AnchorEnd
					lx = a.Bounds.Right() - 4
				}
				a.Canvas.DrawText(lx, midY, label, zcanvas.TextStyle{
					Family: a.Family, Size: size, Color: textColor,
					Anchor: anchor, VAlign: zcanvas.AlignMiddle,
				})
			}
		}
	}
}

func funnelRect(bounds geom.Rect, s *option.FunnelSeries) geom.Rect {
	x := bounds.X + bounds.W*0.14
	y := bounds.Y + bounds.H*0.08
	w := bounds.W * 0.72
	h := bounds.H * 0.78
	if s.Width.Set {
		w = s.Width.Resolve(bounds.W, w)
	}
	if s.Height.Set {
		h = s.Height.Resolve(bounds.H, h)
	}
	if s.Left.Set {
		if s.Left.Keyword == "center" {
			x = bounds.X + (bounds.W-w)/2
		} else {
			x = bounds.X + s.Left.Resolve(bounds.W, x-bounds.X)
		}
	}
	if s.Top.Set {
		if s.Top.Keyword == "middle" || s.Top.Keyword == "center" {
			y = bounds.Y + (bounds.H-h)/2
		} else {
			y = bounds.Y + s.Top.Resolve(bounds.H, y-bounds.Y)
		}
	}
	return geom.Rect{X: x, Y: y, W: w, H: h}
}

func sortFunnelItems(items []option.DataValue, sortMode string) {
	switch sortMode {
	case "ascending":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Number() < items[j].Number() })
	case "none":
		return
	default:
		sort.SliceStable(items, func(i, j int) bool { return items[i].Number() > items[j].Number() })
	}
}

func funnelValueRange(items []option.DataValue, minOverride, maxOverride float64) (float64, float64) {
	minValue, maxValue := math.Inf(1), math.Inf(-1)
	for _, d := range items {
		v := d.Number()
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}
	if minOverride != 0 || maxOverride != 0 {
		minValue, maxValue = minOverride, maxOverride
	}
	if minValue > maxValue {
		minValue, maxValue = 0, 1
	}
	if maxValue <= minValue {
		maxValue = minValue + 1
	}
	return minValue, maxValue
}

func funnelWidth(maxWidth, value, minValue, maxValue float64) float64 {
	ratio := (value - minValue) / (maxValue - minValue)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	minWidth := maxWidth * 0.18
	return minWidth + (maxWidth-minWidth)*ratio
}

func drawFunnelSegment(c zcanvas.Canvas, cx, y0, y1, topW, bottomW float64, fill color.Color) {
	c.SetFill(fill)
	c.SetStroke(color.RGB(255, 255, 255))
	c.SetStrokeWidth(2)
	c.MoveTo(cx-topW/2, y0)
	c.LineTo(cx+topW/2, y0)
	c.LineTo(cx+bottomW/2, y1)
	c.LineTo(cx-bottomW/2, y1)
	c.ClosePath()
	c.FillStroke()
}

func contrastText(fill color.Color) color.Color {
	luma := 0.299*float64(fill.R) + 0.587*float64(fill.G) + 0.114*float64(fill.B)
	if luma > 170 {
		return color.MustParse("#333")
	}
	return color.RGB(255, 255, 255)
}
