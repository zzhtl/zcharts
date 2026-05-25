// docx_report 演示用 zcharts/docx 自由组合标题、富文本正文、列表、富表格与图表，
// 生成一份完整的 Word 报告。
package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/docx"
	"github.com/zzhtl/zcharts/jsonopt"
)

const salesLineJSON = `{
  "title": {"text": "季度销售趋势"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["Q1","Q2","Q3","Q4"]},
  "yAxis": {"type": "value"},
  "series": [
    {"name": "线上", "type": "line", "smooth": true, "data": [120, 200, 150, 280]},
    {"name": "线下", "type": "bar", "data": [90, 140, 110, 190]}
  ]
}`

const sharePieJSON = `{
  "title": {"text": "渠道占比"},
  "series": [{
    "type": "pie", "radius": ["35%","70%"],
    "label": {"show": true, "formatter": "{b}: {d}%"},
    "data": [
      {"name": "直营", "value": 335},
      {"name": "代理", "value": 234},
      {"name": "电商", "value": 548}
    ]
  }]
}`

const miniBarJSON = `{
  "xAxis": {"type": "category", "data": ["一","二","三"]},
  "yAxis": {"type": "value"},
  "series": [{"type": "bar", "data": [20, 35, 28]}]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0o755); err != nil {
		log.Fatal(err)
	}

	doc := docx.New()

	// 文档大标题 + 富文本副标题。
	doc.AddTitle("2026 年度业务分析报告")
	doc.AddStyledParagraph("数据周期：2026-01 ~ 2026-12    制作：业务分析组",
		docx.RunStyle{Italic: true, Color: "#808080", Size: 10.5}, "center")

	// 一级标题 + 正文 + 列表。
	doc.AddHeading("一、整体概况", 1)
	doc.AddRichParagraph(docx.RichParagraph{
		Runs: []docx.Run{
			{Text: "本年度整体营收实现 "},
			{Text: "稳健增长", Style: docx.RunStyle{Bold: true, Color: "#C00000"}},
			{Text: "，主要驱动因素如下："},
		},
	})
	doc.AddBulletList(
		"线上渠道持续放量，Q4 创历史新高",
		"代理体系完成区域整合",
		"电商大促贡献显著增量",
	)

	// 二级标题 + 有序列表。
	doc.AddHeading("二、趋势与渠道", 1)
	doc.AddHeading("2.1 销售趋势", 2)
	doc.AddParagraph("季度销售（折线+柱状组合）：")
	line, err := jsonopt.ParseString(salesLineJSON)
	if err != nil {
		log.Fatal(err)
	}
	if err := doc.AddChart(line, docx.AsNativeChart(640, 360)); err != nil {
		log.Fatal(err)
	}

	doc.AddHeading("2.2 渠道占比", 2)
	pie, err := jsonopt.ParseString(sharePieJSON)
	if err != nil {
		log.Fatal(err)
	}
	if err := doc.AddChart(pie, docx.AsImage(docx.PNG, 560, 360)); err != nil {
		log.Fatal(err)
	}

	// 三级标题 + 富表格（合并表头 + 底色 + 单元格内嵌迷你图表）。
	doc.AddHeading("三、关键指标明细", 1)
	mini, err := jsonopt.ParseString(miniBarJSON)
	if err != nil {
		log.Fatal(err)
	}
	white := docx.RunStyle{Bold: true, Color: "#FFFFFF"}
	if err := doc.AddTable(docx.Table{
		Rows: []docx.Row{
			{
				Header: true,
				Cells: []docx.Cell{
					{Text: "核心指标", GridSpan: 3, Align: "center", Shading: "#4472C4", Style: white},
				},
			},
			{
				Header: true,
				Cells: []docx.Cell{
					{Text: "指标", Shading: "#D9E1F2"},
					{Text: "数值", Shading: "#D9E1F2", Align: "center"},
					{Text: "走势", Shading: "#D9E1F2", Align: "center"},
				},
			},
			{
				Cells: []docx.Cell{
					{Text: "营收(万元)"},
					{Text: "1,250", Align: "center"},
					{Chart: &docx.CellChart{Option: mini, Insert: docx.AsImage(docx.PNG, 240, 130)}, VAlign: "center"},
				},
			},
			{
				Cells: []docx.Cell{
					{Text: "同比"},
					{Text: "+18.6%", Align: "center", Style: docx.RunStyle{Color: "#C00000", Bold: true}},
					{Text: "持续向好", Align: "center"},
				},
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	doc.AddHeading("四、结论", 1)
	doc.AddStyledParagraph("综合来看，全年表现超出预期，建议下一年度加大线上投入。",
		docx.RunStyle{}, "both")

	if err := doc.Save("assets/tmp/full_report.docx"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/full_report.docx")
}
