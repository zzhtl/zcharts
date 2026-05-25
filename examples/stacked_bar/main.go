package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "渠道成本构成", "subtext": "堆积柱状图"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["一月","二月","三月","四月","五月","六月"]},
  "yAxis": {"type": "value", "name": "万元"},
  "series": [
    {"name": "投放", "type": "bar", "stack": "成本", "data": [45, 58, 64, 72, 80, 92]},
    {"name": "人力", "type": "bar", "stack": "成本", "data": [28, 32, 35, 36, 40, 42]},
    {"name": "服务", "type": "bar", "stack": "成本", "data": [18, 22, 26, 28, 31, 34]},
    {"name": "目标线", "type": "line", "smooth": true, "showSymbol": false, "data": [90, 105, 118, 130, 145, 160]}
  ]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/stacked_bar.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(860, 520)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/stacked_bar.png")
}
