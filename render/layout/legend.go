package layout

import (
	"strings"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/theme"
)

// LegendItem 单个图例条目。
type LegendItem struct {
	Name  string
	Color color.Color
}

// DrawLegend 在画布顶部居下方位置水平绘制图例。
// 第一阶段只支持顶部居中/左/右、水平/竖直，简化的布局。
func DrawLegend(c zcanvas.Canvas, bounds geom.Rect, items []LegendItem, opt *option.Legend, th *theme.Theme) {
	if len(items) == 0 || opt != nil && !opt.IsShown() {
		return
	}
	textStyle := zcanvas.TextStyle{
		Family: pickString(th.Legend.Text.FontFamily, th.TextStyle.FontFamily),
		Size:   th.Legend.Text.FontSize,
		Color:  th.Legend.Text.Color,
		Weight: th.Legend.Text.FontWeight,
		VAlign: zcanvas.AlignMiddle,
		Anchor: zcanvas.AnchorStart,
	}
	if textStyle.Family == "" {
		textStyle.Family = th.TextStyle.FontFamily
	}
	if textStyle.Size <= 0 {
		textStyle.Size = th.TextStyle.FontSize
	}
	if opt != nil {
		if opt.TextStyle.FontFamily != "" {
			textStyle.Family = opt.TextStyle.FontFamily
		}
		if opt.TextStyle.FontSize > 0 {
			textStyle.Size = opt.TextStyle.FontSize
		}
		if opt.TextStyle.Color != "" {
			if cc, err := color.Parse(string(opt.TextStyle.Color)); err == nil {
				textStyle.Color = cc
			}
		}
		if opt.TextStyle.FontWeight != "" {
			textStyle.Weight = opt.TextStyle.FontWeight
		}
	}

	iconW, iconH := 14.0, 10.0
	if opt != nil {
		if opt.ItemWidth > 0 {
			iconW = opt.ItemWidth
		}
		if opt.ItemHeight > 0 {
			iconH = opt.ItemHeight
		}
	}

	gap := th.Legend.ItemGap
	if gap <= 0 {
		gap = 10
	}
	if opt != nil && opt.ItemGap > 0 {
		gap = opt.ItemGap
	}

	measures := make([]legendMeasure, len(items))
	for i, it := range items {
		tw, th, ascent := c.MeasureText(it.Name, textStyle)
		measures[i] = legendMeasure{width: iconW + 4 + tw, height: maxFloat(iconH, th), textHeight: th, textAscent: ascent}
	}

	orient := "horizontal"
	if opt != nil && strings.EqualFold(opt.Orient, "vertical") {
		orient = "vertical"
	}
	if orient == "vertical" {
		drawVerticalLegend(c, bounds, items, measures, opt, textStyle, iconW, iconH, gap)
		return
	}
	drawHorizontalLegend(c, bounds, items, measures, opt, textStyle, iconW, iconH, gap)
}

type legendMeasure struct {
	width      float64
	height     float64
	textHeight float64
	textAscent float64
}

func drawHorizontalLegend(c zcanvas.Canvas, bounds geom.Rect, items []LegendItem, measures []legendMeasure, opt *option.Legend, textStyle zcanvas.TextStyle, iconW, iconH, gap float64) {
	maxW := bounds.W - 20
	if maxW < 20 {
		maxW = bounds.W
	}

	rows := make([][]int, 0, 1)
	row := make([]int, 0, len(items))
	rowW := 0.0
	for i := range items {
		nextW := measures[i].width
		if len(row) > 0 {
			nextW += gap
		}
		if len(row) > 0 && rowW+nextW > maxW {
			rows = append(rows, row)
			row = []int{i}
			rowW = measures[i].width
			continue
		}
		row = append(row, i)
		rowW += nextW
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	rowGap := 8.0
	rowHeights := make([]float64, len(rows))
	totalH := 0.0
	for i, r := range rows {
		for _, idx := range r {
			rowHeights[i] = maxFloat(rowHeights[i], measures[idx].height)
		}
		totalH += rowHeights[i]
	}
	totalH += rowGap * float64(len(rows)-1)

	y := defaultLegendY(bounds, opt, totalH)
	for i, r := range rows {
		rowWidth := legendRowWidth(r, measures, gap)
		x := legendX(bounds, opt, rowWidth)
		for _, idx := range r {
			drawLegendItem(c, x, y+rowHeights[i]/2, items[idx], measures[idx], textStyle, iconW, iconH)
			x += measures[idx].width + gap
		}
		y += rowHeights[i] + rowGap
	}
}

func drawVerticalLegend(c zcanvas.Canvas, bounds geom.Rect, items []LegendItem, measures []legendMeasure, opt *option.Legend, textStyle zcanvas.TextStyle, iconW, iconH, gap float64) {
	totalW := 0.0
	totalH := 0.0
	for _, m := range measures {
		totalW = maxFloat(totalW, m.width)
		totalH += m.height
	}
	totalH += gap * float64(len(items)-1)

	x := legendX(bounds, opt, totalW)
	y := defaultLegendY(bounds, opt, totalH)
	for i, it := range items {
		drawLegendItem(c, x, y+measures[i].height/2, it, measures[i], textStyle, iconW, iconH)
		y += measures[i].height + gap
	}
}

func defaultLegendY(bounds geom.Rect, opt *option.Legend, totalH float64) float64 {
	y := bounds.Y + 60
	if opt != nil && opt.Top.Set {
		y = bounds.Y + opt.Top.Resolve(bounds.H, 60)
	}
	if opt != nil && opt.Bottom.Set {
		y = bounds.Bottom() - opt.Bottom.Resolve(bounds.H, 0) - totalH
	}
	if opt != nil && opt.Top.Keyword == "middle" {
		y = bounds.Y + bounds.H/2 - totalH/2
	}
	if opt != nil && opt.Top.Keyword == "bottom" {
		y = bounds.Bottom() - totalH - 10
	}
	return y
}

func legendX(bounds geom.Rect, opt *option.Legend, totalW float64) float64 {
	x := bounds.X + bounds.W/2 - totalW/2
	if opt == nil {
		return x
	}
	if opt.Left.Set {
		x = bounds.X + opt.Left.Resolve(bounds.W, 0)
	}
	if opt.Right.Set {
		x = bounds.Right() - opt.Right.Resolve(bounds.W, 0) - totalW
	}
	switch opt.Left.Keyword {
	case "left":
		x = bounds.X + 10
	case "center":
		x = bounds.X + bounds.W/2 - totalW/2
	case "right":
		x = bounds.Right() - totalW - 10
	}
	return x
}

func legendRowWidth(row []int, measures []legendMeasure, gap float64) float64 {
	w := 0.0
	for i, idx := range row {
		if i > 0 {
			w += gap
		}
		w += measures[idx].width
	}
	return w
}

func drawLegendItem(c zcanvas.Canvas, x, y float64, it LegendItem, measure legendMeasure, textStyle zcanvas.TextStyle, iconW, iconH float64) {
	c.SetFill(it.Color)
	c.NoStroke()
	c.DrawRect(x, y-iconH/2, iconW, iconH)
	textStyle.VAlign = zcanvas.AlignBaseline
	c.DrawText(x+iconW+4, y+measure.textAscent-measure.textHeight/2+iconH+3, it.Name, textStyle)
}

func pickString(over, fallback string) string {
	if over != "" {
		return over
	}
	return fallback
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
