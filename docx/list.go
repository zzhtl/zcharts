package docx

import (
	"fmt"
	"strings"
)

// listParagraph 是一个列表项段落，通过 numPr 引用 numbering.xml 中的列表定义。
type listParagraph struct {
	runs  []Run
	numID int // numBulletID 或 numOrderedID
	level int // 0 起的缩进级别
}

func (p listParagraph) xml(_ *Document) string {
	level := p.level
	if level < 0 {
		level = 0
	}
	if level > 3 {
		level = 3
	}
	var b strings.Builder
	b.WriteString("<w:p><w:pPr>")
	fmt.Fprintf(&b, `<w:numPr><w:ilvl w:val="%d"/><w:numId w:val="%d"/></w:numPr>`, level, p.numID)
	b.WriteString("</w:pPr>")
	for _, r := range p.runs {
		b.WriteString(runXML(r))
	}
	b.WriteString("</w:p>")
	return b.String()
}

// AddBulletList 追加一组无序列表项（均为顶层级别）。
func (d *Document) AddBulletList(items ...string) {
	for _, it := range items {
		d.body = append(d.body, listParagraph{runs: []Run{{Text: it}}, numID: numBulletID})
	}
}

// AddOrderedList 追加一组有序列表项（均为顶层级别）。
func (d *Document) AddOrderedList(items ...string) {
	for _, it := range items {
		d.body = append(d.body, listParagraph{runs: []Run{{Text: it}}, numID: numOrderedID})
	}
}

// AddListItem 追加单个列表项，可指定有序/无序、缩进级别(0~3)与字符样式，便于构造多级嵌套列表。
func (d *Document) AddListItem(text string, ordered bool, level int, style RunStyle) {
	numID := numBulletID
	if ordered {
		numID = numOrderedID
	}
	d.body = append(d.body, listParagraph{runs: []Run{{Text: text, Style: style}}, numID: numID, level: level})
}
