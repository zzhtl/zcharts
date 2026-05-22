package number

import "math"

// NiceRange 把 [min,max] 扩展到一个"漂亮"的范围（用于坐标轴 min/max）。
// expand=true 时把端点扩到刻度的整数倍。
func NiceRange(min, max float64, targetTicks int, expand bool) (lo, hi, step float64) {
	if targetTicks <= 0 {
		targetTicks = 5
	}
	if math.IsNaN(min) || math.IsNaN(max) || min == max {
		if min == 0 && max == 0 {
			return 0, 1, 0.2
		}
		span := math.Abs(min) * 0.5
		if span == 0 {
			span = 1
		}
		min, max = min-span, max+span
	}
	if min > max {
		min, max = max, min
	}
	step = niceStep((max-min)/float64(targetTicks), true)
	if expand {
		lo = math.Floor(min/step) * step
		hi = math.Ceil(max/step) * step
	} else {
		lo, hi = min, max
	}
	return lo, hi, step
}

// Ticks 返回 [lo,hi] 之间步长 step 的刻度数组（包含端点上落到刻度上的值）。
// 内部会做浮点容差处理。
func Ticks(lo, hi, step float64) []float64 {
	if step <= 0 || hi <= lo {
		return nil
	}
	const eps = 1e-9
	start := math.Ceil(lo/step - eps)
	end := math.Floor(hi/step + eps)
	n := int(end - start + 1)
	if n <= 0 {
		return nil
	}
	out := make([]float64, n)
	for i := range n {
		v := (start + float64(i)) * step
		// 浮点误差消除
		if math.Abs(v) < step*1e-10 {
			v = 0
		}
		out[i] = v
	}
	return out
}

// niceStep 返回不小于（或不大于）给定 v 的"漂亮"步长（1/2/2.5/5 × 10^n）。
// round=true 时选择最接近的；false 时选择不小于的。
func niceStep(v float64, round bool) float64 {
	if v <= 0 {
		return 1
	}
	exp := math.Floor(math.Log10(v))
	frac := v / math.Pow(10, exp)
	var nice float64
	if round {
		switch {
		case frac < 1.5:
			nice = 1
		case frac < 3:
			nice = 2
		case frac < 7:
			nice = 5
		default:
			nice = 10
		}
	} else {
		switch {
		case frac <= 1:
			nice = 1
		case frac <= 2:
			nice = 2
		case frac <= 5:
			nice = 5
		default:
			nice = 10
		}
	}
	return nice * math.Pow(10, exp)
}
