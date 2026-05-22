package color

// Palette 是一组按顺序使用的颜色，常用于 series 配色。
type Palette []Color

// At 取第 i 个颜色（循环取模）。当 palette 为空时返回 Transparent。
func (p Palette) At(i int) Color {
	if len(p) == 0 {
		return Transparent
	}
	return p[((i % len(p)) + len(p)) % len(p)]
}

// MustPalette 用一组 hex 字符串构造调色板，遇到非法值直接 panic。
func MustPalette(hexes ...string) Palette {
	out := make(Palette, len(hexes))
	for i, h := range hexes {
		out[i] = MustParse(h)
	}
	return out
}
