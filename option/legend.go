package option

// Legend 对齐 ECharts 的 legend 组件。
type Legend struct {
	Show      *bool     `json:"show,omitempty"`
	Data      Strings   `json:"data,omitempty"`
	Left      Flex      `json:"left,omitempty"`
	Top       Flex      `json:"top,omitempty"`
	Right     Flex      `json:"right,omitempty"`
	Bottom    Flex      `json:"bottom,omitempty"`
	Orient    string    `json:"orient,omitempty"` // "horizontal" / "vertical"
	ItemGap   float64   `json:"itemGap,omitempty"`
	ItemWidth float64   `json:"itemWidth,omitempty"`
	ItemHeight float64  `json:"itemHeight,omitempty"`
	TextStyle TextStyle `json:"textStyle,omitempty"`
}

// IsShown 处理三态 show（未设置=true，true=true，false=false）。
func (l *Legend) IsShown() bool {
	if l == nil {
		return false
	}
	if l.Show == nil {
		return true
	}
	return *l.Show
}
