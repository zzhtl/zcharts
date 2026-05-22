package render

import (
	"github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/layout"
	"github.com/zzhtl/zcharts/render/scale"
	"github.com/zzhtl/zcharts/render/series"
	"github.com/zzhtl/zcharts/theme"
)

// Render 把 opt 渲染到 c 上。th 为 nil 时使用 theme.Default()。
func Render(c canvas.Canvas, opt *option.Option, th *theme.Theme) error {
	if th == nil {
		th = theme.Default()
	}
	w, h := c.Size()
	ctx := &Context{
		Canvas: c,
		Theme:  th,
		Option: opt,
		Bounds: geom.Rect{W: w, H: h},
	}

	drawBackground(ctx)
	if opt == nil {
		return nil
	}

	assignColors(ctx)

	hasCartesian, hasPolar := false, false
	for _, s := range opt.Series {
		switch s.Kind() {
		case option.KindLine, option.KindBar, option.KindScatter, option.KindHeatmap:
			hasCartesian = true
		case option.KindPie, option.KindRadar, option.KindGauge:
			hasPolar = true
		}
	}

	if hasCartesian {
		hasTitle := opt.Title != nil && (opt.Title.Text != "" || opt.Title.Subtext != "")
		ctx.GridRect = layout.ComputeGridRect(ctx.Bounds, opt.Grid, hasTitle)
		buildCartesianScales(ctx)
		drawCartesianAxes(ctx)
		drawCartesianSeries(ctx)
	}

	if hasPolar {
		drawPolarSeries(ctx)
	}

	// 标题
	layout.DrawTitle(c, ctx.Bounds, opt.Title, th)

	// 图例（基于 series 名称）
	drawLegend(ctx)

	return nil
}

// ---- 背景 ----

func drawBackground(ctx *Context) {
	bg := ctx.Theme.BackgroundColor
	if ctx.Option != nil && ctx.Option.BackgroundColor != "" {
		if cc, err := color.Parse(string(ctx.Option.BackgroundColor)); err == nil {
			bg = cc
		}
	}
	if bg.IsZero() {
		return
	}
	ctx.Canvas.SetFill(bg)
	ctx.Canvas.NoStroke()
	ctx.Canvas.DrawRect(0, 0, ctx.Bounds.W, ctx.Bounds.H)
}

// ---- 颜色分配 ----

func assignColors(ctx *Context) {
	palette := ctx.Theme.Palette
	if ctx.Option != nil && len(ctx.Option.Color) > 0 {
		// 用户覆盖
		p := make(color.Palette, 0, len(ctx.Option.Color))
		for _, hex := range ctx.Option.Color {
			if cc, err := color.Parse(hex); err == nil {
				p = append(p, cc)
			}
		}
		if len(p) > 0 {
			palette = p
		}
	}
	if ctx.Option == nil {
		return
	}
	ctx.SeriesColors = make([]color.Color, len(ctx.Option.Series))
	for i, s := range ctx.Option.Series {
		// series.color 覆盖
		base := palette.At(i)
		switch v := s.(type) {
		case *option.LineSeries:
			if v.Color != "" {
				if c, err := color.Parse(string(v.Color)); err == nil {
					base = c
				}
			}
		case *option.BarSeries:
			if v.Color != "" {
				if c, err := color.Parse(string(v.Color)); err == nil {
					base = c
				}
			}
		case *option.ScatterSeries:
			if v.Color != "" {
				if c, err := color.Parse(string(v.Color)); err == nil {
					base = c
				}
			}
		case *option.PieSeries:
			if v.Color != "" {
				if c, err := color.Parse(string(v.Color)); err == nil {
					base = c
				}
			}
		}
		ctx.SeriesColors[i] = base
	}
}

// ---- 直角坐标系 ----

func buildCartesianScales(ctx *Context) {
	opt := ctx.Option
	// X 轴
	if len(opt.XAxis) == 0 {
		opt.XAxis = option.AxisList{{Type: option.AxisCategory}}
	}
	if len(opt.YAxis) == 0 {
		opt.YAxis = option.AxisList{{Type: option.AxisValue}}
	}
	ctx.XScales = make([]scale.Scale, len(opt.XAxis))
	ctx.YScales = make([]scale.Scale, len(opt.YAxis))

	for i, ax := range opt.XAxis {
		ctx.XScales[i] = buildAxisScale(ax, ctx, true)
		ctx.XScales[i].SetPixelRange(ctx.GridRect.X, ctx.GridRect.Right())
	}
	for i, ax := range opt.YAxis {
		ctx.YScales[i] = buildAxisScale(ax, ctx, false)
		ctx.YScales[i].SetPixelRange(ctx.GridRect.Bottom(), ctx.GridRect.Y) // 反向：Y 值大 → 像素小
	}
}

func buildAxisScale(ax option.Axis, ctx *Context, isX bool) scale.Scale {
	switch ax.Type {
	case option.AxisCategory:
		return scale.NewCategory(ax.Data, ax.IsBoundaryGap())
	default: // AxisValue 默认
		min, max := collectValueRange(ctx, isX)
		if ax.Min != nil {
			min = *ax.Min
		}
		if ax.Max != nil {
			max = *ax.Max
		}
		target := ax.SplitNumber
		if target <= 0 {
			target = 5
		}
		if ax.Min != nil || ax.Max != nil {
			return scale.NewLinearFixed(min, max, target)
		}
		return scale.NewLinear(min, max, target, true)
	}
}

// collectValueRange 扫描所有相关 series 的数据值，返回 [min, max]。
// isX=true 时取 scatter 的 X 数值；false 时取 line/bar/scatter 的 Y 数值。
func collectValueRange(ctx *Context, isX bool) (float64, float64) {
	min, max := +1e30, -1e30
	for _, s := range ctx.Option.Series {
		switch v := s.(type) {
		case *option.LineSeries:
			if isX {
				continue
			}
			for _, d := range v.Data {
				val := d.Number()
				if val < min {
					min = val
				}
				if val > max {
					max = val
				}
			}
		case *option.BarSeries:
			if isX {
				continue
			}
			for _, d := range v.Data {
				val := d.Number()
				if val < min {
					min = val
				}
				if val > max {
					max = val
				}
			}
		case *option.ScatterSeries:
			for _, d := range v.Data {
				x, y := d.Pair()
				val := y
				if isX {
					val = x
				}
				if val < min {
					min = val
				}
				if val > max {
					max = val
				}
			}
		}
	}
	if min > max {
		min, max = 0, 1
	}
	// 让 0 自动入界，避免柱状图从负值开始
	if !isX && min > 0 {
		min = 0
	}
	return min, max
}

func drawCartesianAxes(ctx *Context) {
	// 先画 splitLine，再画 axis 线，让 axis 在最上层（视觉更清晰）
	for i, ax := range ctx.Option.XAxis {
		layout.DrawAxis(ctx.Canvas, ctx.GridRect, layout.SideBottom, ctx.XScales[i], ax, ctx.Theme, true)
	}
	for i, ax := range ctx.Option.YAxis {
		layout.DrawAxis(ctx.Canvas, ctx.GridRect, layout.SideLeft, ctx.YScales[i], ax, ctx.Theme, true)
	}
}

// ---- series 分发 ----

func drawCartesianSeries(ctx *Context) {
	// 预扫一遍 Bar，记录每个 BarSeries 在所有 Bar 中的索引
	barIndex := map[int]int{}
	totalBars := 0
	for i, s := range ctx.Option.Series {
		if _, ok := s.(*option.BarSeries); ok {
			barIndex[i] = totalBars
			totalBars++
		}
	}

	for i, s := range ctx.Option.Series {
		switch v := s.(type) {
		case *option.LineSeries:
			xs := ctx.XScale(v.XAxisIndex)
			ys := ctx.YScale(v.YAxisIndex)
			if xs == nil || ys == nil {
				continue
			}
			series.DrawLine(series.DrawLineArgs{
				Canvas:   ctx.Canvas,
				Series:   v,
				XScale:   xs,
				YScale:   ys,
				GridRect: ctx.GridRect,
				Color:    ctx.SeriesColor(i),
				BaseY:    ctx.GridRect.Bottom(),
				Family:   ctx.Theme.TextStyle.FontFamily,
			})
		case *option.BarSeries:
			xs := ctx.XScale(v.XAxisIndex)
			ys := ctx.YScale(v.YAxisIndex)
			if xs == nil || ys == nil {
				continue
			}
			var bandWidth float64
			if cs, ok := xs.(*scale.Category); ok {
				bandWidth = cs.BandWidth()
			}
			series.DrawBar(series.DrawBarArgs{
				Canvas:         ctx.Canvas,
				Series:         v,
				XScale:         xs,
				YScale:         ys,
				GridRect:       ctx.GridRect,
				Color:          ctx.SeriesColor(i),
				SeriesIndex:    barIndex[i],
				TotalBarSeries: totalBars,
				BandWidth:      bandWidth,
			})
		case *option.ScatterSeries:
			xs := ctx.XScale(v.XAxisIndex)
			ys := ctx.YScale(v.YAxisIndex)
			if xs == nil || ys == nil {
				continue
			}
			series.DrawScatter(series.DrawScatterArgs{
				Canvas:   ctx.Canvas,
				Series:   v,
				XScale:   xs,
				YScale:   ys,
				GridRect: ctx.GridRect,
				Color:    ctx.SeriesColor(i),
			})
		case *option.HeatmapSeries:
			xs := ctx.XScale(v.XAxisIndex)
			ys := ctx.YScale(v.YAxisIndex)
			if xs == nil || ys == nil {
				continue
			}
			series.DrawHeatmap(series.DrawHeatmapArgs{
				Canvas:    ctx.Canvas,
				Series:    v,
				XScale:    xs,
				YScale:    ys,
				GridRect:  ctx.GridRect,
				VisualMap: ctx.Option.VisualMap,
				Theme:     ctx.Theme.Palette,
			})
		}
	}
}

// ---- 极坐标系 series 分发 ----

func drawPolarSeries(ctx *Context) {
	// 为多个 Pie/Radar/Gauge 共用画布时，按 series 顺序绘制（后绘制覆盖先绘制）
	// 第一阶段不做多图分屏布局。
	hasTitle := ctx.Option.Title != nil && (ctx.Option.Title.Text != "" || ctx.Option.Title.Subtext != "")
	innerBounds := ctx.Bounds
	if hasTitle {
		// 给标题留 60px 顶部
		innerBounds = innerBounds.Inset(60, 0, 0, 0)
	}

	for i, s := range ctx.Option.Series {
		switch v := s.(type) {
		case *option.PieSeries:
			series.DrawPie(series.DrawPieArgs{
				Canvas:  ctx.Canvas,
				Series:  v,
				Bounds:  innerBounds,
				Palette: paletteFor(ctx),
				Family:  ctx.Theme.TextStyle.FontFamily,
			})
			_ = i
		case *option.RadarSeries:
			series.DrawRadar(series.DrawRadarArgs{
				Canvas:    ctx.Canvas,
				Radar:     ctx.Option.Radar,
				Series:    v,
				Bounds:    innerBounds,
				Palette:   paletteFor(ctx),
				Family:    ctx.Theme.TextStyle.FontFamily,
				GridColor: ctx.Theme.Axis.SplitLine.Color,
				AxisColor: ctx.Theme.Axis.AxisLine.Color,
				TextColor: ctx.Theme.Axis.Label.Color,
			})
		case *option.GaugeSeries:
			series.DrawGauge(series.DrawGaugeArgs{
				Canvas:     ctx.Canvas,
				Series:     v,
				Bounds:     innerBounds,
				Palette:    paletteFor(ctx),
				Family:     ctx.Theme.TextStyle.FontFamily,
				TrackColor: ctx.Theme.Axis.SplitLine.Color,
				TextColor:  ctx.Theme.TextStyle.Color,
			})
		}
	}
}

func paletteFor(ctx *Context) color.Palette {
	if ctx.Option != nil && len(ctx.Option.Color) > 0 {
		p := make(color.Palette, 0, len(ctx.Option.Color))
		for _, hex := range ctx.Option.Color {
			if cc, err := color.Parse(hex); err == nil {
				p = append(p, cc)
			}
		}
		if len(p) > 0 {
			return p
		}
	}
	return ctx.Theme.Palette
}

// ---- 图例 ----

func drawLegend(ctx *Context) {
	if ctx.Option == nil {
		return
	}
	items := make([]layout.LegendItem, 0, len(ctx.Option.Series))
	for i, s := range ctx.Option.Series {
		name := s.GetName()
		if name == "" {
			continue
		}
		items = append(items, layout.LegendItem{
			Name:  name,
			Color: ctx.SeriesColor(i),
		})
	}
	layout.DrawLegend(ctx.Canvas, ctx.Bounds, items, ctx.Option.Legend, ctx.Theme)
}
