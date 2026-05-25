package render

import (
	"fmt"

	"github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render/layout"
	"github.com/zzhtl/zcharts/render/scale"
	"github.com/zzhtl/zcharts/render/series"
	"github.com/zzhtl/zcharts/theme"
)

// SeriesRenderer 允许用户为未内置的 series.type 注册渲染逻辑。
type SeriesRenderer func(ctx *Context, s option.Series, index int) error

// RenderOptions 控制渲染引擎的扩展行为。
type RenderOptions struct {
	SeriesRenderers map[option.SeriesKind]SeriesRenderer
}

// Render 把 opt 渲染到 c 上。th 为 nil 时使用 theme.Default()。
func Render(c canvas.Canvas, opt *option.Option, th *theme.Theme) error {
	return RenderWithOptions(c, opt, th, RenderOptions{})
}

// RenderWithOptions 把 opt 渲染到 c 上，并允许传入自定义 series 渲染器。
func RenderWithOptions(c canvas.Canvas, opt *option.Option, th *theme.Theme, opts RenderOptions) error {
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

	hasCartesian, hasPolar, hasFree := false, false, false
	for _, s := range opt.Series {
		switch s.Kind() {
		case option.KindLine, option.KindBar, option.KindScatter, option.KindHeatmap:
			hasCartesian = true
		case option.KindPie, option.KindRadar, option.KindGauge:
			hasPolar = true
		case option.KindFunnel, option.KindTimeline, option.KindWordCloud:
			hasFree = true
		default:
			if v, ok := s.(*option.CustomSeries); ok && v.CoordinateSystem == "cartesian2d" {
				hasCartesian = true
			}
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
	if hasFree {
		drawFreeSeries(ctx)
	}
	if err := drawCustomSeries(ctx, opts.SeriesRenderers); err != nil {
		return err
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
		case *option.FunnelSeries:
			if v.Color != "" {
				if c, err := color.Parse(string(v.Color)); err == nil {
					base = c
				}
			}
		case *option.TimelineSeries:
			if v.Color != "" {
				if c, err := color.Parse(string(v.Color)); err == nil {
					base = c
				}
			}
		case *option.WordCloudSeries:
			if v.Color != "" {
				if c, err := color.Parse(string(v.Color)); err == nil {
					base = c
				}
			}
		case *option.CustomSeries:
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
	case option.AxisTime:
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
		return scale.NewTime(min, max, target)
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
	barStacks := map[string][]stackRange{}
	for _, s := range ctx.Option.Series {
		switch v := s.(type) {
		case *option.LineSeries:
			for _, d := range v.Data {
				val := d.Number()
				if isX {
					if len(d.Values) < 2 {
						continue
					}
					val = d.Values[0]
				} else if usesDataX(ctx, v.XAxisIndex) && len(d.Values) >= 2 {
					val = d.Values[1]
				}
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
			if v.Stack != "" {
				key := barStackKey(v)
				values := barStacks[key]
				if len(values) < len(v.Data) {
					next := make([]stackRange, len(v.Data))
					copy(next, values)
					values = next
				}
				for i, d := range v.Data {
					val := d.Number()
					if val >= 0 {
						values[i].Positive += val
					} else {
						values[i].Negative += val
					}
				}
				barStacks[key] = values
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
	for _, values := range barStacks {
		for _, v := range values {
			if v.Positive < min {
				min = v.Positive
			}
			if v.Positive > max {
				max = v.Positive
			}
			if v.Negative < min {
				min = v.Negative
			}
			if v.Negative > max {
				max = v.Negative
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

type stackRange struct {
	Positive float64
	Negative float64
}

func usesDataX(ctx *Context, axisIndex int) bool {
	if ctx == nil || ctx.Option == nil {
		return false
	}
	if axisIndex < 0 || axisIndex >= len(ctx.Option.XAxis) {
		axisIndex = 0
	}
	if axisIndex >= len(ctx.Option.XAxis) {
		return false
	}
	t := ctx.Option.XAxis[axisIndex].Type
	return t == option.AxisTime || t == option.AxisValue || t == option.AxisLog
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
	barInfos := computeBarDrawInfos(ctx.Option.Series)

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
				UseDataX: usesDataX(ctx, v.XAxisIndex),
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
			info := barInfos[i]
			series.DrawBar(series.DrawBarArgs{
				Canvas:         ctx.Canvas,
				Series:         v,
				XScale:         xs,
				YScale:         ys,
				GridRect:       ctx.GridRect,
				Color:          ctx.SeriesColor(i),
				SeriesIndex:    info.GroupIndex,
				TotalBarSeries: info.TotalGroups,
				BandWidth:      bandWidth,
				StackBases:     info.StackBases,
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

type barDrawInfo struct {
	GroupIndex  int
	TotalGroups int
	StackBases  []float64
}

func computeBarDrawInfos(seriesList option.SeriesList) map[int]barDrawInfo {
	groupIndex := map[string]int{}
	infos := map[int]barDrawInfo{}
	positiveStacks := map[string][]float64{}
	negativeStacks := map[string][]float64{}
	totalGroups := 0

	for i, s := range seriesList {
		bar, ok := s.(*option.BarSeries)
		if !ok {
			continue
		}
		groupKey := barGroupKey(i, bar)
		if _, ok := groupIndex[groupKey]; !ok {
			groupIndex[groupKey] = totalGroups
			totalGroups++
		}

		bases := make([]float64, len(bar.Data))
		if bar.Stack != "" {
			stackKey := barStackKey(bar)
			pos := ensureStackLen(positiveStacks[stackKey], len(bar.Data))
			neg := ensureStackLen(negativeStacks[stackKey], len(bar.Data))
			for j, d := range bar.Data {
				val := d.Number()
				if val >= 0 {
					bases[j] = pos[j]
					pos[j] += val
				} else {
					bases[j] = neg[j]
					neg[j] += val
				}
			}
			positiveStacks[stackKey] = pos
			negativeStacks[stackKey] = neg
		}

		infos[i] = barDrawInfo{GroupIndex: groupIndex[groupKey], StackBases: bases}
	}
	for i, info := range infos {
		info.TotalGroups = totalGroups
		infos[i] = info
	}
	return infos
}

func ensureStackLen(values []float64, n int) []float64 {
	if len(values) >= n {
		return values
	}
	out := make([]float64, n)
	copy(out, values)
	return out
}

func barGroupKey(index int, s *option.BarSeries) string {
	if s.Stack != "" {
		return barStackKey(s)
	}
	return fmt.Sprintf("%d/%d/series-%d", s.XAxisIndex, s.YAxisIndex, index)
}

func barStackKey(s *option.BarSeries) string {
	return fmt.Sprintf("%d/%d/%s", s.XAxisIndex, s.YAxisIndex, s.Stack)
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

func drawFreeSeries(ctx *Context) {
	innerBounds := contentBounds(ctx)
	for _, s := range ctx.Option.Series {
		switch v := s.(type) {
		case *option.FunnelSeries:
			series.DrawFunnel(series.DrawFunnelArgs{
				Canvas:  ctx.Canvas,
				Series:  v,
				Bounds:  innerBounds,
				Palette: paletteFor(ctx),
				Family:  ctx.Theme.TextStyle.FontFamily,
			})
		case *option.TimelineSeries:
			series.DrawTimeline(series.DrawTimelineArgs{
				Canvas:    ctx.Canvas,
				Series:    v,
				Bounds:    innerBounds,
				Palette:   paletteFor(ctx),
				Family:    ctx.Theme.TextStyle.FontFamily,
				TextColor: ctx.Theme.TextStyle.Color,
				MutedText: ctx.Theme.Axis.Label.Color,
				LineColor: ctx.Theme.Axis.SplitLine.Color,
			})
		case *option.WordCloudSeries:
			series.DrawWordCloud(series.DrawWordCloudArgs{
				Canvas:  ctx.Canvas,
				Series:  v,
				Bounds:  innerBounds,
				Palette: paletteFor(ctx),
				Family:  ctx.Theme.TextStyle.FontFamily,
			})
		}
	}
}

func contentBounds(ctx *Context) geom.Rect {
	hasTitle := ctx.Option.Title != nil && (ctx.Option.Title.Text != "" || ctx.Option.Title.Subtext != "")
	if hasTitle {
		return ctx.Bounds.Inset(60, 0, 0, 0)
	}
	return ctx.Bounds
}

func drawCustomSeries(ctx *Context, renderers map[option.SeriesKind]SeriesRenderer) error {
	if len(renderers) == 0 {
		return nil
	}
	for i, s := range ctx.Option.Series {
		renderer := renderers[s.Kind()]
		if renderer == nil {
			continue
		}
		if err := renderer(ctx, s, i); err != nil {
			return fmt.Errorf("render custom series %q: %w", s.Kind(), err)
		}
	}
	return nil
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
	seen := map[string]bool{}
	for i, s := range ctx.Option.Series {
		switch v := s.(type) {
		case *option.PieSeries:
			for j, d := range v.Data {
				addLegendItem(&items, seen, d.Name, paletteFor(ctx).At(j))
			}
		case *option.RadarSeries:
			for j, d := range v.Data {
				addLegendItem(&items, seen, d.Name, paletteFor(ctx).At(j))
			}
			addLegendItem(&items, seen, v.GetName(), ctx.SeriesColor(i))
		case *option.FunnelSeries:
			for j, d := range v.Data {
				addLegendItem(&items, seen, d.Name, paletteFor(ctx).At(j))
			}
		default:
			addLegendItem(&items, seen, s.GetName(), ctx.SeriesColor(i))
		}
	}
	if ctx.Option.Legend != nil && len(ctx.Option.Legend.Data) > 0 {
		items = filterLegendItems(items, ctx.Option.Legend.Data)
	}
	layout.DrawLegend(ctx.Canvas, ctx.Bounds, items, ctx.Option.Legend, ctx.Theme)
}

func addLegendItem(items *[]layout.LegendItem, seen map[string]bool, name string, col color.Color) {
	if name == "" || seen[name] {
		return
	}
	seen[name] = true
	*items = append(*items, layout.LegendItem{Name: name, Color: col})
}

func filterLegendItems(items []layout.LegendItem, names option.Strings) []layout.LegendItem {
	byName := make(map[string]layout.LegendItem, len(items))
	for _, item := range items {
		byName[item.Name] = item
	}
	out := make([]layout.LegendItem, 0, len(names))
	for _, name := range names {
		if item, ok := byName[name]; ok {
			out = append(out, item)
		}
	}
	return out
}
