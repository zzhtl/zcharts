package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "服务器负载"},
  "legend": {},
  "series": [{
    "name": "CPU 使用率",
    "type": "gauge",
    "min": 0,
    "max": 100,
    "center": ["50%", "55%"],
    "radius": "70%",
    "data": [{"value": 72, "name": "CPU 使用率"}]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/gauge.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(600, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/gauge.png")
}
