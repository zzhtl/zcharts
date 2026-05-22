package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "身高体重分布"},
  "legend": {},
  "xAxis": {"type": "value", "name": "体重 (kg)"},
  "yAxis": {"type": "value", "name": "身高 (cm)"},
  "series": [{"name": "样本数据", "type": "scatter", "symbolSize": 16,
    "data": [[60, 170],[55, 165],[70, 175],[80, 180],[75, 178],[50, 160],[85, 185],[68, 172],[62, 168],[77, 182]]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/scatter.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/scatter.png")
}
