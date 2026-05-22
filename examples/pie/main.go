package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "流量来源", "subtext": "2026 第一季度"},
  "legend": {},
  "series": [{
    "name": "访问来源",
    "type": "pie",
    "radius": ["35%", "70%"],
    "center": ["50%", "55%"],
    "label": {"show": true, "formatter": "{b}: {d}%"},
    "data": [
      {"name": "直接访问", "value": 335},
      {"name": "邮件营销", "value": 310},
      {"name": "联盟广告", "value": 234},
      {"name": "视频广告", "value": 135},
      {"name": "搜索引擎", "value": 548}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/pie.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/pie.png")
}
