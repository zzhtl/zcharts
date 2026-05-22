package layout

import (
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
	// 默认顶部居中（在 bounds 顶部 8px 处）
	textStyle := zcanvas.TextStyle{
		Family: th.TextStyle.FontFamily,
		Size:   th.Legend.Text.FontSize,
		Color:  th.Legend.Text.Color,
		VAlign: zcanvas.AlignMiddle,
		Anchor: zcanvas.AnchorStart,
	}

	iconW, iconH := 14.0, 10.0
	gap := th.Legend.ItemGap
	if gap <= 0 {
		gap = 10
	}

	// 计算总宽
	totalW := 0.0
	measures := make([]float64, len(items))
	for i, it := range items {
		w, _, _ := c.MeasureText(it.Name, textStyle)
		measures[i] = w
		totalW += iconW + 4 + w
	}
	totalW += gap * float64(len(items)-1)

	x := bounds.X + bounds.W/2 - totalW/2
	if opt != nil {
		if opt.Left.Set {
			x = bounds.X + opt.Left.Resolve(bounds.W, 0)
		}
		switch opt.Left.Keyword {
		case "left":
			x = bounds.X + 10
		case "right":
			x = bounds.X + bounds.W - totalW - 10
		}
	}
	// 默认在标题+副标题下方：标题 16+18=34，副标题 12+gap=20，再留 6px → 60
	y := bounds.Y + 60
	if opt != nil && opt.Top.Set {
		y = bounds.Y + opt.Top.Resolve(bounds.H, 60)
	}

	for i, it := range items {
		// 图标（实心小矩形）
		c.SetFill(it.Color)
		c.NoStroke()
		c.DrawRect(x, y-iconH/2, iconW, iconH)
		// 文字
		c.DrawText(x+iconW+4, y, it.Name, textStyle)
		x += iconW + 4 + measures[i] + gap
	}
}
