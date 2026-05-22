// Package color 提供 zcharts 内部使用的颜色类型与解析能力。
//
// 颜色统一以 RGBA（每通道 0~255，alpha 0~255）保存。支持从以下字符串解析：
//   - "#RGB" / "#RRGGBB" / "#RRGGBBAA"
//   - "rgb(r, g, b)"
//   - "rgba(r, g, b, a)"  其中 a 为 0~1 浮点
//   - "transparent" / "none" / ""  → 透明
package color

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Color 表示一个 RGBA 颜色。
type Color struct {
	R, G, B, A uint8
}

// Transparent 完全透明。
var Transparent = Color{}

// RGBA 实现 image/color.Color 接口。
func (c Color) RGBA() (r, g, b, a uint32) {
	return uint32(c.R) * 257, uint32(c.G) * 257, uint32(c.B) * 257, uint32(c.A) * 257
}

// NRGBA 转为标准库的 NRGBA，便于和 Go 图像 API 互操作。
func (c Color) NRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// Hex 输出 #RRGGBB 或 #RRGGBBAA。
func (c Color) Hex() string {
	if c.A == 0xFF {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}

// IsZero 表示颜色未设置（结构体零值，等同于完全透明）。
func (c Color) IsZero() bool { return c == Color{} }

// WithAlpha 返回一个新的颜色，alpha 通道被替换为 a（0~255）。
func (c Color) WithAlpha(a uint8) Color {
	c.A = a
	return c
}

// RGB 用 0~255 数值快速构造一个不透明颜色。
func RGB(r, g, b uint8) Color { return Color{R: r, G: g, B: b, A: 0xFF} }

// RGBA 用 0~255 数值快速构造一个带 alpha 的颜色。
func RGBA(r, g, b, a uint8) Color { return Color{R: r, G: g, B: b, A: a} }

// MustParse 等同 Parse，解析失败时 panic。仅用于常量初始化。
func MustParse(s string) Color {
	c, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return c
}

// Parse 解析颜色字符串，规则参见包注释。
func Parse(s string) (Color, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "transparent") || strings.EqualFold(s, "none") {
		return Transparent, nil
	}
	if strings.HasPrefix(s, "#") {
		return parseHex(s[1:])
	}
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "rgba("):
		return parseRGBA(s[len("rgba("):])
	case strings.HasPrefix(lower, "rgb("):
		return parseRGB(s[len("rgb("):])
	}
	return Color{}, fmt.Errorf("zcharts/color: invalid color %q", s)
}

func parseHex(s string) (Color, error) {
	switch len(s) {
	case 3: // #RGB
		r, g, b := nibble(s[0]), nibble(s[1]), nibble(s[2])
		if r < 0 || g < 0 || b < 0 {
			return Color{}, fmt.Errorf("zcharts/color: invalid hex %q", s)
		}
		return Color{R: uint8(r*17), G: uint8(g*17), B: uint8(b*17), A: 0xFF}, nil
	case 6: // #RRGGBB
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return Color{}, fmt.Errorf("zcharts/color: invalid hex %q", s)
		}
		return Color{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}, nil
	case 8: // #RRGGBBAA
		v, err := strconv.ParseUint(s, 16, 64)
		if err != nil {
			return Color{}, fmt.Errorf("zcharts/color: invalid hex %q", s)
		}
		return Color{R: uint8(v >> 24), G: uint8(v >> 16), B: uint8(v >> 8), A: uint8(v)}, nil
	}
	return Color{}, fmt.Errorf("zcharts/color: invalid hex length %q", s)
}

func nibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func parseRGB(s string) (Color, error) {
	s = strings.TrimSuffix(strings.TrimSpace(s), ")")
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return Color{}, fmt.Errorf("zcharts/color: rgb expects 3 components, got %d", len(parts))
	}
	r, err1 := parseByte(parts[0])
	g, err2 := parseByte(parts[1])
	b, err3 := parseByte(parts[2])
	if err := firstErr(err1, err2, err3); err != nil {
		return Color{}, err
	}
	return Color{R: r, G: g, B: b, A: 0xFF}, nil
}

func parseRGBA(s string) (Color, error) {
	s = strings.TrimSuffix(strings.TrimSpace(s), ")")
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return Color{}, fmt.Errorf("zcharts/color: rgba expects 4 components, got %d", len(parts))
	}
	r, err1 := parseByte(parts[0])
	g, err2 := parseByte(parts[1])
	b, err3 := parseByte(parts[2])
	af, err4 := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
	if err := firstErr(err1, err2, err3, err4); err != nil {
		return Color{}, err
	}
	if af < 0 {
		af = 0
	} else if af > 1 {
		af = 1
	}
	return Color{R: r, G: g, B: b, A: uint8(af*255 + 0.5)}, nil
}

func parseByte(s string) (uint8, error) {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("zcharts/color: invalid component %q", s)
	}
	if v < 0 {
		v = 0
	} else if v > 255 {
		v = 255
	}
	return uint8(v), nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
