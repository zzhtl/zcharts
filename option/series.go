package option

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// SeriesKind 是 series.type 字段的取值。
type SeriesKind string

const (
	KindLine      SeriesKind = "line"
	KindBar       SeriesKind = "bar"
	KindPie       SeriesKind = "pie"
	KindScatter   SeriesKind = "scatter"
	KindRadar     SeriesKind = "radar"
	KindHeatmap   SeriesKind = "heatmap"
	KindGauge     SeriesKind = "gauge"
	KindFunnel    SeriesKind = "funnel"
	KindTimeline  SeriesKind = "timeline"
	KindWordCloud SeriesKind = "wordCloud"
)

// Series 是所有 series 子类型实现的接口。
type Series interface {
	Kind() SeriesKind
	GetName() string
	GetXAxisIndex() int
	GetYAxisIndex() int
}

// BaseSeries 提供 series 公共字段，可被各类型 series 嵌入。
type BaseSeries struct {
	Type       string      `json:"type"`
	Name       string      `json:"name,omitempty"`
	Color      ColorString `json:"color,omitempty"`
	ItemStyle  ItemStyle   `json:"itemStyle,omitempty"`
	Label      Label       `json:"label,omitempty"`
	ZLevel     int         `json:"zlevel,omitempty"`
	XAxisIndex int         `json:"xAxisIndex,omitempty"`
	YAxisIndex int         `json:"yAxisIndex,omitempty"`
}

func (b BaseSeries) GetName() string    { return b.Name }
func (b BaseSeries) GetXAxisIndex() int { return b.XAxisIndex }
func (b BaseSeries) GetYAxisIndex() int { return b.YAxisIndex }

// SeriesList 是 series 数组的容器，按 type 字段分发到具体类型。
type SeriesList []Series

func (sl *SeriesList) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	// 兼容单个 series 对象的写法
	if data[0] == '{' {
		raw := json.RawMessage(data)
		s, err := decodeSeries(raw)
		if err != nil {
			return err
		}
		*sl = append(*sl, s)
		return nil
	}
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return err
	}
	for i, raw := range raws {
		s, err := decodeSeries(raw)
		if err != nil {
			return fmt.Errorf("option/series[%d]: %w", i, err)
		}
		*sl = append(*sl, s)
	}
	return nil
}

func decodeSeries(raw json.RawMessage) (Series, error) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	switch SeriesKind(head.Type) {
	case KindLine, "":
		var s LineSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindBar:
		var s BarSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindPie:
		var s PieSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindScatter:
		var s ScatterSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindRadar:
		var s RadarSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindHeatmap:
		var s HeatmapSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindGauge:
		var s GaugeSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindFunnel:
		var s FunnelSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindTimeline:
		var s TimelineSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	case KindWordCloud:
		var s WordCloudSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	default:
		var s CustomSeries
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	}
}

// DataValue 是 ECharts 的多态数据点，支持 number / string / [num,num,...] / {value, name}。
type DataValue struct {
	Values []float64       // 数值通道（1 维或多维）
	Name   string          // 可选的名称（pie/gauge 用）
	Label  string          // 可选的 category 标签（line/bar 用，作为字符串数据时）
	Raw    json.RawMessage // 原始 JSON，用于其他渲染器自取字段
}

// Number 返回首个数值（多数 series 取此值）。
func (d DataValue) Number() float64 {
	if len(d.Values) > 0 {
		return d.Values[0]
	}
	return 0
}

// Pair 返回前两个数值，缺失补 0（scatter 用）。
func (d DataValue) Pair() (float64, float64) {
	switch len(d.Values) {
	case 0:
		return 0, 0
	case 1:
		return d.Values[0], 0
	default:
		return d.Values[0], d.Values[1]
	}
}

func (d *DataValue) UnmarshalJSON(data []byte) error {
	d.Raw = append(d.Raw[:0], data...)
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	switch data[0] {
	case '"':
		// 字符串：当作 category label 也作为名称
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		d.Label = s
		d.Name = s
		return nil
	case '[':
		var arr []json.RawMessage
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		for _, el := range arr {
			el = bytes.TrimSpace(el)
			if len(el) == 0 {
				continue
			}
			if el[0] == '"' {
				var s string
				_ = json.Unmarshal(el, &s)
				if d.Label == "" {
					d.Label = s
				}
				if v, ok := parseTimeNumber(s); ok {
					d.Values = append(d.Values, v)
				}
				continue
			}
			var v float64
			if err := json.Unmarshal(el, &v); err == nil {
				d.Values = append(d.Values, v)
			}
		}
		return nil
	case '{':
		var obj struct {
			Value json.RawMessage `json:"value"`
			Name  string          `json:"name"`
		}
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		d.Name = obj.Name
		if len(obj.Value) > 0 {
			vd := &DataValue{}
			if err := vd.UnmarshalJSON(obj.Value); err == nil {
				d.Values = vd.Values
				if d.Label == "" {
					d.Label = vd.Label
				}
			}
		}
		return nil
	}
	// number
	var v float64
	if err := json.Unmarshal(data, &v); err == nil {
		d.Values = []float64{v}
	}
	return nil
}

func parseTimeNumber(s string) (float64, bool) {
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return float64(t.UnixMilli()), true
		}
	}
	return 0, false
}

// DataList 是 [DataValue]，data 数组的载体。
type DataList []DataValue

// AsValues 把每个 DataValue 的首数值提取出来，供折线/柱状图使用。
func (dl DataList) AsValues() []float64 {
	out := make([]float64, len(dl))
	for i, v := range dl {
		out[i] = v.Number()
	}
	return out
}
