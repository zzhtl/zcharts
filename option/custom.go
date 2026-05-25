package option

import "encoding/json"

// CustomSeries 保留未内置支持的 series.type，供用户通过自定义渲染器处理。
type CustomSeries struct {
	BaseSeries
	CoordinateSystem string          `json:"coordinateSystem,omitempty"`
	Data             DataList        `json:"data,omitempty"`
	Raw              json.RawMessage `json:"-"`
}

func (s *CustomSeries) Kind() SeriesKind { return SeriesKind(s.Type) }

func (s *CustomSeries) UnmarshalJSON(data []byte) error {
	type alias CustomSeries
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*s = CustomSeries(out)
	s.Raw = append(s.Raw[:0], data...)
	return nil
}
