package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Traffic Sources", "subtext": "2026 Q1"},
  "series": [{
    "name": "Source",
    "type": "pie",
    "radius": ["35%", "70%"],
    "center": ["50%", "55%"],
    "label": {"show": true, "formatter": "{b}: {d}%"},
    "data": [
      {"name": "Direct",   "value": 335},
      {"name": "Email",    "value": 310},
      {"name": "Ads",      "value": 234},
      {"name": "Video",    "value": 135},
      {"name": "Search",   "value": 548}
    ]
  }]
}`

func main() {
	out, err := os.Create("pie.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote pie.png")
}
