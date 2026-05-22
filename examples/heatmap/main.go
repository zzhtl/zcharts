package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "活跃度热力图"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["周一","周二","周三","周四","周五","周六","周日"]},
  "yAxis": {"type": "category", "data": ["夜间","傍晚","中午","早晨"]},
  "visualMap": {"min": 0, "max": 10, "calculable": true,
    "inRange": {"color": ["#313695","#74add1","#fed976","#f46d43","#a50026"]}},
  "series": [{"name": "活跃人数", "type": "heatmap", "data": [
    [0,3,5],[1,3,2],[2,3,4],[3,3,9],[4,3,1],[5,3,3],[6,3,6],
    [0,2,7],[1,2,4],[2,2,8],[3,2,6],[4,2,2],[5,2,5],[6,2,3],
    [0,1,1],[1,1,3],[2,1,5],[3,1,4],[4,1,7],[5,1,9],[6,1,8],
    [0,0,2],[1,0,1],[2,0,4],[3,0,3],[4,0,6],[5,0,4],[6,0,5]
  ]}]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/heatmap.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/heatmap.png")
}
