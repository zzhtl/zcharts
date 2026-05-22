package theme

import (
	"encoding/json"
	"fmt"

	"github.com/zzhtl/zcharts/common/color"
)

// ECharts 主题 JSON 中关心的字段（仅取本阶段用得到的部分）。
type echartsThemeJSON struct {
	Color           []string `json:"color"`
	BackgroundColor string   `json:"backgroundColor"`
	TextStyle       struct {
		Color      string  `json:"color"`
		FontFamily string  `json:"fontFamily"`
		FontSize   float64 `json:"fontSize"`
		FontWeight string  `json:"fontWeight"`
	} `json:"textStyle"`
	Title struct {
		TextStyle struct {
			Color      string  `json:"color"`
			FontWeight string  `json:"fontWeight"`
			FontSize   float64 `json:"fontSize"`
		} `json:"textStyle"`
	} `json:"title"`
	Legend struct {
		TextStyle struct {
			Color string `json:"color"`
		} `json:"textStyle"`
	} `json:"legend"`
}

// ParseECharts 解析 ECharts 主题 JSON 为 *Theme。base 为 nil 时以 Default() 为底。
// 未识别字段会被忽略。
func ParseECharts(name string, data []byte, base *Theme) (*Theme, error) {
	if base == nil {
		base = Default()
	}
	out := *base
	out.Name = name

	var src echartsThemeJSON
	if err := json.Unmarshal(data, &src); err != nil {
		return nil, fmt.Errorf("zcharts/theme: parse echarts json: %w", err)
	}
	if len(src.Color) > 0 {
		p := make(color.Palette, 0, len(src.Color))
		for _, hex := range src.Color {
			c, err := color.Parse(hex)
			if err != nil {
				return nil, err
			}
			p = append(p, c)
		}
		out.Palette = p
	}
	if src.BackgroundColor != "" {
		if c, err := color.Parse(src.BackgroundColor); err == nil {
			out.BackgroundColor = c
		}
	}
	if src.TextStyle.Color != "" {
		if c, err := color.Parse(src.TextStyle.Color); err == nil {
			out.TextStyle.Color = c
		}
	}
	if src.TextStyle.FontSize > 0 {
		out.TextStyle.FontSize = src.TextStyle.FontSize
	}
	if src.TextStyle.FontFamily != "" {
		out.TextStyle.FontFamily = src.TextStyle.FontFamily
	}
	if src.TextStyle.FontWeight != "" {
		out.TextStyle.FontWeight = src.TextStyle.FontWeight
	}
	if src.Title.TextStyle.Color != "" {
		if c, err := color.Parse(src.Title.TextStyle.Color); err == nil {
			out.Title.Text.Color = c
		}
	}
	if src.Title.TextStyle.FontSize > 0 {
		out.Title.Text.FontSize = src.Title.TextStyle.FontSize
	}
	if src.Title.TextStyle.FontWeight != "" {
		out.Title.Text.FontWeight = src.Title.TextStyle.FontWeight
	}
	if src.Legend.TextStyle.Color != "" {
		if c, err := color.Parse(src.Legend.TextStyle.Color); err == nil {
			out.Legend.Text.Color = c
		}
	}
	return &out, nil
}

// RegisterECharts 解析 ECharts 主题 JSON 并注册。
func RegisterECharts(name string, data []byte) error {
	t, err := ParseECharts(name, data, nil)
	if err != nil {
		return err
	}
	Register(name, func() *Theme { return cloneTheme(t) })
	return nil
}

func cloneTheme(t *Theme) *Theme {
	if t == nil {
		return nil
	}
	cp := *t
	if t.Palette != nil {
		cp.Palette = append(color.Palette(nil), t.Palette...)
	}
	return &cp
}
