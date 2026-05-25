package option

// WordCloudSeries 词云图 series，兼容常见 echarts-wordcloud 配置字段。
type WordCloudSeries struct {
	BaseSeries
	Shape         string    `json:"shape,omitempty"`
	Left          Flex      `json:"left,omitempty"`
	Top           Flex      `json:"top,omitempty"`
	Width         Flex      `json:"width,omitempty"`
	Height        Flex      `json:"height,omitempty"`
	SizeRange     []float64 `json:"sizeRange,omitempty"`
	RotationRange []float64 `json:"rotationRange,omitempty"`
	GridSize      float64   `json:"gridSize,omitempty"`
	TextStyle     TextStyle `json:"textStyle,omitempty"`
	Data          DataList  `json:"data,omitempty"`
}

func (s *WordCloudSeries) Kind() SeriesKind { return KindWordCloud }
