package option

// Option 是 ECharts 顶层配置的 Go 镜像。
// 字段命名严格对齐 ECharts 5.x；未识别字段会被 encoding/json 忽略。
type Option struct {
	Title           *Title      `json:"title,omitempty"`
	Legend          *Legend     `json:"legend,omitempty"`
	Tooltip         *Tooltip    `json:"tooltip,omitempty"`
	Grid            *Grid       `json:"grid,omitempty"`
	XAxis           AxisList    `json:"xAxis,omitempty"`
	YAxis           AxisList    `json:"yAxis,omitempty"`
	Color           ColorList   `json:"color,omitempty"`
	BackgroundColor ColorString `json:"backgroundColor,omitempty"`
	Series          SeriesList  `json:"series,omitempty"`
	VisualMap       *VisualMap  `json:"visualMap,omitempty"`
	Radar           *Radar      `json:"radar,omitempty"`
	Animation       *bool       `json:"animation,omitempty"`
}
