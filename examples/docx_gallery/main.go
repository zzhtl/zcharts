// docx_gallery 把所有支持的图表类型都画进一个 Word 文档，每种类型左右并排展示
// 「图片样式（PNG 嵌入）」与「Word 图表样式（原生 chart XML / 形状绘制）」，
// 方便直观对照两种渲染效果。
package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/docx"
	"github.com/zzhtl/zcharts/jsonopt"
	"github.com/zzhtl/zcharts/option"
)

// galleryItem 描述一种图表：名称、ECharts JSON、说明。
type galleryItem struct {
	name string
	note string // Word 侧的渲染方式说明
	json string
}

var items = []galleryItem{
	{"折线图", "Word 原生 lineChart", `{
		"title": {"text": "折线图"},
		"xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu","Fri"]},
		"yAxis": {"type": "value"},
		"series": [{"name": "销量", "type": "line", "smooth": true, "data": [120, 200, 150, 80, 170]}]
	}`},
	{"柱状图", "Word 原生 barChart", `{
		"title": {"text": "柱状图"},
		"xAxis": {"type": "category", "data": ["A","B","C","D"]},
		"yAxis": {"type": "value"},
		"series": [{"name": "数量", "type": "bar", "data": [50, 80, 60, 95]}]
	}`},
	{"堆积柱状图", "Word 原生堆积 barChart", `{
		"title": {"text": "堆积柱状图"},
		"legend": {},
		"xAxis": {"type": "category", "data": ["Q1","Q2","Q3"]},
		"yAxis": {"type": "value"},
		"series": [
			{"name": "线上", "type": "bar", "stack": "t", "data": [30, 45, 38]},
			{"name": "线下", "type": "bar", "stack": "t", "data": [20, 25, 30]}
		]
	}`},
	{"面积图", "Word 原生 areaChart", `{
		"title": {"text": "面积图"},
		"legend": {},
		"xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu","Fri"]},
		"yAxis": {"type": "value"},
		"series": [
			{"name": "PV", "type": "line", "areaStyle": {}, "stack": "t", "data": [120, 150, 130, 180, 160]},
			{"name": "UV", "type": "line", "areaStyle": {}, "stack": "t", "data": [60, 72, 68, 90, 84]}
		]
	}`},
	{"横向条形图", "Word 原生横向 barChart", `{
		"title": {"text": "横向条形图"},
		"xAxis": {"type": "value"},
		"yAxis": {"type": "category", "data": ["华东","华北","华南","西南"]},
		"series": [{"name": "销量", "type": "bar", "data": [120, 90, 60, 45]}]
	}`},
	{"饼图", "Word 原生 pieChart", `{
		"title": {"text": "饼图"},
		"series": [{"type": "pie", "label": {"show": true, "formatter": "{b}: {d}%"},
			"data": [{"name":"A","value":40},{"name":"B","value":30},{"name":"C","value":30}]}]
	}`},
	{"环图", "Word 原生 doughnutChart", `{
		"title": {"text": "环图"},
		"series": [{"type": "pie", "radius": ["42%","70%"], "label": {"show": true, "formatter": "{b}: {d}%"},
			"data": [{"name":"官网","value":35},{"name":"搜索","value":28},{"name":"合作","value":22},{"name":"其他","value":15}]}]
	}`},
	{"散点图", "Word 原生 scatterChart", `{
		"title": {"text": "散点图"},
		"xAxis": {"type": "value"},
		"yAxis": {"type": "value"},
		"series": [{"name": "样本", "type": "scatter", "symbolSize": 12,
			"data": [[10,42],[18,68],[24,81],[32,96],[40,126]]}]
	}`},
	{"雷达图", "Word 原生 radarChart", `{
		"title": {"text": "雷达图"},
		"radar": {"indicator": [{"name":"质量"},{"name":"效率"},{"name":"稳定"},{"name":"体验"}]},
		"series": [{"type": "radar", "data": [
			{"name":"当前","value":[80,68,91,76]},
			{"name":"目标","value":[92,86,95,88]}]}]
	}`},
	{"热力图", "Word 形状绘制（VML）", `{
		"title": {"text": "热力图"},
		"xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu"]},
		"yAxis": {"type": "category", "data": ["早","中","晚"]},
		"visualMap": {"min": 0, "max": 10, "inRange": {"color": ["#EEF3FF", "#5470C6"]}},
		"series": [{"type": "heatmap",
			"data": [[0,0,3],[1,0,5],[2,0,8],[3,0,6],[0,1,6],[1,1,9],[2,1,7],[3,1,4],[0,2,2],[1,2,4],[2,2,6],[3,2,9]]}]
	}`},
	{"仪表盘", "Word 形状绘制（VML）", `{
		"title": {"text": "仪表盘"},
		"series": [{"type": "gauge", "min": 0, "max": 100, "data": [{"name": "完成率", "value": 76}]}]
	}`},
	{"漏斗图", "Word 形状绘制（VML）", `{
		"title": {"text": "漏斗图"},
		"series": [{"type": "funnel", "data": [
			{"name":"访问","value":100},{"name":"咨询","value":68},{"name":"试用","value":42},{"name":"成交","value":18}]}]
	}`},
	{"时间轴", "Word 形状绘制（VML）", `{
		"title": {"text": "时间轴"},
		"series": [{"type": "timeline", "data": [
			{"time":"05-01","title":"需求确认","content":"完成范围确认"},
			{"time":"05-10","title":"开发联调","content":"完成核心图表"},
			{"time":"05-22","title":"文档验收","content":"输出 Word 图表"}]}]
	}`},
	{"词云", "Word 文本框绘制（VML）", `{
		"title": {"text": "词云"},
		"series": [{"type": "wordCloud", "sizeRange": [14, 34], "data": [
			{"name":"Word","value":80},{"name":"WPS","value":70},{"name":"原生图表","value":66},
			{"name":"散点","value":45},{"name":"雷达","value":42},{"name":"形状绘制","value":58}]}]
	}`},
}

func main() {
	if err := os.MkdirAll("assets/tmp", 0o755); err != nil {
		log.Fatal(err)
	}

	doc := docx.New()
	doc.AddTitle("zcharts 图表样式对照")
	doc.AddStyledParagraph(
		"下表每种图表左右并排：左侧为「图片样式」（渲染成 PNG 后嵌入），右侧为「Word 图表样式」"+
			"（优先 Word 原生可编辑 chart XML，非标准类型走 Word/WPS 形状绘制）。",
		docx.RunStyle{Color: "#555555"}, "")

	const w, h = 320, 220
	for _, it := range items {
		doc.AddHeading(it.name, 2)
		opt, err := jsonopt.ParseString(it.json)
		if err != nil {
			log.Fatalf("%s: %v", it.name, err)
		}
		if err := doc.AddTable(galleryRow(opt, it.note, w, h)); err != nil {
			log.Fatalf("%s: %v", it.name, err)
		}
	}

	if err := doc.Save("assets/tmp/gallery.docx"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/gallery.docx")
}

// galleryRow 构造一行对照表格：表头 + 左图片右 Word 图表。
func galleryRow(opt *option.Option, note string, w, h int) docx.Table {
	return docx.Table{
		Rows: []docx.Row{
			{
				Header: true,
				Cells: []docx.Cell{
					{Text: "图片样式 (PNG)", Align: "center", Shading: "#E7EEF8"},
					{Text: "Word 图表样式 · " + note, Align: "center", Shading: "#E7EEF8"},
				},
			},
			{
				Cells: []docx.Cell{
					{Chart: &docx.CellChart{Option: opt, Insert: docx.AsImage(docx.PNG, w, h)}, VAlign: "center"},
					{Chart: &docx.CellChart{Option: opt, Insert: docx.AsNativeChart(w, h)}, VAlign: "center"},
				},
			},
		},
	}
}
