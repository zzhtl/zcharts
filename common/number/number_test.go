package number

import (
	"math"
	"testing"
)

func TestFormatFloat(t *testing.T) {
	cases := []struct{ in float64; d int; want string }{
		{1.2300, 4, "1.23"},
		{100, 2, "100"},
		{0, 2, "0"},
		{-0.5, 2, "-0.5"},
	}
	for _, c := range cases {
		if got := FormatFloat(c.in, c.d); got != c.want {
			t.Errorf("FormatFloat(%v,%d)=%q want %q", c.in, c.d, got, c.want)
		}
	}
}

func TestNiceRange(t *testing.T) {
	lo, hi, step := NiceRange(3, 97, 5, true)
	if step <= 0 {
		t.Fatalf("step=%v", step)
	}
	if lo > 3 || hi < 97 {
		t.Errorf("expand failed lo=%v hi=%v", lo, hi)
	}
	// step 应为 1/2/5/10/20/50/100 之一
	exp := math.Round(math.Log10(step))
	frac := step / math.Pow(10, exp)
	if frac != 1 && frac != 2 && frac != 5 {
		t.Errorf("step %v not nice (frac=%v)", step, frac)
	}
}

func TestTicks(t *testing.T) {
	out := Ticks(0, 10, 2)
	want := []float64{0, 2, 4, 6, 8, 10}
	if len(out) != len(want) {
		t.Fatalf("len=%d want=%d", len(out), len(want))
	}
	for i, v := range out {
		if math.Abs(v-want[i]) > 1e-9 {
			t.Errorf("tick[%d]=%v want %v", i, v, want[i])
		}
	}
}

func TestMap(t *testing.T) {
	if v := Map(5, 0, 10, 100, 200); v != 150 {
		t.Errorf("Map=%v", v)
	}
}

func TestClamp(t *testing.T) {
	if Clamp(-1, 0, 10) != 0 {
		t.Error("clamp lo")
	}
	if Clamp(11, 0, 10) != 10 {
		t.Error("clamp hi")
	}
	if Clamp(5, 0, 10) != 5 {
		t.Error("clamp mid")
	}
}
