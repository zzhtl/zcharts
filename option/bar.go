package option

// BarSeries 柱状图 series。
type BarSeries struct {
	BaseSeries
	Stack           string   `json:"stack,omitempty"`
	BarWidth        Flex     `json:"barWidth,omitempty"`
	BarMaxWidth     Flex     `json:"barMaxWidth,omitempty"`
	BarMinWidth     Flex     `json:"barMinWidth,omitempty"`
	BarGap          string   `json:"barGap,omitempty"`          // 同坐标的多 series 之间间距，如 "30%"
	BarCategoryGap  string   `json:"barCategoryGap,omitempty"`  // 不同 category 间距，默认 "20%"
	Data            DataList `json:"data,omitempty"`
}

func (s *BarSeries) Kind() SeriesKind { return KindBar }
