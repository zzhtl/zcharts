package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "多格式导出示例"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["产品A","产品B","产品C","产品D","产品E"]},
  "yAxis": {"type": "value"},
  "series": [{"name": "销量", "type": "bar", "data": [12, 28, 15, 30, 22]}]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	for _, f := range []struct {
		ext string
		fmt chart.Format
	}{
		{"png", chart.FormatPNG},
		{"svg", chart.FormatSVG},
		{"pdf", chart.FormatPDF},
	} {
		name := "assets/tmp/multi." + f.ext
		out, err := os.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		if err := chart.RenderFromJSON([]byte(optionJSON), f.fmt, out, chart.WithSize(640, 400)); err != nil {
			log.Fatal(err)
		}
		out.Close()
		log.Println("wrote " + name)
	}
}
