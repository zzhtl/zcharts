// Package theme 提供 zcharts 的主题（颜色、字体、组件样式）抽象。
//
// 主题结构对齐 Apache ECharts 5.x 的主题文件：
//   - color：调色板（series 自动取用）
//   - backgroundColor：画布背景
//   - textStyle：全局文字样式
//   - title/legend/axis/grid 等子组件覆盖样式
//
// 第一阶段实现常用字段；ECharts 主题 JSON 中未识别的字段会被忽略。
package theme

import "github.com/zzhtl/zcharts/common/color"

// TextStyle 描述一段文字的渲染样式。
type TextStyle struct {
	Color      color.Color
	FontFamily string
	FontSize   float64
	FontWeight string // "normal" / "bold"
}

// Merge 用 over 中的非零字段覆盖 t。
func (t TextStyle) Merge(over TextStyle) TextStyle {
	if !over.Color.IsZero() {
		t.Color = over.Color
	}
	if over.FontFamily != "" {
		t.FontFamily = over.FontFamily
	}
	if over.FontSize > 0 {
		t.FontSize = over.FontSize
	}
	if over.FontWeight != "" {
		t.FontWeight = over.FontWeight
	}
	return t
}

// LineStyle 描述线条样式。
type LineStyle struct {
	Color color.Color
	Width float64
	// DashArray 为空表示实线；非空时按数组循环虚线。
	DashArray []float64
}

// AreaStyle 描述填充样式。
type AreaStyle struct {
	Color   color.Color
	Opacity float64 // 0~1
}

// AxisStyle 一组坐标轴样式（同时给 X、Y 轴共用）。
type AxisStyle struct {
	AxisLine  LineStyle // 轴主线
	AxisTick  LineStyle // 刻度短线
	SplitLine LineStyle // 网格分割线
	Label     TextStyle // 刻度标签
	NameStyle TextStyle // 轴名称
}

// LegendStyle 图例样式。
type LegendStyle struct {
	Text       TextStyle
	IconRadius float64 // 图例图标圆角
	ItemGap    float64 // 项间距
}

// TitleStyle 标题样式。
type TitleStyle struct {
	Text    TextStyle
	Subtext TextStyle
}

// Theme 是完整的主题定义。
type Theme struct {
	Name            string
	Palette         color.Palette
	BackgroundColor color.Color
	TextStyle       TextStyle

	Title  TitleStyle
	Legend LegendStyle
	Axis   AxisStyle

	// 直角坐标系中表示绘图区背景（一般透明）。
	GridBackground color.Color
}
