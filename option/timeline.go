package option

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// TimelineSeries 时间轴图，按时间线从上到下展示事件。
type TimelineSeries struct {
	BaseSeries
	Left      Flex         `json:"left,omitempty"`
	Top       Flex         `json:"top,omitempty"`
	Width     Flex         `json:"width,omitempty"`
	Height    Flex         `json:"height,omitempty"`
	LineStyle LineStyle    `json:"lineStyle,omitempty"`
	ItemStyle ItemStyle    `json:"itemStyle,omitempty"`
	Label     Label        `json:"label,omitempty"`
	Data      TimelineData `json:"data,omitempty"`
}

func (s *TimelineSeries) Kind() SeriesKind { return KindTimeline }

// TimelineData 是时间轴事件列表。
type TimelineData []TimelineEvent

// TimelineEvent 描述时间轴上的单个事件。
type TimelineEvent struct {
	Time      string          `json:"time,omitempty"`
	Title     string          `json:"title,omitempty"`
	Content   string          `json:"content,omitempty"`
	Name      string          `json:"name,omitempty"`
	ItemStyle ItemStyle       `json:"itemStyle,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

func (e *TimelineEvent) UnmarshalJSON(data []byte) error {
	e.Raw = append(e.Raw[:0], data...)
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	switch data[0] {
	case '"':
		e.Title = scalarText(data)
		return nil
	case '[':
		var arr []json.RawMessage
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		if len(arr) > 0 {
			e.Time = scalarText(arr[0])
		}
		if len(arr) > 1 {
			e.Title = scalarText(arr[1])
		}
		if len(arr) > 2 {
			e.Content = scalarText(arr[2])
		}
		return nil
	case '{':
		var obj struct {
			Time        string          `json:"time"`
			Date        string          `json:"date"`
			Title       string          `json:"title"`
			Name        string          `json:"name"`
			Content     string          `json:"content"`
			Description string          `json:"description"`
			Desc        string          `json:"desc"`
			Value       json.RawMessage `json:"value"`
			ItemStyle   ItemStyle       `json:"itemStyle"`
		}
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		e.Time = firstString(obj.Time, obj.Date)
		e.Title = firstString(obj.Title, obj.Name)
		e.Name = obj.Name
		e.Content = firstString(obj.Content, obj.Description, obj.Desc)
		if e.Content == "" && len(obj.Value) > 0 {
			e.Content = scalarText(obj.Value)
		}
		if e.Title == "" && e.Content != "" {
			e.Title = e.Content
			e.Content = ""
		}
		e.ItemStyle = obj.ItemStyle
		return nil
	}
	e.Title = scalarText(data)
	return nil
}

func firstString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func scalarText(data []byte) string {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return s
	}
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return ""
}
