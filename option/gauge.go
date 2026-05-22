package option

// GaugeSeries 仪表盘 series。
type GaugeSeries struct {
	BaseSeries
	Min         float64  `json:"min,omitempty"`
	Max         float64  `json:"max,omitempty"`
	StartAngle  *float64 `json:"startAngle,omitempty"` // 默认 225
	EndAngle    *float64 `json:"endAngle,omitempty"`   // 默认 -45
	Radius      Flex     `json:"radius,omitempty"`
	Center      FlexList `json:"center,omitempty"`
	SplitNumber int      `json:"splitNumber,omitempty"`
	Data        DataList `json:"data,omitempty"`
}

func (s *GaugeSeries) Kind() SeriesKind { return KindGauge }
