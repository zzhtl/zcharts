package series

import (
	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/option"
)

// labelDefaultColor 是数据标签的默认文字色（深灰，保证可读）。
var labelDefaultColor = color.MustParse("#555555")

// drawDataLabel 在 (x,y) 处绘制一个数据标签，文字按 anchor/valign 对齐，供 bar/line/scatter 共用。
// 颜色/字号缺省时回退到深灰 11px。text 为空则不绘制。
func drawDataLabel(c zcanvas.Canvas, lbl option.Label, family, text string, x, y float64, anchor zcanvas.TextAnchor, valign zcanvas.VerticalAlign) {
	if text == "" {
		return
	}
	col := labelDefaultColor
	if lbl.Color != "" {
		if cc, err := color.Parse(string(lbl.Color)); err == nil {
			col = cc
		}
	}
	size := lbl.FontSize
	if size <= 0 {
		size = 11
	}
	c.DrawText(x, y, text, zcanvas.TextStyle{
		Family: family,
		Size:   size,
		Color:  col,
		Anchor: anchor,
		VAlign: valign,
	})
}
