package color

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Color
	}{
		{"", Transparent},
		{"transparent", Transparent},
		{"#000", RGB(0, 0, 0)},
		{"#fff", RGB(255, 255, 255)},
		{"#FFFFFF", RGB(255, 255, 255)},
		{"#1A2B3C", RGB(0x1A, 0x2B, 0x3C)},
		{"#11223344", RGBA(0x11, 0x22, 0x33, 0x44)},
		{"rgb(10, 20, 30)", RGB(10, 20, 30)},
		{"rgba(10, 20, 30, 0.5)", RGBA(10, 20, 30, 128)},
		{"rgba(10, 20, 30, 1)", RGBA(10, 20, 30, 255)},
		{"rgba(10, 20, 30, 0)", RGBA(10, 20, 30, 0)},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) err=%v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	bad := []string{"#xyz", "#12", "rgb(1,2)", "rgba(1,2,3)", "purple"}
	for _, s := range bad {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) expected error", s)
		}
	}
}

func TestHex(t *testing.T) {
	if got := RGB(0x1A, 0x2B, 0x3C).Hex(); got != "#1A2B3C" {
		t.Errorf("Hex = %s", got)
	}
	if got := RGBA(0x11, 0x22, 0x33, 0x44).Hex(); got != "#11223344" {
		t.Errorf("Hex = %s", got)
	}
}

func TestPaletteAt(t *testing.T) {
	p := MustPalette("#000", "#fff")
	if p.At(0) != RGB(0, 0, 0) {
		t.Error("At(0)")
	}
	if p.At(1) != RGB(255, 255, 255) {
		t.Error("At(1)")
	}
	if p.At(2) != RGB(0, 0, 0) {
		t.Error("At(2) should wrap")
	}
	var empty Palette
	if empty.At(0) != Transparent {
		t.Error("empty palette should return transparent")
	}
}
