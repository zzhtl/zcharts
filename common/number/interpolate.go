package number

// Lerp 在 [a,b] 上按 t∈[0,1] 线性插值。
func Lerp(a, b, t float64) float64 { return a + (b-a)*t }

// Map 把 v 从 [a0,a1] 线性映射到 [b0,b1]，会按比例外推。
func Map(v, a0, a1, b0, b1 float64) float64 {
	if a1 == a0 {
		return (b0 + b1) / 2
	}
	return b0 + (v-a0)*(b1-b0)/(a1-a0)
}

// Clamp 把 v 限制到 [lo,hi] 内。
func Clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
