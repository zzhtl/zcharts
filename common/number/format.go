// Package number 提供数值格式化、刻度计算、插值等通用工具。
package number

import (
	"math"
	"strconv"
	"strings"
)

// FormatFloat 以指定小数位输出，自动去掉无意义的尾随 0。
// 例如 FormatFloat(1.2300, 4) = "1.23"，FormatFloat(100, 2) = "100"。
func FormatFloat(v float64, maxDecimals int) string {
	if math.IsNaN(v) {
		return "NaN"
	}
	if math.IsInf(v, 0) {
		if v > 0 {
			return "Inf"
		}
		return "-Inf"
	}
	s := strconv.FormatFloat(v, 'f', maxDecimals, 64)
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}

// FormatAuto 按数值规模自动选择小数位（绝对值越小，精度越高）。
func FormatAuto(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 100:
		return FormatFloat(v, 0)
	case abs >= 10:
		return FormatFloat(v, 1)
	case abs >= 1:
		return FormatFloat(v, 2)
	case abs >= 0.01:
		return FormatFloat(v, 3)
	default:
		return FormatFloat(v, 6)
	}
}

// FormatSI 用 SI 后缀（K/M/G/T）压缩大数。
func FormatSI(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 1e12:
		return FormatFloat(v/1e12, 1) + "T"
	case abs >= 1e9:
		return FormatFloat(v/1e9, 1) + "G"
	case abs >= 1e6:
		return FormatFloat(v/1e6, 1) + "M"
	case abs >= 1e3:
		return FormatFloat(v/1e3, 1) + "K"
	}
	return FormatAuto(v)
}

// FormatPercent 以百分比形式输出（如 0.123 → "12.3%"）。
func FormatPercent(v float64, decimals int) string {
	return FormatFloat(v*100, decimals) + "%"
}
