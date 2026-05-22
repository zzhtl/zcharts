package option

// Grid 对应 ECharts 的 grid 组件——直角坐标系的绘图区。
type Grid struct {
	Show            *bool       `json:"show,omitempty"`
	Left            Flex        `json:"left,omitempty"`
	Top             Flex        `json:"top,omitempty"`
	Right           Flex        `json:"right,omitempty"`
	Bottom          Flex        `json:"bottom,omitempty"`
	Width           Flex        `json:"width,omitempty"`
	Height          Flex        `json:"height,omitempty"`
	ContainLabel    bool        `json:"containLabel,omitempty"`
	BackgroundColor ColorString `json:"backgroundColor,omitempty"`
	BorderColor     ColorString `json:"borderColor,omitempty"`
	BorderWidth     float64     `json:"borderWidth,omitempty"`
}
