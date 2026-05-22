package option

// Radar 雷达组件（与 series.radar 配合）。
type Radar struct {
	Shape     string         `json:"shape,omitempty"` // "polygon"/"circle"
	Indicator []RadarIndicator `json:"indicator,omitempty"`
	Center    FlexList       `json:"center,omitempty"`
	Radius    Flex           `json:"radius,omitempty"`
	StartAngle float64       `json:"startAngle,omitempty"`
	SplitNumber int          `json:"splitNumber,omitempty"`
	AxisName   TextStyle     `json:"axisName,omitempty"`
	SplitLine  SplitLine     `json:"splitLine,omitempty"`
	SplitArea  AreaStyle     `json:"splitArea,omitempty"`
	AxisLine   LineStyle     `json:"axisLine,omitempty"`
}

// RadarIndicator 维度定义。
type RadarIndicator struct {
	Name string  `json:"name"`
	Max  float64 `json:"max,omitempty"`
	Min  float64 `json:"min,omitempty"`
}

// RadarSeries 雷达图 series（依赖顶层 radar 组件）。
type RadarSeries struct {
	BaseSeries
	LineStyle LineStyle  `json:"lineStyle,omitempty"`
	AreaStyle *AreaStyle `json:"areaStyle,omitempty"`
	Symbol    string     `json:"symbol,omitempty"`
	SymbolSize float64   `json:"symbolSize,omitempty"`
	Data      []RadarDataItem `json:"data,omitempty"`
}

// RadarDataItem 雷达图数据：每项是一组数值 + 名称。
type RadarDataItem struct {
	Name      string      `json:"name,omitempty"`
	Value     []float64   `json:"value"`
	LineStyle LineStyle   `json:"lineStyle,omitempty"`
	AreaStyle *AreaStyle  `json:"areaStyle,omitempty"`
	ItemStyle ItemStyle   `json:"itemStyle,omitempty"`
}

func (s *RadarSeries) Kind() SeriesKind { return KindRadar }
