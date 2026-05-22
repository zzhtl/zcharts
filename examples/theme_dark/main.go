package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Weekly Sales (Dark)", "subtext": "Week 2026-21"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu","Fri","Sat","Sun"]},
  "yAxis": {"type": "value"},
  "series": [
    {"name": "Cookies", "type": "line", "smooth": true,
     "data": [120, 200, 150, 80, 70, 110, 130]},
    {"name": "Drinks",  "type": "line", "smooth": true, "areaStyle": {},
     "data": [60, 90, 120, 150, 200, 240, 210]}
  ]
}`

func main() {
	out, err := os.Create("dark.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500), chart.WithTheme("dark")); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote dark.png")
}
