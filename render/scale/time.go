package scale

import (
	"time"

	"github.com/zzhtl/zcharts/common/number"
)

// Time 是毫秒时间戳轴，数据值按 Unix milliseconds 映射。
type Time struct {
	dataLo, dataHi float64
	targetTicks    int
	pxLo, pxHi     float64
}

// NewTime 根据原始 [min,max] 构造时间轴。
func NewTime(min, max float64, targetTicks int) *Time {
	if targetTicks <= 0 {
		targetTicks = 5
	}
	if max <= min {
		max = min + float64(24*time.Hour/time.Millisecond)
	}
	return &Time{dataLo: min, dataHi: max, targetTicks: targetTicks}
}

func (s *Time) SetPixelRange(lo, hi float64) { s.pxLo, s.pxHi = lo, hi }

func (s *Time) Pixel(v float64) float64 {
	return number.Map(v, s.dataLo, s.dataHi, s.pxLo, s.pxHi)
}

func (s *Time) Range() (float64, float64) { return s.dataLo, s.dataHi }

func (s *Time) Ticks() []Tick {
	count := s.targetTicks
	if count < 2 {
		count = 2
	}
	out := make([]Tick, count)
	step := (s.dataHi - s.dataLo) / float64(count-1)
	for i := range out {
		v := s.dataLo + step*float64(i)
		out[i] = Tick{Value: v, Label: formatTimeLabel(v, s.dataHi-s.dataLo)}
	}
	return out
}

func formatTimeLabel(v, span float64) string {
	t := time.UnixMilli(int64(v)).UTC()
	switch {
	case span >= float64(365*24*time.Hour/time.Millisecond):
		return t.Format("2006-01")
	case span >= float64(2*24*time.Hour/time.Millisecond):
		return t.Format("01-02")
	default:
		return t.Format("15:04")
	}
}
