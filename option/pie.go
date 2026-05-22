package option

// PieSeries 饼图/环图 series。
// radius 为 [内径, 外径]，单一字符串如 "60%" 视为外径，环图传 ["40%","70%"]。
type PieSeries struct {
	BaseSeries
	Radius   FlexList `json:"radius,omitempty"`
	Center   FlexList `json:"center,omitempty"`
	RoseType string   `json:"roseType,omitempty"` // ""/"radius"/"area"
	Data     DataList `json:"data,omitempty"`
}

func (s *PieSeries) Kind() SeriesKind { return KindPie }
