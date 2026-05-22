package font

import "testing"

func TestNewHasDefault(t *testing.T) {
	m := New()
	if !m.Has(DefaultFamily) {
		t.Fatal("default family missing")
	}
	face, err := m.Face(DefaultFamily, 14)
	if err != nil {
		t.Fatalf("Face err=%v", err)
	}
	if face == nil {
		t.Fatal("face nil")
	}
	if advance, ok := face.GlyphAdvance('中'); !ok || advance == 0 {
		t.Fatalf("default font should support chinese glyph, ok=%v advance=%v", ok, advance)
	}
}

func TestFallbackFamily(t *testing.T) {
	m := New()
	face, err := m.Face("does-not-exist", 12)
	if err != nil || face == nil {
		t.Fatalf("should fall back, err=%v face=%v", err, face)
	}
}
