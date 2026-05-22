package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Server Load"},
  "series": [{
    "type": "gauge",
    "min": 0,
    "max": 100,
    "center": ["50%", "55%"],
    "radius": "70%",
    "data": [{"value": 72, "name": "CPU"}]
  }]
}`

func main() {
	out, err := os.Create("gauge.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(600, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote gauge.png")
}
