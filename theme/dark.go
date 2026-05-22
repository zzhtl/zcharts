package theme

import "github.com/zzhtl/zcharts/common/color"

// Dark 返回 ECharts 内置 dark 风格的主题。
func Dark() *Theme {
	return &Theme{
		Name:            "dark",
		Palette:         color.MustPalette("#dd6b66", "#759aa0", "#e69d87", "#8dc1a9", "#ea7e53", "#eedd78", "#73a373", "#73b9bc", "#7289ab", "#91ca8c", "#f49f42"),
		BackgroundColor: color.MustParse("#100C2A"),
		TextStyle: TextStyle{
			Color:      color.MustParse("#B9B8CE"),
			FontFamily: "default",
			FontSize:   12,
			FontWeight: "normal",
		},
		Title: TitleStyle{
			Text:    TextStyle{Color: color.MustParse("#EEF1FA"), FontSize: 18, FontWeight: "bold"},
			Subtext: TextStyle{Color: color.MustParse("#B9B8CE"), FontSize: 12},
		},
		Legend: LegendStyle{
			Text:    TextStyle{Color: color.MustParse("#B9B8CE"), FontSize: 12},
			ItemGap: 10,
		},
		Axis: AxisStyle{
			AxisLine:  LineStyle{Color: color.MustParse("#B9B8CE"), Width: 1},
			AxisTick:  LineStyle{Color: color.MustParse("#B9B8CE"), Width: 1},
			SplitLine: LineStyle{Color: color.MustParse("#484753"), Width: 1},
			Label:     TextStyle{Color: color.MustParse("#B9B8CE"), FontSize: 12},
			NameStyle: TextStyle{Color: color.MustParse("#B9B8CE"), FontSize: 12},
		},
		GridBackground: color.Transparent,
	}
}
