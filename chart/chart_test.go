package chart

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render"
)

func TestRenderCustomSeries(t *testing.T) {
	src := []byte(`{"series":[{"type":"metricCard","data":[{"name":"A","value":1}]}]}`)
	var out bytes.Buffer
	err := RenderFromJSON(src, FormatPNG, &out,
		WithSize(320, 200),
		WithSeriesRenderer("metricCard", func(ctx *render.Context, s option.Series, index int) error {
			ctx.Canvas.SetFill(color.RGB(84, 112, 198))
			ctx.Canvas.NoStroke()
			ctx.Canvas.DrawRect(40, 50, 240, 100)
			return nil
		}))
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if got := img.Bounds().Dx(); got != 320 {
		t.Fatalf("width=%d", got)
	}
}

func TestRenderTimelineSeries(t *testing.T) {
	src := []byte(`{"legend":{"show":false},"series":[{"type":"timeline","data":[{"time":"2026-05-01","title":"启动","content":"项目启动"},{"time":"2026-05-02","title":"发布","content":"版本发布"}]}]}`)
	var out bytes.Buffer
	err := RenderFromJSON(src, FormatPNG, &out, WithSize(360, 240))
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if got := img.Bounds().Dy(); got != 240 {
		t.Fatalf("height=%d", got)
	}
}
