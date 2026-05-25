package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/docx"
	"github.com/zzhtl/zcharts/jsonopt"
)

const comboJSON = `{
  "title": {"text": "季度经营数据"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["Q1","Q2","Q3","Q4"]},
  "yAxis": {"type": "value"},
  "series": [
    {"name": "收入", "type": "bar", "data": [120, 200, 180, 240]},
    {"name": "利润率", "type": "line", "smooth": true, "data": [18, 26, 22, 31]}
  ]
}`

const pieJSON = `{
  "title": {"text": "渠道占比"},
  "legend": {},
  "series": [{
    "name": "渠道",
    "type": "pie",
    "radius": ["42%", "68%"],
    "data": [
      {"name": "官网", "value": 35},
      {"name": "搜索", "value": 28},
      {"name": "合作伙伴", "value": 22},
      {"name": "其他", "value": 15}
    ]
  }]
}`

const stackedBarJSON = `{
  "title": {"text": "堆积柱状图"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu"]},
  "yAxis": {"type": "value"},
  "series": [
    {"name": "线上", "type": "bar", "stack": "total", "data": [32, 44, 39, 52]},
    {"name": "线下", "type": "bar", "stack": "total", "data": [18, 24, 21, 29]}
  ]
}`

const scatterJSON = `{
  "title": {"text": "转化关系"},
  "xAxis": {"type": "value"},
  "yAxis": {"type": "value"},
  "series": [{
    "name": "样本",
    "type": "scatter",
    "data": [[10, 42], [18, 68], [24, 81], [32, 96], [40, 126]]
  }]
}`

const radarJSON = `{
  "title": {"text": "能力雷达"},
  "radar": {
    "indicator": [
      {"name": "质量"},
      {"name": "效率"},
      {"name": "稳定"},
      {"name": "体验"}
    ]
  },
  "series": [{
    "name": "团队",
    "type": "radar",
    "data": [
      {"name": "当前", "value": [80, 68, 91, 76]},
      {"name": "目标", "value": [92, 86, 95, 88]}
    ]
  }]
}`

const heatmapJSON = `{
  "title": {"text": "活跃热力"},
  "xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu"]},
  "yAxis": {"type": "category", "data": ["早","中","晚"]},
  "visualMap": {"min": 0, "max": 10, "inRange": {"color": ["#EEF3FF", "#5470C6"]}},
  "series": [{
    "type": "heatmap",
    "data": [[0,0,3],[1,0,5],[2,0,8],[3,0,6],[0,1,6],[1,1,9],[2,1,7],[3,1,4],[0,2,2],[1,2,4],[2,2,6],[3,2,9]]
  }]
}`

const gaugeJSON = `{
  "title": {"text": "达成率"},
  "series": [{
    "name": "完成率",
    "type": "gauge",
    "min": 0,
    "max": 100,
    "data": [{"name": "完成率", "value": 76}]
  }]
}`

const funnelJSON = `{
  "title": {"text": "转化漏斗"},
  "series": [{
    "type": "funnel",
    "data": [
      {"name": "访问", "value": 100},
      {"name": "咨询", "value": 68},
      {"name": "试用", "value": 42},
      {"name": "成交", "value": 18}
    ]
  }]
}`

const timelineJSON = `{
  "title": {"text": "项目时间轴"},
  "series": [{
    "type": "timeline",
    "data": [
      {"time": "05-01", "title": "需求确认", "content": "完成范围确认"},
      {"time": "05-10", "title": "开发联调", "content": "完成核心图表"},
      {"time": "05-22", "title": "文档验收", "content": "输出 Word 原生图表"}
    ]
  }]
}`

const wordCloudJSON = `{
  "title": {"text": "关键词"},
  "series": [{
    "type": "wordCloud",
    "sizeRange": [14, 34],
    "data": [
      {"name": "Word", "value": 80},
      {"name": "WPS", "value": 70},
      {"name": "原生图表", "value": 66},
      {"name": "散点", "value": 45},
      {"name": "雷达", "value": 42},
      {"name": "形状绘制", "value": 58},
      {"name": "无图片", "value": 60}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	doc := docx.New()
	doc.AddParagraph("Word 原生图表示例")
	doc.AddParagraph("以下图表优先以 OOXML chart XML 写入；标准 chart 覆盖不到时使用 Word 形状绘制，最后才回退 PNG。")

	if err := addNativeChart(doc, comboJSON, 720, 420); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("堆积柱状图（Word 原生堆积柱）：")
	if err := addNativeChart(doc, stackedBarJSON, 640, 420); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("渠道占比（Word 原生环图）：")
	if err := addNativeChart(doc, pieJSON, 640, 420); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("散点图（Word 原生散点图）：")
	if err := addNativeChart(doc, scatterJSON, 640, 420); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("雷达图（Word 原生雷达图）：")
	if err := addNativeChart(doc, radarJSON, 640, 420); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("热力图（Word 形状绘制）：")
	if err := addNativeChart(doc, heatmapJSON, 640, 360); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("仪表盘（Word 形状绘制）：")
	if err := addNativeChart(doc, gaugeJSON, 520, 320); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("漏斗图（Word 形状绘制）：")
	if err := addNativeChart(doc, funnelJSON, 560, 360); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("时间轴（Word 形状绘制）：")
	if err := addNativeChart(doc, timelineJSON, 640, 360); err != nil {
		log.Fatal(err)
	}
	doc.AddParagraph("词云（Word 文本框绘制）：")
	if err := addNativeChart(doc, wordCloudJSON, 640, 320); err != nil {
		log.Fatal(err)
	}

	if err := doc.Save("assets/tmp/native_charts.docx"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/native_charts.docx")
}

func addNativeChart(doc *docx.Document, optionJSON string, width, height int) error {
	opt, err := jsonopt.ParseString(optionJSON)
	if err != nil {
		return err
	}
	return doc.AddChart(opt, docx.AsNativeChart(width, height))
}
