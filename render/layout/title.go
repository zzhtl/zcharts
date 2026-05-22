package layout

import (
	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/theme"
)

// DrawTitle 在 bounds 顶部绘制标题。
// 未配置 title 或 text/subtext 都为空时不画。
func DrawTitle(c zcanvas.Canvas, bounds geom.Rect, opt *option.Title, th *theme.Theme) {
	if opt == nil {
		return
	}
	if opt.Text == "" && opt.Subtext == "" {
		return
	}

	// 默认位置：水平居中、距顶 16px
	var x, y float64
	if opt.Left.Set {
		x = opt.Left.Resolve(bounds.W, bounds.W/2)
	} else {
		x = bounds.X + bounds.W/2
	}
	if opt.Top.Set {
		y = opt.Top.Resolve(bounds.H, 16) + bounds.Y
	} else {
		y = bounds.Y + 16
	}

	anchor := zcanvas.AnchorMiddle
	switch opt.Left.Keyword {
	case "left":
		anchor = zcanvas.AnchorStart
		x = bounds.X + 10
	case "right":
		anchor = zcanvas.AnchorEnd
		x = bounds.X + bounds.W - 10
	}

	// 主标题
	titleStyle := mergeStyle(opt.TextStyle, th.Title.Text, th)
	titleStyle.Anchor = anchor
	titleStyle.VAlign = zcanvas.AlignTop
	if opt.Text != "" {
		c.DrawText(x, y, opt.Text, titleStyle)
		y += titleStyle.Size + 6
	}
	// 副标题
	if opt.Subtext != "" {
		subStyle := mergeStyle(opt.SubtextStyle, th.Title.Subtext, th)
		subStyle.Anchor = anchor
		subStyle.VAlign = zcanvas.AlignTop
		c.DrawText(x, y, opt.Subtext, subStyle)
	}
}

// mergeStyle 把 option.TextStyle 套到 theme.TextStyle 之上，再转换为 canvas.TextStyle。
func mergeStyle(over option.TextStyle, base theme.TextStyle, th *theme.Theme) zcanvas.TextStyle {
	out := zcanvas.TextStyle{
		Family: base.FontFamily,
		Size:   base.FontSize,
		Color:  base.Color,
		Weight: base.FontWeight,
	}
	if over.Color != "" {
		if c, err := color.Parse(string(over.Color)); err == nil {
			out.Color = c
		}
	}
	if over.FontFamily != "" {
		out.Family = over.FontFamily
	}
	if over.FontSize > 0 {
		out.Size = over.FontSize
	}
	if over.FontWeight != "" {
		out.Weight = over.FontWeight
	}
	// 在主题中也回退一次默认字体
	if out.Family == "" && th != nil {
		out.Family = th.TextStyle.FontFamily
	}
	return out
}
