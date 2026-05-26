package docx

import (
	"strings"

	"github.com/zzhtl/zcharts/chart"
	"github.com/zzhtl/zcharts/option"
)

// ContentBlock 是一组可按顺序插入到已有文档章节下的内容。
type ContentBlock struct {
	doc Document
}

// NewContentBlock 创建一个空内容块。
func NewContentBlock() *ContentBlock {
	return &ContentBlock{}
}

// AddParagraph 向内容块追加普通文字段落。
func (b *ContentBlock) AddParagraph(text string) {
	b.doc.AddParagraph(text)
}

// AddRichParagraph 向内容块追加富文本段落。
func (b *ContentBlock) AddRichParagraph(p RichParagraph) {
	b.doc.AddRichParagraph(p)
}

// AddStyledParagraph 向内容块追加统一字符样式的正文段落。
func (b *ContentBlock) AddStyledParagraph(text string, style RunStyle, align string) {
	b.doc.AddStyledParagraph(text, style, align)
}

// AddHeading 向内容块追加指定级别标题，level 取 1~6。
func (b *ContentBlock) AddHeading(text string, level int) {
	b.doc.AddHeading(text, level)
}

// AddChart 向内容块追加一个图表。
func (b *ContentBlock) AddChart(opt *option.Option, ins ChartInsertOption, renderOpts ...chart.RenderOption) error {
	return b.doc.AddChart(opt, ins, renderOpts...)
}

// AddTable 向内容块追加一个表格。
func (b *ContentBlock) AddTable(t Table) error {
	return b.doc.AddTable(t)
}

// xml 把内容块累积的 body 元素序列化为 OOXML 片段，供 EditableDocument 直接插入。
func (b *ContentBlock) xml() string {
	var body strings.Builder
	for _, el := range b.doc.body {
		body.WriteString(el.xml(&b.doc))
	}
	return body.String()
}
