package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "月度营收", "subtext": "2026 上半年"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["一月","二月","三月","四月","五月","六月"]},
  "yAxis": {"type": "value", "name": "万元"},
  "series": [
    {"name": "直营收入", "type": "bar", "data": [120, 200, 150, 80, 70, 110]},
    {"name": "渠道收入", "type": "bar", "data": [60, 90, 120, 150, 200, 240]},
    {"name": "搜索收入", "type": "bar", "data": [30, 50, 80, 95, 120, 140]}
  ]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/bar.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/bar.png")
}
