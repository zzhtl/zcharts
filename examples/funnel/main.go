package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "销售转化漏斗", "subtext": "线索到成交"},
  "legend": {},
  "series": [{
    "name": "转化阶段",
    "type": "funnel",
    "left": "center",
    "top": "middle",
    "width": "70%",
    "height": "76%",
    "gap": 6,
    "label": {"show": true, "fontSize": 13},
    "data": [
      {"name": "访问", "value": 1000},
      {"name": "注册", "value": 760},
      {"name": "试用", "value": 520},
      {"name": "报价", "value": 310},
      {"name": "成交", "value": 180}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/funnel.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(720, 560)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/funnel.png")
}
