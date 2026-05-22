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
	if r.W < 50 || r.H < 50 {
		// 退化保护：避免负数尺寸
		r = bounds.InsetUniform(20)
	}
	return r
}
