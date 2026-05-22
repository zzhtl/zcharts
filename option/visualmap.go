package option

// VisualMap 视觉映射组件（热力图必需）。
// 第一阶段只实现连续型映射（continuous），按 [min,max] 区间在 inRange.color 中插值。
type VisualMap struct {
	Type       string    `json:"type,omitempty"` // "continuous"/"piecewise"
	Min        float64   `json:"min,omitempty"`
	Max        float64   `json:"max,omitempty"`
	Show       *bool     `json:"show,omitempty"`
	Left       Flex      `json:"left,omitempty"`
	Top        Flex      `json:"top,omitempty"`
	Right      Flex      `json:"right,omitempty"`
	Bottom     Flex      `json:"bottom,omitempty"`
	Orient     string    `json:"orient,omitempty"` // "horizontal"/"vertical"，默认 vertical
	Calculable bool      `json:"calculable,omitempty"`
	Text       []string  `json:"text,omitempty"`
	InRange    InRange   `json:"inRange,omitempty"`
}

// InRange 视觉通道映射（当前仅支持 color）。
type InRange struct {
	Color []string `json:"color,omitempty"`
}
