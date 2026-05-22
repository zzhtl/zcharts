package main

import (
	"log"

	"github.com/zzhtl/zcharts/docx"
	"github.com/zzhtl/zcharts/jsonopt"
)

const lineJSON = `{
  "title": {"text": "Weekly Sales"},
  "xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu","Fri","Sat","Sun"]},
  "yAxis": {"type": "value"},
  "series": [{"name": "A", "type": "line", "smooth": true, "data": [120, 200, 150, 80, 70, 110, 130]}]
}`

const pieJSON = `{
  "title": {"text": "Traffic Sources"},
  "series": [{
    "type": "pie",
    "radius": ["35%", "70%"],
    "label": {"show": true, "formatter": "{b}: {d}%"},
    "data": [
      {"name": "Direct", "value": 335},
      {"name": "Email",  "value": 310},
      {"name": "Ads",    "value": 234},
      {"name": "Search", "value": 548}
    ]
  }]
}`

func main() {
	doc := docx.New()
	doc.AddParagraph("月度业务报告")
	doc.AddParagraph("以下为本月销售折线图：")

	line, _ := jsonopt.ParseString(lineJSON)
	if err := doc.AddChart(line, docx.AsImage(docx.PNG, 720, 400)); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("流量来源分布：")

	pie, _ := jsonopt.ParseString(pieJSON)
	if err := doc.AddChart(pie, docx.AsImage(docx.PNG, 640, 480)); err != nil {
		log.Fatal(err)
	}

	if err := doc.Save("report.docx"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote report.docx")
}
