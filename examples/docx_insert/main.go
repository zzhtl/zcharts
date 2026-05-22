package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/docx"
	"github.com/zzhtl/zcharts/jsonopt"
)

const lineJSON = `{
  "title": {"text": "每周销量"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["周一","周二","周三","周四","周五","周六","周日"]},
  "yAxis": {"type": "value"},
  "series": [{"name": "线上销量", "type": "line", "smooth": true, "data": [120, 200, 150, 80, 70, 110, 130]}]
}`

const pieJSON = `{
  "title": {"text": "流量来源"},
  "legend": {},
  "series": [{
    "name": "访问来源",
    "type": "pie",
    "radius": ["35%", "70%"],
    "label": {"show": true, "formatter": "{b}: {d}%"},
    "data": [
      {"name": "直接访问", "value": 335},
      {"name": "邮件营销", "value": 310},
      {"name": "联盟广告", "value": 234},
      {"name": "搜索引擎", "value": 548}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
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

	if err := doc.Save("assets/tmp/report.docx"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/report.docx")
}
