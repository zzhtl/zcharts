// docx_edit 演示打开已有 Word 文档，读取标准化标题目录，
// 并在指定标题章节下追加同级标题、正文、图表和表格。
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zzhtl/zcharts/docx"
	"github.com/zzhtl/zcharts/jsonopt"
)

const editChartJSON = `{
  "title": {"text": "新增渠道趋势"},
  "legend": {},
  "xAxis": {"type": "category", "data": ["一月","二月","三月","四月"]},
  "yAxis": {"type": "value"},
  "series": [
    {"name": "自然流量", "type": "line", "smooth": true, "data": [120, 142, 151, 180]},
    {"name": "付费流量", "type": "bar", "data": [88, 96, 110, 128]}
  ]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0o755); err != nil {
		log.Fatal(err)
	}

	source := "assets/tmp/edit_source.docx"
	output := "assets/tmp/edit_updated.docx"
	if err := writeSourceDoc(source); err != nil {
		log.Fatal(err)
	}

	doc, err := docx.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	headings, err := doc.Headings()
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range headings {
		log.Printf("heading index=%d level=%d style=%s text=%s", h.Index, h.Level, h.StyleID, h.Text)
	}

	trend := mustFindHeading(headings, "3.1 华东")
	if err := doc.InsertHeadingUnderIndex(trend.Index, "3.2 新增同级标题"); err != nil {
		log.Fatal(err)
	}

	operation := mustFindHeading(headings, "2.2 运营")
	block := docx.NewContentBlock()
	block.AddParagraph("这段正文会追加到“2.2 运营”的原有内容之后，并且位于下一个同级标题之前。")
	block.AddRichParagraph(docx.RichParagraph{
		Runs: []docx.Run{
			{Text: "新增内容块支持 "},
			{Text: "正文、富文本、图表和表格", Style: docx.RunStyle{Bold: true, Color: "#C00000"}},
			{Text: " 按顺序自由组合。"},
		},
	})

	opt, err := jsonopt.ParseString(editChartJSON)
	if err != nil {
		log.Fatal(err)
	}
	if err := block.AddChart(opt, docx.AsImage(docx.PNG, 560, 320)); err != nil {
		log.Fatal(err)
	}
	if err := block.AddTable(docx.Table{
		Rows: []docx.Row{
			{Header: true, Cells: []docx.Cell{
				{Text: "指标", Shading: "#D9E1F2"},
				{Text: "当前值", Shading: "#D9E1F2", Align: "center"},
				{Text: "备注", Shading: "#D9E1F2"},
			}},
			{Cells: []docx.Cell{
				{Text: "活跃用户"},
				{Text: "18,240", Align: "center"},
				{Text: "新增章节写入"},
			}},
			{Cells: []docx.Cell{
				{Text: "转化率"},
				{Text: "12.8%", Align: "center", Style: docx.RunStyle{Color: "#C00000", Bold: true}},
				{Text: "环比提升"},
			}},
		},
	}); err != nil {
		log.Fatal(err)
	}
	if err := doc.InsertContentUnderIndex(operation.Index, block); err != nil {
		log.Fatal(err)
	}

	if err := doc.InsertParagraphUnderIndex(mustFindHeading(headings, "5.1 SEO").Index, "追加到 5 级标题原有内容下的正文。"); err != nil {
		log.Fatal(err)
	}

	if err := doc.Save(output); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s and %s", source, output)
}

func writeSourceDoc(path string) error {
	doc := docx.New()
	doc.AddTitle("已有 Word 文档")
	doc.AddHeading("一、总览", 1)
	doc.AddParagraph("这是一份用于演示“打开已有文档并追加内容”的基础文档。")
	doc.AddHeading("2.1 销售", 2)
	doc.AddParagraph("销售章节原有正文。")
	doc.AddHeading("3.1 华东", 3)
	doc.AddParagraph("华东章节原有正文。")
	doc.AddHeading("4.1 上海", 4)
	doc.AddParagraph("上海章节原有正文。")
	doc.AddHeading("5.1 SEO", 5)
	doc.AddParagraph("SEO 章节原有正文。")
	doc.AddHeading("3.2 华南", 3)
	doc.AddParagraph("华南章节原有正文。")
	doc.AddHeading("2.2 运营", 2)
	doc.AddParagraph("运营章节原有正文。")
	doc.AddHeading("3.3 流量", 3)
	doc.AddParagraph("流量章节原有正文。")
	doc.AddHeading("二、结论", 1)
	doc.AddParagraph("结论章节原有正文。")
	return doc.Save(path)
}

func mustFindHeading(headings []docx.Heading, text string) docx.Heading {
	for _, h := range headings {
		if h.Text == text {
			return h
		}
	}
	panic(fmt.Sprintf("heading %q not found", text))
}
