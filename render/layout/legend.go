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

// LegendPlacement 描述图例最终占据的区域，供主布局为绘图区预留空间。
// Side 取 "top"/"bottom"/"left"/"right"；无图例时 Side 为空、Rect 为零值。
type LegendPlacement struct {
	Rect geom.Rect
	Side string
}

// DrawLegend 在 bounds 内绘制图例。topInset 为标题底部到 bounds 顶的距离，
// 用于让默认（顶部）图例落在标题下方。
func DrawLegend(c zcanvas.Canvas, bounds geom.Rect, items []LegendItem, opt *option.Legend, th *theme.Theme, topInset float64) {
	if len(items) == 0 || opt != nil && !opt.IsShown() {
		return
	}
	st := resolveLegendStyle(c, items, opt, th)
	g := layoutLegend(bounds, items, st, opt, topInset)
	drawLegendGeom(c, items, st, g)
}

// MeasureLegend 在不绘制的前提下计算图例占据的区域，供主布局预留空间。
// 与 DrawLegend 使用同一套排布逻辑，保证测量与实际绘制位置一致。
func MeasureLegend(c zcanvas.Canvas, bounds geom.Rect, items []LegendItem, opt *option.Legend, th *theme.Theme, topInset float64) LegendPlacement {
	if len(items) == 0 || opt != nil && !opt.IsShown() {
		return LegendPlacement{}
	}
	st := resolveLegendStyle(c, items, opt, th)
	g := layoutLegend(bounds, items, st, opt, topInset)
	return LegendPlacement{Rect: g.box, Side: legendSide(opt)}
}

// legendStyle 是解析后的图例样式与度量。
type legendStyle struct {
	text     zcanvas.TextStyle
	iconW    float64
	iconH    float64
	gap      float64
	rowGap   float64
	measures []legendMeasure
}

type legendMeasure struct {
	width      float64
	height     float64
	textHeight float64
	textAscent float64
}

// legendGeom 是图例的几何排布结果。
type legendGeom struct {
	orient     string
	rows       [][]int   // 横向布局每行包含的 item 下标
	rowHeights []float64 // 与 rows 对应
	box        geom.Rect // 图例外接矩形（横向为整条带，纵向为整列）
}

func resolveLegendStyle(c zcanvas.Canvas, items []LegendItem, opt *option.Legend, th *theme.Theme) legendStyle {
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
		tw, th2, ascent := c.MeasureText(it.Name, textStyle)
		measures[i] = legendMeasure{
			width:      iconW + 4 + tw,
			height:     maxFloat(iconH, th2),
			textHeight: th2,
			textAscent: ascent,
		}
	}

	return legendStyle{text: textStyle, iconW: iconW, iconH: iconH, gap: gap, rowGap: 8, measures: measures}
}

func layoutLegend(bounds geom.Rect, items []LegendItem, st legendStyle, opt *option.Legend, topInset float64) legendGeom {
	if opt != nil && strings.EqualFold(opt.Orient, "vertical") {
		return layoutVerticalLegend(bounds, items, st, opt, topInset)
	}
	return layoutHorizontalLegend(bounds, items, st, opt, topInset)
}

func layoutHorizontalLegend(bounds geom.Rect, items []LegendItem, st legendStyle, opt *option.Legend, topInset float64) legendGeom {
	maxW := bounds.W - 20
	if maxW < 20 {
		maxW = bounds.W
	}

	rows := make([][]int, 0, 1)
	row := make([]int, 0, len(items))
	rowW := 0.0
	for i := range items {
		nextW := st.measures[i].width
		if len(row) > 0 {
			nextW += st.gap
		}
		if len(row) > 0 && rowW+nextW > maxW {
			rows = append(rows, row)
			row = []int{i}
			rowW = st.measures[i].width
			continue
		}
		row = append(row, i)
		rowW += nextW
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	rowHeights := make([]float64, len(rows))
	totalH := 0.0
	for i, r := range rows {
		for _, idx := range r {
			rowHeights[i] = maxFloat(rowHeights[i], st.measures[idx].height)
		}
		totalH += rowHeights[i]
	}
	totalH += st.rowGap * float64(len(rows)-1)

	y := defaultLegendY(bounds, opt, totalH, topInset)
	return legendGeom{
		orient:     "horizontal",
		rows:       rows,
		rowHeights: rowHeights,
		box:        geom.Rect{X: bounds.X, Y: y, W: bounds.W, H: totalH},
	}
}

func layoutVerticalLegend(bounds geom.Rect, items []LegendItem, st legendStyle, opt *option.Legend, topInset float64) legendGeom {
	totalW := 0.0
	totalH := 0.0
	for _, m := range st.measures {
		totalW = maxFloat(totalW, m.width)
		totalH += m.height
	}
	totalH += st.gap * float64(len(items)-1)

	x := legendX(bounds, opt, totalW)
	y := defaultLegendY(bounds, opt, totalH, topInset)
	return legendGeom{
		orient: "vertical",
		box:    geom.Rect{X: x, Y: y, W: totalW, H: totalH},
	}
}

func drawLegendGeom(c zcanvas.Canvas, items []LegendItem, st legendStyle, g legendGeom) {
	if g.orient == "vertical" {
		y := g.box.Y
		for i := range items {
			drawLegendItem(c, g.box.X, y+st.measures[i].height/2, items[i], st.text, st.iconW, st.iconH)
			y += st.measures[i].height + st.gap
		}
		return
	}
	y := g.box.Y
	for i, r := range g.rows {
		rowWidth := legendRowWidth(r, st.measures, st.gap)
		x := legendRowX(g.box, r, rowWidth)
		for _, idx := range r {
			drawLegendItem(c, x, y+g.rowHeights[i]/2, items[idx], st.text, st.iconW, st.iconH)
			x += st.measures[idx].width + st.gap
		}
		y += g.rowHeights[i] + st.rowGap
	}
}

// legendRowX 在外接矩形内按其对齐方式定位单行起点。横向 box 的 X/W 与 bounds 一致，
// 因此这里直接用 box 居中，水平对齐由 layout 阶段的 box 决定。
func legendRowX(box geom.Rect, _ []int, rowWidth float64) float64 {
	return box.X + box.W/2 - rowWidth/2
}

func legendSide(opt *option.Legend) string {
	if opt != nil && strings.EqualFold(opt.Orient, "vertical") {
		if opt.Right.Set || opt.Left.Keyword == "right" {
			return "right"
		}
		return "left"
	}
	if opt != nil && (opt.Bottom.Set || opt.Top.Keyword == "bottom") {
		return "bottom"
	}
	return "top"
}

func defaultLegendY(bounds geom.Rect, opt *option.Legend, totalH, topInset float64) float64 {
	y := bounds.Y + topInset
	if opt != nil && opt.Top.Set {
		y = bounds.Y + opt.Top.Resolve(bounds.H, topInset)
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

// drawLegendItem 在 (x, y) 处绘制色块与文字，y 为该条目的垂直中线。
// 色块以 y 为中心，文字用 AlignMiddle 在同一中线居中。
func drawLegendItem(c zcanvas.Canvas, x, y float64, it LegendItem, textStyle zcanvas.TextStyle, iconW, iconH float64) {
	c.SetFill(it.Color)
	c.NoStroke()
	c.DrawRect(x, y-iconH/2, iconW, iconH)
	textStyle.VAlign = zcanvas.AlignMiddle
	c.DrawText(x+iconW+4, y, it.Name, textStyle)
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
