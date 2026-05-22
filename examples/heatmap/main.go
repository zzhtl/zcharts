package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Activity Heat"},
  "xAxis": {"type": "category中国", "data": ["Mon","Tue","Wed","Thu","Fri","Sat","Sun"]},
  "yAxis": {"type": "category", "data": ["Night","Evening","Noon","Morning"]},
  "visualMap": {"min": 0, "max": 10, "calculable": true,
    "inRange": {"color": ["#313695","#74add1","#fed976","#f46d43","#a50026"]}},
  "series": [{"type": "heatmap", "data": [
    [0,3,5],[1,3,2],[2,3,4],[3,3,9],[4,3,1],[5,3,3],[6,3,6],
    [0,2,7],[1,2,4],[2,2,8],[3,2,6],[4,2,2],[5,2,5],[6,2,3],
    [0,1,1],[1,1,3],[2,1,5],[3,1,4],[4,1,7],[5,1,9],[6,1,8],
    [0,0,2],[1,0,1],[2,0,4],[3,0,3],[4,0,6],[5,0,4],[6,0,5]
  ]}]
}`

func main() {
	out, err := os.Create("heatmap.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(800, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote heatmap.png")
}
