// Package option 是 zcharts 的配置数据模型，字段命名与结构对齐 Apache ECharts 5.x。
//
// 第一阶段只覆盖渲染 7 种核心图表（line/bar/pie/scatter/radar/heatmap/gauge）所必需的字段；
// 未识别的字段会被 encoding/json 默默忽略，不会报错，保证 ECharts JSON 的最大兼容性。
package option

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Flex 表示一个可以是数字、像素、百分比的尺寸/位置值。
// 例如 ECharts 中 grid.left 可以是 30、"10%"、"30px" 或 "center"。
type Flex struct {
	Value   float64 // 数值部分
	Unit    string  // "" / "%" / "px"
	Keyword string  // "left"/"center"/"right"/"top"/"middle"/"bottom"/"auto"
	Set     bool    // 用户是否显式设置
}

// IsPercent 报告该值是否是百分比。
func (f Flex) IsPercent() bool { return f.Unit == "%" }

// Resolve 把 Flex 解析为绝对像素值；total 为参考总长（如 grid 宽度）。
// 当 Keyword 非空时，按 "left"/"top"=0，"center"/"middle"=total/2，"right"/"bottom"=total。
// 未设置时返回 fallback。
func (f Flex) Resolve(total, fallback float64) float64 {
	if !f.Set {
		return fallback
	}
	if f.Keyword != "" {
		switch strings.ToLower(f.Keyword) {
		case "left", "top":
			return 0
		case "center", "middle":
			return total / 2
		case "right", "bottom":
			return total
		}
		return fallback
	}
	switch f.Unit {
	case "%":
		return total * f.Value / 100
	default:
		return f.Value
	}
}

func (f *Flex) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		return f.parseString(s)
	}
	// number
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("option/flex: unsupported value %s: %w", data, err)
	}
	f.Value = v
	f.Set = true
	return nil
}

func (f *Flex) parseString(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	f.Set = true
	switch strings.ToLower(s) {
	case "left", "right", "center", "top", "bottom", "middle", "auto":
		f.Keyword = strings.ToLower(s)
		return nil
	}
	switch {
	case strings.HasSuffix(s, "%"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil {
			return fmt.Errorf("option/flex: bad percent %q", s)
		}
		f.Value, f.Unit = v, "%"
		return nil
	case strings.HasSuffix(s, "px"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "px"), 64)
		if err != nil {
			return fmt.Errorf("option/flex: bad px %q", s)
		}
		f.Value, f.Unit = v, "px"
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("option/flex: cannot parse %q", s)
	}
	f.Value = v
	return nil
}

// Strings 是可以从 string 或 []string 解析的字符串数组。
type Strings []string

func (s *Strings) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '[' {
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		*s = arr
		return nil
	}
	var single string
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*s = []string{single}
	return nil
}

// FlexList 是 [Flex]，允许 JSON 中传入 number / string / array。
// 例如 pie 的 radius 可以是 "60%" 或 ["40%", "70%"]。
type FlexList []Flex

func (l *FlexList) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '[' {
		// 元素逐个 Unmarshal 走 Flex.UnmarshalJSON
		var raws []json.RawMessage
		if err := json.Unmarshal(data, &raws); err != nil {
			return err
		}
		out := make([]Flex, len(raws))
		for i, raw := range raws {
			if err := out[i].UnmarshalJSON(raw); err != nil {
				return err
			}
		}
		*l = out
		return nil
	}
	var single Flex
	if err := single.UnmarshalJSON(data); err != nil {
		return err
	}
	*l = []Flex{single}
	return nil
}

// At 取下标 i 的值，越界时返回 fallback。
func (l FlexList) At(i int, fallback Flex) Flex {
	if i < 0 || i >= len(l) {
		return fallback
	}
	return l[i]
}

// ColorString 是 ECharts 中的颜色字符串字段。
// 支持纯字符串（"#abc"/"rgb()"）；ECharts 的 LinearGradient/RadialGradient/Pattern 对象
// 第一阶段不解析，会被静默丢弃。
type ColorString string

func (c *ColorString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*c = ColorString(s)
	}
	// object/array → 静默忽略
	return nil
}

// ColorList 是可以从 string、[]string 解析的颜色数组（对应 ECharts 顶层 color）。
type ColorList []string

func (cl *ColorList) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		out := make([]string, 0, len(arr))
		for _, raw := range arr {
			raw = bytes.TrimSpace(raw)
			if len(raw) > 0 && raw[0] == '"' {
				var s string
				if err := json.Unmarshal(raw, &s); err == nil {
					out = append(out, s)
				}
			}
		}
		*cl = out
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*cl = []string{s}
	}
	return nil
}

// TextStyle 描述文字样式（对齐 ECharts textStyle 字段）。
type TextStyle struct {
	Color      ColorString `json:"color,omitempty"`
	FontFamily string      `json:"fontFamily,omitempty"`
	FontSize   float64     `json:"fontSize,omitempty"`
	FontWeight string      `json:"fontWeight,omitempty"`
	FontStyle  string      `json:"fontStyle,omitempty"`
}

// LineStyle 描述线条样式（对齐 ECharts lineStyle 字段）。
type LineStyle struct {
	Color   ColorString `json:"color,omitempty"`
	Width   float64     `json:"width,omitempty"`
	Type    string      `json:"type,omitempty"` // "solid"/"dashed"/"dotted"
	Opacity *float64    `json:"opacity,omitempty"`
}

// AreaStyle 描述填充样式。
type AreaStyle struct {
	Color   ColorString `json:"color,omitempty"`
	Opacity *float64    `json:"opacity,omitempty"`
}

// ItemStyle 描述单个图元样式。
type ItemStyle struct {
	Color        ColorString `json:"color,omitempty"`
	BorderColor  ColorString `json:"borderColor,omitempty"`
	BorderWidth  float64     `json:"borderWidth,omitempty"`
	BorderRadius float64     `json:"borderRadius,omitempty"` // 柱状圆角半径（像素）
	Opacity      *float64    `json:"opacity,omitempty"`
}

// Label 描述图元上的文字标签（对齐 ECharts label 字段）。
type Label struct {
	Show      bool        `json:"show,omitempty"`
	Position  string      `json:"position,omitempty"` // "top"/"left"/"right"/"bottom"/"inside" 等
	Color     ColorString `json:"color,omitempty"`
	FontSize  float64     `json:"fontSize,omitempty"`
	Formatter string      `json:"formatter,omitempty"` // 仅支持 "{a}"/"{b}"/"{c}"/"{d}" 简单替换
}
