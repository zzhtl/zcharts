package layout

import (
	"strings"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/scale"
	"github.com/zzhtl/zcharts/theme"
)

// AxisSide 描述轴所在方位。
type AxisSide int

const (
	SideBottom AxisSide = iota // X 轴默认
	SideLeft                    // Y 轴默认
	SideTop
	SideRight
)

// DrawAxis 在 grid 矩形周边绘制一根坐标轴。
//
// 参数说明：
//   - rect: 网格绘图区（笛卡尔坐标系内部）
//   - side: 轴所在边
//   - sc:   该轴的 scale（已经 SetPixelRange）
//   - opt:  axis 配置（用于读取 axisLine/label/splitLine 等）
//   - th:   主题
//   - drawSplit: 是否同时绘制 splitLine（分割网格线）
func DrawAxis(c zcanvas.Canvas, rect geom.Rect, side AxisSide, sc scale.Scale, opt option.Axis, th *theme.Theme, drawSplit bool) {
	ticks := sc.Ticks()

	axisColor := pickColor(opt.AxisLine.LineStyle.Color, th.Axis.AxisLine.Color)
	axisWidth := pickFloat(opt.AxisLine.LineStyle.Width, th.Axis.AxisLine.Width)
	labelColor := pickColor(opt.AxisLabel.Color, th.Axis.Label.Color)
	labelSize := pickFloat(opt.AxisLabel.FontSize, th.Axis.Label.FontSize)
	splitColor := pickColor(opt.SplitLine.LineStyle.Color, th.Axis.SplitLine.Color)

	// 轴主线（除非显式 show=false）
	if !isFalse(opt.AxisLine.Show) {
		c.SetStroke(axisColor)
		c.SetStrokeWidth(axisWidth)
		c.NoFill()
		switch side {
		case SideBottom:
			c.DrawLine(rect.X, rect.Bottom(), rect.Right(), rect.Bottom())
		case SideTop:
			c.DrawLine(rect.X, rect.Y, rect.Right(), rect.Y)
		case SideLeft:
			c.DrawLine(rect.X, rect.Y, rect.X, rect.Bottom())
		case SideRight:
			c.DrawLine(rect.Right(), rect.Y, rect.Right(), rect.Bottom())
		}
	}

	// 分割线 + 刻度短线 + label
	showSplit := drawSplit && !isFalse(opt.SplitLine.Show)
	showLabel := !isFalse(opt.AxisLabel.Show)
	showTick := !isFalse(opt.AxisTick.Show)
	tickLen := pickFloat(opt.AxisTick.Length, 5)
	margin := pickFloat(opt.AxisLabel.Margin, 8)

	interval := 1
	if opt.AxisLabel.Interval != nil {
		interval = *opt.AxisLabel.Interval + 1
	}

	for i, tk := range ticks {
		px := sc.Pixel(tk.Value)
		// 越界过滤（避免在 rect 外画）
		switch side {
		case SideBottom, SideTop:
			if px < rect.X-0.5 || px > rect.Right()+0.5 {
				continue
			}
		case SideLeft, SideRight:
			if px < rect.Y-0.5 || px > rect.Bottom()+0.5 {
				continue
			}
		}

		// splitLine
		if showSplit && i > 0 && i < len(ticks)-1 {
			// 端点不画分割线（与 axisLine 重叠）
			c.SetStroke(splitColor)
			c.SetStrokeWidth(1)
			switch side {
			case SideBottom, SideTop:
				c.DrawLine(px, rect.Y, px, rect.Bottom())
			case SideLeft, SideRight:
				c.DrawLine(rect.X, px, rect.Right(), px)
			}
		}

		// 刻度短线
		if showTick {
			c.SetStroke(axisColor)
			c.SetStrokeWidth(axisWidth)
			switch side {
			case SideBottom:
				c.DrawLine(px, rect.Bottom(), px, rect.Bottom()+tickLen)
			case SideTop:
				c.DrawLine(px, rect.Y-tickLen, px, rect.Y)
			case SideLeft:
				c.DrawLine(rect.X-tickLen, px, rect.X, px)
			case SideRight:
				c.DrawLine(rect.Right(), px, rect.Right()+tickLen, px)
			}
		}

		// label
		if !showLabel || (interval > 1 && i%interval != 0) {
			continue
		}
		label := tk.Label
		if opt.AxisLabel.Formatter != "" {
			label = strings.ReplaceAll(opt.AxisLabel.Formatter, "{value}", label)
		}
		labelStyle := zcanvas.TextStyle{
			Family: th.TextStyle.FontFamily,
			Size:   labelSize,
			Color:  labelColor,
			Rotation: opt.AxisLabel.Rotate,
		}
		switch side {
		case SideBottom:
			labelStyle.Anchor = zcanvas.AnchorMiddle
			labelStyle.VAlign = zcanvas.AlignTop
			c.DrawText(px, rect.Bottom()+tickLen+margin, label, labelStyle)
		case SideTop:
			labelStyle.Anchor = zcanvas.AnchorMiddle
			labelStyle.VAlign = zcanvas.AlignBottom
			c.DrawText(px, rect.Y-tickLen-margin, label, labelStyle)
		case SideLeft:
			labelStyle.Anchor = zcanvas.AnchorEnd
			labelStyle.VAlign = zcanvas.AlignMiddle
			c.DrawText(rect.X-tickLen-margin, px, label, labelStyle)
		case SideRight:
			labelStyle.Anchor = zcanvas.AnchorStart
			labelStyle.VAlign = zcanvas.AlignMiddle
			c.DrawText(rect.Right()+tickLen+margin, px, label, labelStyle)
		}
	}

	// 轴名称
	if opt.Name != "" {
		nameStyle := zcanvas.TextStyle{
			Family: th.TextStyle.FontFamily,
			Size:   pickFloat(opt.NameTextStyle.FontSize, th.Axis.NameStyle.FontSize),
			Color:  pickColor(opt.NameTextStyle.Color, th.Axis.NameStyle.Color),
		}
		gap := pickFloat(opt.NameGap, 18)
		switch side {
		case SideBottom:
			nameStyle.Anchor = zcanvas.AnchorMiddle
			nameStyle.VAlign = zcanvas.AlignTop
			c.DrawText(rect.X+rect.W/2, rect.Bottom()+gap+labelSize+10, opt.Name, nameStyle)
		case SideLeft:
			nameStyle.Anchor = zcanvas.AnchorMiddle
			nameStyle.VAlign = zcanvas.AlignBottom
			c.DrawText(rect.X, rect.Y-gap, opt.Name, nameStyle)
		}
	}
}

func isFalse(p *bool) bool { return p != nil && !*p }

func pickColor(over option.ColorString, fallback color.Color) color.Color {
	if over != "" {
		if c, err := color.Parse(string(over)); err == nil {
			return c
		}
	}
	return fallback
}

func pickFloat(over, fallback float64) float64 {
	if over > 0 {
		return over
	}
	return fallback
}
