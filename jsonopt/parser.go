// Package jsonopt 把 ECharts 风格的 JSON option 解析为 option.Option。
//
// 入口：Parse([]byte) (*option.Option, error)
//
// 未识别的字段会被静默丢弃（不会报错），保证向前兼容 ECharts 的字段扩展。
package jsonopt

import (
	"encoding/json"
	"fmt"

	"github.com/zzhtl/zcharts/option"
)

// Parse 解析 ECharts JSON option。
func Parse(data []byte) (*option.Option, error) {
	opt := &option.Option{}
	if err := json.Unmarshal(data, opt); err != nil {
		return nil, fmt.Errorf("jsonopt: %w", err)
	}
	return opt, nil
}

// ParseString 是 Parse 的字符串便捷形式。
func ParseString(s string) (*option.Option, error) { return Parse([]byte(s)) }
