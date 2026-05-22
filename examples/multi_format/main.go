package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Multi-Format Demo"},
  "xAxis": {"type": "category", "data": ["A","B","C","D","E"]},
  "yAxis": {"type": "value"},
  "series": [{"type": "bar", "data": [12, 28, 15, 30, 22]}]
}`

func main() {
	for _, f := range []struct {
		ext string
		fmt chart.Format
	}{
		{"png", chart.FormatPNG},
		{"svg", chart.FormatSVG},
		{"pdf", chart.FormatPDF},
	} {
		out, err := os.Create("multi." + f.ext)
		if err != nil {
			log.Fatal(err)
		}
		if err := chart.RenderFromJSON([]byte(optionJSON), f.fmt, out, chart.WithSize(640, 400)); err != nil {
			log.Fatal(err)
		}
		out.Close()
		log.Println("wrote multi." + f.ext)
	}
}
