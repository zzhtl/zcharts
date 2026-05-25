// Package render 是 zcharts 的渲染引擎。
//
// 它接受一个 option.Option（图表配置）和一个 canvas.Canvas（绘制目标），
// 计算布局后把各 series 绘制到 canvas 上。
package render

import (
	"github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/layout"
	"github.com/zzhtl/zcharts/render/scale"
	"github.com/zzhtl/zcharts/theme"
)

// Context 在渲染过程中被所有 layout/series 共享。
type Context struct {
	Canvas canvas.Canvas
	Theme  *theme.Theme
	Option *option.Option

	// 整个画布矩形（像素）
	Bounds geom.Rect

	// 直角坐标系的绘图区
	GridRect geom.Rect

	// 图例条目（已按 legend.data 过滤）及其占位，供绘图区预留空间。
	LegendItems []layout.LegendItem
	Legend      layout.LegendPlacement

	// 各 axis 的 scale
	XScales []scale.Scale
	YScales []scale.Scale

	// 已分配给 series 的颜色（按顺序）
	SeriesColors []color.Color
}

// XScale 返回第 idx 个 X 轴的 scale，越界时返回第 0 个。
func (c *Context) XScale(idx int) scale.Scale {
	if idx < 0 || idx >= len(c.XScales) {
		if len(c.XScales) > 0 {
			return c.XScales[0]
		}
		return nil
	}
	return c.XScales[idx]
}

// YScale 同上。
func (c *Context) YScale(idx int) scale.Scale {
	if idx < 0 || idx >= len(c.YScales) {
		if len(c.YScales) > 0 {
			return c.YScales[0]
		}
		return nil
	}
	return c.YScales[idx]
}

// SeriesColor 取第 i 个 series 分配的默认颜色（从主题调色板）。
func (c *Context) SeriesColor(i int) color.Color {
	if i >= 0 && i < len(c.SeriesColors) {
		return c.SeriesColors[i]
	}
	if c.Theme != nil {
		return c.Theme.Palette.At(i)
	}
	return color.Color{}
}
