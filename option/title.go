package option

// Title 对齐 ECharts 的 title 组件（仅核心字段）。
type Title struct {
	Show         bool      `json:"show,omitempty"`
	Text         string    `json:"text,omitempty"`
	Subtext      string    `json:"subtext,omitempty"`
	Left         Flex      `json:"left,omitempty"`
	Top          Flex      `json:"top,omitempty"`
	Right        Flex      `json:"right,omitempty"`
	Bottom       Flex      `json:"bottom,omitempty"`
	TextAlign    string    `json:"textAlign,omitempty"`
	TextStyle    TextStyle `json:"textStyle,omitempty"`
	SubtextStyle TextStyle `json:"subtextStyle,omitempty"`
	ItemGap      float64   `json:"itemGap,omitempty"`
}
