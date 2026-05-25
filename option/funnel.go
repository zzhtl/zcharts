package option

// FunnelSeries 漏斗图 series。
type FunnelSeries struct {
	BaseSeries
	Left   Flex `json:"left,omitempty"`
	Top    Flex `json:"top,omitempty"`
	Width  Flex `json:"width,omitempty"`
	Height Flex `json:"height,omitempty"`

	Min  float64 `json:"min,omitempty"`
	Max  float64 `json:"max,omitempty"`
	Sort string  `json:"sort,omitempty"` // "descending"/"ascending"/"none"
	Gap  float64 `json:"gap,omitempty"`

	Data DataList `json:"data,omitempty"`
}

func (s *FunnelSeries) Kind() SeriesKind { return KindFunnel }
