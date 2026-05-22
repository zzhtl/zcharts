package theme

import "github.com/zzhtl/zcharts/common/color"

// Default 返回与 ECharts 默认主题接近的浅色主题。
// 调色板取自 ECharts 5 默认色系。
func Default() *Theme {
	return &Theme{
		Name:            "default",
		Palette:         color.MustPalette("#5470c6", "#91cc75", "#fac858", "#ee6666", "#73c0de", "#3ba272", "#fc8452", "#9a60b4", "#ea7ccc"),
		BackgroundColor: color.RGB(255, 255, 255),
		TextStyle: TextStyle{
			Color:      color.MustParse("#333"),
			FontFamily: "default",
			FontSize:   12,
			FontWeight: "normal",
		},
		Title: TitleStyle{
			Text:    TextStyle{Color: color.MustParse("#464646"), FontSize: 18, FontWeight: "bold"},
			Subtext: TextStyle{Color: color.MustParse("#6E7079"), FontSize: 12},
		},
		Legend: LegendStyle{
			Text:    TextStyle{Color: color.MustParse("#333"), FontSize: 12},
			ItemGap: 10,
		},
		Axis: AxisStyle{
			AxisLine:  LineStyle{Color: color.MustParse("#6E7079"), Width: 1},
			AxisTick:  LineStyle{Color: color.MustParse("#6E7079"), Width: 1},
			SplitLine: LineStyle{Color: color.MustParse("#E0E6F1"), Width: 1},
			Label:     TextStyle{Color: color.MustParse("#6E7079"), FontSize: 12},
			NameStyle: TextStyle{Color: color.MustParse("#6E7079"), FontSize: 12},
		},
		GridBackground: color.Transparent,
	}
}
