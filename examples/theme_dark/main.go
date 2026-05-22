package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "每周销量（暗色）", "subtext": "2026 第 21 周"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["周一","周二","周三","周四","周五","周六","周日"]},
  "yAxis": {"type": "value"},
  "series": [
    {"name": "饼干销量", "type": "line", "smooth": true,
     "data": [120, 200, 150, 80, 70, 110, 130]},
    {"name": "饮料销量", "type": "line", "smooth": true, "areaStyle": {},
     "data": [60, 90, 120, 150, 200, 240, 210]}
  ]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/dark.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500), chart.WithTheme("dark")); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/dark.png")
}
