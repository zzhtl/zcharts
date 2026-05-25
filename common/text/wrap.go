package text

import (
	"strings"
	"unicode"

	"golang.org/x/image/font"
)

// Wrap 按最大宽度对文本做贪心换行。
//
// 拉丁文按空格切分为词；CJK 文字（中日韩）逐字可断行，行内不插入多余空格。
// 当 face 为 nil 或 maxWidth <= 0 时，原样返回单行。
func Wrap(face font.Face, s string, maxWidth float64) []string {
	if face == nil || maxWidth <= 0 || s == "" {
		return []string{s}
	}
	tokens := tokenize(s)
	if len(tokens) == 0 {
		return []string{s}
	}
	var lines []string
	var cur strings.Builder
	pendingSpace := false
	for _, t := range tokens {
		if t.isSpace {
			if cur.Len() > 0 {
				pendingSpace = true
			}
			continue
		}
		sep := ""
		if pendingSpace {
			sep = " "
		}
		candidate := cur.String() + sep + t.text
		if cur.Len() == 0 || Measure(face, candidate).Width <= maxWidth {
			cur.Reset()
			cur.WriteString(candidate)
		} else {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(t.text)
		}
		pendingSpace = false
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	if len(lines) == 0 {
		return []string{s}
	}
	return lines
}

type wrapToken struct {
	text    string
	isSpace bool
}

// tokenize 把字符串切成可断行的最小单元：空白折叠为一个分隔符、CJK 逐字成词、其余按连续非空白成词。
func tokenize(s string) []wrapToken {
	var toks []wrapToken
	var buf []rune
	flush := func() {
		if len(buf) > 0 {
			toks = append(toks, wrapToken{text: string(buf)})
			buf = buf[:0]
		}
	}
	for _, r := range s {
		switch {
		case unicode.IsSpace(r):
			flush()
			toks = append(toks, wrapToken{text: " ", isSpace: true})
		case isCJK(r):
			flush()
			toks = append(toks, wrapToken{text: string(r)})
		default:
			buf = append(buf, r)
		}
	}
	flush()
	return toks
}

// isCJK 判断一个 rune 是否为可逐字断行的 CJK 字符（含 CJK 标点）。
func isCJK(r rune) bool {
	switch {
	case unicode.Is(unicode.Han, r),
		unicode.Is(unicode.Hiragana, r),
		unicode.Is(unicode.Katakana, r),
		unicode.Is(unicode.Hangul, r):
		return true
	case r >= 0x3000 && r <= 0x303F: // CJK 符号与标点
		return true
	case r >= 0xFF00 && r <= 0xFFEF: // 全角字符
		return true
	}
	return false
}
