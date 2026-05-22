package scale

// Category 是分类轴。数据值用 float64 表示 category 索引（0..n-1）。
type Category struct {
	labels      []string
	boundaryGap bool // true: 把每个类目当作"区间"，刻度在区间中点；false: 刻度在轴端点
	pxLo, pxHi  float64
}

// NewCategory 创建分类轴。boundaryGap=true 适合柱状图；false 适合折线图。
func NewCategory(labels []string, boundaryGap bool) *Category {
	return &Category{labels: labels, boundaryGap: boundaryGap}
}

func (s *Category) SetPixelRange(lo, hi float64) { s.pxLo, s.pxHi = lo, hi }

func (s *Category) Pixel(v float64) float64 {
	n := len(s.labels)
	if n == 0 {
		return s.pxLo
	}
	width := s.pxHi - s.pxLo
	if s.boundaryGap {
		// 每个 category 占据宽度 width/n，刻度在区间中点
		band := width / float64(n)
		return s.pxLo + band*(v+0.5)
	}
	// 不留间隙：v=0 在 pxLo，v=n-1 在 pxHi
	if n == 1 {
		return s.pxLo + width/2
	}
	return s.pxLo + width*v/float64(n-1)
}

func (s *Category) Range() (float64, float64) { return 0, float64(len(s.labels) - 1) }

func (s *Category) Ticks() []Tick {
	out := make([]Tick, len(s.labels))
	for i, l := range s.labels {
		out[i] = Tick{Value: float64(i), Label: l}
	}
	return out
}

// BandWidth 返回每个 category 占据的像素宽度（柱状图柱宽计算用）。
func (s *Category) BandWidth() float64 {
	n := len(s.labels)
	if n == 0 {
		return 0
	}
	return (s.pxHi - s.pxLo) / float64(n)
}

// Labels 返回原始类目标签数组（拷贝）。
func (s *Category) Labels() []string {
	out := make([]string, len(s.labels))
	copy(out, s.labels)
	return out
}
