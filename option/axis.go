package option

import (
	"bytes"
	"encoding/json"
)

// AxisType ECharts 轴类型。
const (
	AxisCategory = "category"
	AxisValue    = "value"
	AxisTime     = "time"
	AxisLog      = "log"
)

// Axis 描述一根坐标轴（xAxis/yAxis 共用）。
type Axis struct {
	Show        *bool       `json:"show,omitempty"`
	Type        string      `json:"type,omitempty"` // category/value/time/log
	Name        string      `json:"name,omitempty"`
	NameGap     float64     `json:"nameGap,omitempty"`
	NameLocation string     `json:"nameLocation,omitempty"` // "start"/"middle"/"end"
	NameTextStyle TextStyle `json:"nameTextStyle,omitempty"`

	Data Strings `json:"data,omitempty"` // category 用

	// 数值轴范围；指针型以便区分"未设置"和"设置为 0"
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`

	Inverse     bool `json:"inverse,omitempty"`
	BoundaryGap *boundaryGap `json:"boundaryGap,omitempty"`

	GridIndex int `json:"gridIndex,omitempty"`

	AxisLine  AxisLine  `json:"axisLine,omitempty"`
	AxisTick  AxisTick  `json:"axisTick,omitempty"`
	SplitLine SplitLine `json:"splitLine,omitempty"`
	AxisLabel AxisLabel `json:"axisLabel,omitempty"`

	// SplitNumber 仅作为目标刻度数提示，渲染时按 nice 算法处理
	SplitNumber int `json:"splitNumber,omitempty"`
}

// boundaryGap 在 ECharts 里可以是 bool 或 [string, string]（百分比/像素）。
// 简化为 bool（默认 true）+ 一个 raw 字段保留原始信息（暂未使用）。
type boundaryGap struct {
	Bool bool
}

func (b *boundaryGap) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == 't' || data[0] == 'f' {
		var v bool
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		b.Bool = v
		return nil
	}
	// 数组形式 ["10%", "10%"]：当前不解析，按 true 处理
	b.Bool = true
	return nil
}

// IsBoundaryGap 处理三态：未设置时按 type 决定（category=true, value=false）。
func (a *Axis) IsBoundaryGap() bool {
	if a.BoundaryGap == nil {
		return a.Type == AxisCategory || a.Type == ""
	}
	return a.BoundaryGap.Bool
}

// AxisLine 轴主线样式。
type AxisLine struct {
	Show      *bool     `json:"show,omitempty"`
	OnZero    *bool     `json:"onZero,omitempty"`
	LineStyle LineStyle `json:"lineStyle,omitempty"`
}

// AxisTick 刻度短线样式。
type AxisTick struct {
	Show      *bool     `json:"show,omitempty"`
	Length    float64   `json:"length,omitempty"`
	Inside    bool      `json:"inside,omitempty"`
	LineStyle LineStyle `json:"lineStyle,omitempty"`
}

// SplitLine 网格分割线。
type SplitLine struct {
	Show      *bool     `json:"show,omitempty"`
	LineStyle LineStyle `json:"lineStyle,omitempty"`
}

// AxisLabel 坐标轴文字。
type AxisLabel struct {
	Show       *bool       `json:"show,omitempty"`
	Color      ColorString `json:"color,omitempty"`
	FontSize   float64     `json:"fontSize,omitempty"`
	FontWeight string      `json:"fontWeight,omitempty"`
	Margin     float64     `json:"margin,omitempty"`
	Rotate     float64     `json:"rotate,omitempty"`
	Formatter  string      `json:"formatter,omitempty"` // 简单支持 "{value}"
	Interval   *int        `json:"interval,omitempty"`  // 0=全部，1=隔一显示一
}

// AxisList 既可以是一个单 axis 也可以是 []axis（ECharts 兼容）。
type AxisList []Axis

func (al *AxisList) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '[' {
		var arr []Axis
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		*al = arr
		return nil
	}
	var single Axis
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*al = []Axis{single}
	return nil
}

// First 返回第一个 axis；不存在时返回零值。
func (al AxisList) First() Axis {
	if len(al) == 0 {
		return Axis{}
	}
	return al[0]
}
