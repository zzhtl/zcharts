package text

import (
	"strings"

	"golang.org/x/image/font"
)

// Wrap 按最大宽度对文本做贪心换行（按空格切分，超长单字符不切）。
// 当 face 为 nil 或 maxWidth <= 0 时，原样返回单行。
func Wrap(face font.Face, s string, maxWidth float64) []string {
	if face == nil || maxWidth <= 0 || s == "" {
		return []string{s}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{s}
	}
	var lines []string
	var cur string
	for _, w := range words {
		candidate := w
		if cur != "" {
			candidate = cur + " " + w
		}
		if Measure(face, candidate).Width <= maxWidth || cur == "" {
			cur = candidate
			continue
		}
		lines = append(lines, cur)
		cur = w
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
