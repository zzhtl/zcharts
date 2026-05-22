package geom

import (
	"math"
	"testing"
)

func TestPointDistance(t *testing.T) {
	a, b := Point{0, 0}, Point{3, 4}
	if d := a.Distance(b); math.Abs(d-5) > 1e-9 {
		t.Errorf("distance=%v", d)
	}
}

func TestRectInset(t *testing.T) {
	r := Rect{X: 0, Y: 0, W: 100, H: 50}
	in := r.InsetUniform(10)
	if in != (Rect{10, 10, 80, 30}) {
		t.Errorf("inset=%+v", in)
	}
	if !r.Contains(Point{50, 25}) {
		t.Error("center should be contained")
	}
	if r.Contains(Point{101, 0}) {
		t.Error("outside should not be contained")
	}
}
