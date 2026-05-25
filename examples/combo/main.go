package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "成交与目标趋势", "subtext": "折线 + 柱状图"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["周一","周二","周三","周四","周五","周六","周日"]},
  "yAxis": {"type": "value", "name": "单量"},
  "series": [
    {"name": "成交单量", "type": "bar", "data": [320, 420, 510, 460, 620, 710, 680]},
    {"name": "目标趋势", "type": "line", "smooth": true, "symbolSize": 7, "data": [300, 390, 480, 540, 600, 660, 720]}
  ]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/combo.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(820, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/combo.png")
}
