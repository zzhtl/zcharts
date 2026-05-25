package main

import (
	"log"
	"os"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/chart"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/number"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render"
)

const optionJSON = `{
  "title": {"text": "自定义指标卡片", "subtext": "通过 WithSeriesRenderer 扩展 series.type"},
  "series": [{
    "name": "指标",
    "type": "metricCard",
    "data": [
      {"name": "新增客户", "value": 128},
      {"name": "转化率", "value": 42},
      {"name": "满意度", "value": 96}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/custom.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(760, 420),
		chart.WithSeriesRenderer("metricCard", drawMetricCards)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/custom.png")
}

func drawMetricCards(ctx *render.Context, s option.Series, _ int) error {
	custom, ok := s.(*option.CustomSeries)
	if !ok {
		return nil
	}
	if len(custom.Data) == 0 {
		return nil
	}
	area := ctx.Bounds.Inset(120, 58, 70, 58)
	gap := 18.0
	cardW := (area.W - gap*float64(len(custom.Data)-1)) / float64(len(custom.Data))
	cardH := area.H
	for i, d := range custom.Data {
		x := area.X + float64(i)*(cardW+gap)
		y := area.Y
		mainColor := ctx.Theme.Palette.At(i)
		ctx.Canvas.SetFill(color.RGBA(mainColor.R, mainColor.G, mainColor.B, 28))
		ctx.Canvas.SetStroke(color.RGBA(mainColor.R, mainColor.G, mainColor.B, 140))
		ctx.Canvas.SetStrokeWidth(1)
		ctx.Canvas.DrawRect(x, y, cardW, cardH)

		ctx.Canvas.SetFill(mainColor)
		ctx.Canvas.NoStroke()
		ctx.Canvas.DrawCircle(x+36, y+42, 16)

		ctx.Canvas.DrawText(x+64, y+36, d.Name, zcanvas.TextStyle{
			Family: ctx.Theme.TextStyle.FontFamily,
			Size:   14,
			Color:  ctx.Theme.TextStyle.Color,
			Anchor: zcanvas.AnchorStart,
			VAlign: zcanvas.AlignMiddle,
		})
		ctx.Canvas.DrawText(x+cardW/2, y+cardH/2+18, formatMetricValue(d.Number(), i), zcanvas.TextStyle{
			Family: ctx.Theme.TextStyle.FontFamily,
			Size:   34,
			Weight: "bold",
			Color:  mainColor,
			Anchor: zcanvas.AnchorMiddle,
			VAlign: zcanvas.AlignMiddle,
		})
	}
	return nil
}

func formatMetricValue(value float64, index int) string {
	switch index {
	case 1:
		return number.FormatAuto(value) + "%"
	default:
		return number.FormatAuto(value)
	}
}
