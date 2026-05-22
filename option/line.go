package option

// LineSeries 折线图 series。
type LineSeries struct {
	BaseSeries
	Smooth     bool      `json:"smooth,omitempty"`
	Stack      string    `json:"stack,omitempty"`
	Symbol     string    `json:"symbol,omitempty"`     // "circle"/"emptyCircle"/"none"/...
	SymbolSize float64   `json:"symbolSize,omitempty"`
	ShowSymbol *bool     `json:"showSymbol,omitempty"`
	LineStyle  LineStyle `json:"lineStyle,omitempty"`
	AreaStyle  *AreaStyle `json:"areaStyle,omitempty"`
	Data       DataList  `json:"data,omitempty"`
}

func (s *LineSeries) Kind() SeriesKind { return KindLine }
