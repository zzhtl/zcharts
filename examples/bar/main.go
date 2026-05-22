package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Monthly Revenue", "subtext": "2026 H1"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["Jan","Feb","Mar","Apr","May","Jun"]},
  "yAxis": {"type": "value", "name": "USD K"},
  "series": [
    {"name": "Direct",   "type": "bar", "data": [120, 200, 150, 80, 70, 110]},
    {"name": "Referral", "type": "bar", "data": [60, 90, 120, 150, 200, 240]},
    {"name": "Search",   "type": "bar", "data": [30, 50, 80, 95, 120, 140]}
  ]
}`

func main() {
	out, err := os.Create("bar.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote bar.png")
}
