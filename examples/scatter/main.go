package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Height vs Weight"},
  "xAxis": {"type": "value", "name": "Weight (kg)"},
  "yAxis": {"type": "value", "name": "Height (cm)"},
  "series": [{"type": "scatter", "symbolSize": 16,
    "data": [[60, 170],[55, 165],[70, 175],[80, 180],[75, 178],[50, 160],[85, 185],[68, 172],[62, 168],[77, 182]]
  }]
}`

func main() {
	out, err := os.Create("scatter.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote scatter.png")
}
