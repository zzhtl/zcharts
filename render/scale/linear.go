package scale

import (
	"github.com/zzhtl/zcharts/common/number"
)

// Linear 是线性数值轴。
type Linear struct {
	dataLo, dataHi float64 // 数据值范围（含 nice 扩展后）
	step           float64
	pxLo, pxHi     float64
}

// NewLinear 根据原始 [min,max] 构造线性轴。
// 当 nice=true 时会把范围扩展到漂亮的整数刻度。
func NewLinear(min, max float64, targetTicks int, nice bool) *Linear {
	lo, hi, step := number.NiceRange(min, max, targetTicks, nice)
	return &Linear{dataLo: lo, dataHi: hi, step: step}
}

// NewLinearFixed 直接指定 [lo, hi] 区间（不做 nice 扩展），用于用户显式指定 min/max。
func NewLinearFixed(lo, hi float64, targetTicks int) *Linear {
	_, _, step := number.NiceRange(lo, hi, targetTicks, false)
	return &Linear{dataLo: lo, dataHi: hi, step: step}
}

func (s *Linear) SetPixelRange(lo, hi float64) { s.pxLo, s.pxHi = lo, hi }

func (s *Linear) Pixel(v float64) float64 {
	return number.Map(v, s.dataLo, s.dataHi, s.pxLo, s.pxHi)
}

func (s *Linear) Range() (float64, float64) { return s.dataLo, s.dataHi }

func (s *Linear) Ticks() []Tick {
	values := number.Ticks(s.dataLo, s.dataHi, s.step)
	out := make([]Tick, len(values))
	for i, v := range values {
		out[i] = Tick{Value: v, Label: number.FormatAuto(v)}
	}
	return out
}
