package series

import (
	"math"
	"sort"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
)

// DrawWordCloudArgs WordCloud 渲染参数。
type DrawWordCloudArgs struct {
	Canvas  zcanvas.Canvas
	Series  *option.WordCloudSeries
	Bounds  geom.Rect
	Palette color.Palette
	Family  string
}

// DrawWordCloud 用确定性螺旋布局绘制词云。
func DrawWordCloud(a DrawWordCloudArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 {
		return
	}
	rect := wordCloudRect(a.Bounds, a.Series)
	if rect.IsZero() {
		return
	}

	items := make([]option.DataValue, len(a.Series.Data))
	copy(items, a.Series.Data)
	sort.SliceStable(items, func(i, j int) bool { return items[i].Number() > items[j].Number() })

	minValue, maxValue := wordCloudValueRange(items)
	minSize, maxSize := wordCloudSizeRange(a.Series.SizeRange)
	grid := a.Series.GridSize
	if grid <= 0 {
		grid = 8
	}

	placed := make([]geom.Rect, 0, len(items))
	for i, d := range items {
		word := d.Name
		if word == "" {
			word = d.Label
		}
		if word == "" {
			continue
		}
		size := minSize
		if maxValue > minValue {
			size = minSize + (maxSize-minSize)*(d.Number()-minValue)/(maxValue-minValue)
		}
		style := zcanvas.TextStyle{
			Family:   pickWordCloudFamily(a.Family, a.Series.TextStyle.FontFamily),
			Color:    wordColor(a.Series, a.Palette, i),
			Weight:   a.Series.TextStyle.FontWeight,
			Anchor:   zcanvas.AnchorMiddle,
			VAlign:   zcanvas.AlignMiddle,
			Rotation: wordRotation(a.Series.RotationRange, i),
		}
		minPlacedSize := math.Max(8, minSize*0.72)
		for trySize := size; trySize >= minPlacedSize; trySize *= 0.9 {
			style.Size = trySize
			w, h, _ := a.Canvas.MeasureText(word, style)
			if w <= 0 || h <= 0 {
				break
			}
			if pos, box, ok := placeWord(rect, placed, w, h, style.Rotation, grid, i); ok {
				a.Canvas.DrawText(pos.X, pos.Y, word, style)
				placed = append(placed, box)
				break
			}
		}
	}
}

func wordCloudRect(bounds geom.Rect, s *option.WordCloudSeries) geom.Rect {
	x := bounds.X + bounds.W*0.08
	y := bounds.Y + bounds.H*0.06
	w := bounds.W * 0.84
	h := bounds.H * 0.82
	if s.Width.Set {
		w = s.Width.Resolve(bounds.W, w)
	}
	if s.Height.Set {
		h = s.Height.Resolve(bounds.H, h)
	}
	if s.Left.Set {
		if s.Left.Keyword == "center" {
			x = bounds.X + (bounds.W-w)/2
		} else {
			x = bounds.X + s.Left.Resolve(bounds.W, x-bounds.X)
		}
	}
	if s.Top.Set {
		if s.Top.Keyword == "middle" || s.Top.Keyword == "center" {
			y = bounds.Y + (bounds.H-h)/2
		} else {
			y = bounds.Y + s.Top.Resolve(bounds.H, y-bounds.Y)
		}
	}
	return geom.Rect{X: x, Y: y, W: w, H: h}
}

func wordCloudValueRange(items []option.DataValue) (float64, float64) {
	minValue, maxValue := math.Inf(1), math.Inf(-1)
	for _, d := range items {
		v := d.Number()
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}
	if minValue > maxValue {
		return 0, 1
	}
	return minValue, maxValue
}

func wordCloudSizeRange(values []float64) (float64, float64) {
	if len(values) >= 2 && values[1] > values[0] {
		return values[0], values[1]
	}
	return 16, 54
}

func pickWordCloudFamily(fallback, override string) string {
	if override != "" {
		return override
	}
	return fallback
}

func wordColor(s *option.WordCloudSeries, palette color.Palette, index int) color.Color {
	if s.TextStyle.Color != "" {
		if c, err := color.Parse(string(s.TextStyle.Color)); err == nil {
			return c
		}
	}
	return palette.At(index)
}

func wordRotation(rotationRange []float64, index int) float64 {
	if len(rotationRange) < 2 {
		return 0
	}
	minRot, maxRot := rotationRange[0], rotationRange[1]
	if maxRot <= minRot {
		return minRot
	}
	steps := []float64{minRot, 0, maxRot}
	return steps[index%len(steps)]
}

func placeWord(bounds geom.Rect, placed []geom.Rect, textW, textH, rotation, grid float64, seed int) (geom.Point, geom.Rect, bool) {
	cx, cy := wordOrigin(bounds, seed)
	for step := 0; step < 900; step++ {
		angle := float64(step) * 0.42
		radius := grid * angle / 1.8
		x := cx + math.Cos(angle)*radius
		y := cy + math.Sin(angle)*radius
		box := rotatedTextBox(x, y, textW, textH, rotation).InsetUniform(-2)
		if box.X < bounds.X || box.Y < bounds.Y || box.Right() > bounds.Right() || box.Bottom() > bounds.Bottom() {
			continue
		}
		if intersectsAny(box, placed) {
			continue
		}
		return geom.Point{X: x, Y: y}, box, true
	}
	return geom.Point{}, geom.Rect{}, false
}

func wordOrigin(bounds geom.Rect, seed int) (float64, float64) {
	if seed == 0 {
		return bounds.CenterX(), bounds.CenterY()
	}
	angle := float64(seed) * 2.399963229728653
	spread := math.Min(bounds.W, bounds.H) * 0.06 * math.Sqrt(float64(seed))
	x := bounds.CenterX() + math.Cos(angle)*spread
	y := bounds.CenterY() + math.Sin(angle)*spread
	margin := math.Min(bounds.W, bounds.H) * 0.12
	if x < bounds.X+margin {
		x = bounds.X + margin
	} else if x > bounds.Right()-margin {
		x = bounds.Right() - margin
	}
	if y < bounds.Y+margin {
		y = bounds.Y + margin
	} else if y > bounds.Bottom()-margin {
		y = bounds.Bottom() - margin
	}
	return x, y
}

func rotatedTextBox(cx, cy, w, h, rotation float64) geom.Rect {
	rad := math.Mod(math.Abs(rotation), 180) * math.Pi / 180
	bw := math.Abs(w*math.Cos(rad)) + math.Abs(h*math.Sin(rad))
	bh := math.Abs(w*math.Sin(rad)) + math.Abs(h*math.Cos(rad))
	return geom.Rect{X: cx - bw/2, Y: cy - bh/2, W: bw, H: bh}
}

func intersectsAny(box geom.Rect, placed []geom.Rect) bool {
	for _, other := range placed {
		if rectsIntersect(box, other) {
			return true
		}
	}
	return false
}

func rectsIntersect(a, b geom.Rect) bool {
	return a.X < b.Right() && a.Right() > b.X && a.Y < b.Bottom() && a.Bottom() > b.Y
}
