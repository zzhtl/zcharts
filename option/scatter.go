package option

// ScatterSeries 散点图 series。
type ScatterSeries struct {
	BaseSeries
	Symbol     string   `json:"symbol,omitempty"`
	SymbolSize float64  `json:"symbolSize,omitempty"`
	Data       DataList `json:"data,omitempty"`
}

func (s *ScatterSeries) Kind() SeriesKind { return KindScatter }
