// Package layout 负责把图表的子组件（标题、图例、网格、坐标轴）布置到画布上。
package layout

import (
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
)

// 直角坐标系默认 padding，给坐标轴标签预留空间。
const (
	defaultLeftMargin   = 60
	defaultRightMargin  = 30
	defaultTopMargin    = 80 // 给标题 + 图例预留
	defaultBottomMargin = 50
)

// ComputeGridRect 在 bounds（整张画布）中计算直角坐标系绘图区。
// 优先使用 grid 配置中的 left/right/top/bottom；未设置时使用默认 margin。
func ComputeGridRect(bounds geom.Rect, grid *option.Grid, hasTitle bool) geom.Rect {
	top := defaultTopMargin
	if !hasTitle {
		top = 30
	}
	left, right, bottom := defaultLeftMargin, defaultRightMargin, defaultBottomMargin
	// 小画布自适应：默认 margin 不应吃掉过多绘图区，否则小尺寸图表坐标轴会挤成一团。
	left = clampMargin(left, bounds.W, 0.34)
	right = clampMargin(right, bounds.W, 0.18)
	top = clampMargin(top, bounds.H, 0.40)
	bottom = clampMargin(bottom, bounds.H, 0.34)
	if grid != nil {
		if grid.Left.Set {
			left = int(grid.Left.Resolve(bounds.W, float64(left)))
		}
		if grid.Right.Set {
			right = int(grid.Right.Resolve(bounds.W, float64(right)))
		}
		if grid.Top.Set {
			top = int(grid.Top.Resolve(bounds.H, float64(top)))
		}
		if grid.Bottom.Set {
			bottom = int(grid.Bottom.Resolve(bounds.H, float64(bottom)))
		}
	}
	r := bounds.Inset(float64(top), float64(right), float64(bottom), float64(left))
	if r.W < 30 || r.H < 30 {
		// 退化保护：避免负数尺寸
		r = bounds.InsetUniform(12)
	}
	return r
}

// clampMargin 把默认 margin 限制在画布尺寸的 frac 比例内，保证小画布仍有足够绘图区。
func clampMargin(m int, extent, frac float64) int {
	if max := int(extent * frac); m > max {
		return max
	}
	return m
}
