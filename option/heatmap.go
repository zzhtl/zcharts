package option

// HeatmapSeries 热力图 series（依赖 visualMap 配色）。
// 数据格式：[[x_index, y_index, value], ...]
type HeatmapSeries struct {
	BaseSeries
	Data DataList `json:"data,omitempty"`
}

func (s *HeatmapSeries) Kind() SeriesKind { return KindHeatmap }
