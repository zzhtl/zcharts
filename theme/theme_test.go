package theme

import "testing"

func TestRegistryDefaults(t *testing.T) {
	for _, name := range []string{"default", "light", "dark"} {
		th, err := Get(name)
		if err != nil {
			t.Fatalf("get %s: %v", name, err)
		}
		if len(th.Palette) == 0 {
			t.Errorf("%s: empty palette", name)
		}
		if th.TextStyle.FontSize <= 0 {
			t.Errorf("%s: invalid font size", name)
		}
	}
}

func TestParseECharts(t *testing.T) {
	src := []byte(`{"color":["#fff","#000"],"backgroundColor":"#222","textStyle":{"color":"#abc","fontSize":14}}`)
	th, err := ParseECharts("my", src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(th.Palette) != 2 || th.Palette.At(0).Hex() != "#FFFFFF" {
		t.Errorf("palette=%v", th.Palette)
	}
	if th.BackgroundColor.Hex() != "#222222" {
		t.Errorf("bg=%v", th.BackgroundColor.Hex())
	}
	if th.TextStyle.FontSize != 14 {
		t.Errorf("font size=%v", th.TextStyle.FontSize)
	}
}
