package docx

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/zzhtl/zcharts/chart"
	"github.com/zzhtl/zcharts/docx/ooxml"
	"github.com/zzhtl/zcharts/option"
)

// Headings 返回已有文档中的 1~6 级标题目录。
func (d *EditableDocument) Headings() ([]Heading, error) {
	if d == nil {
		return nil, errors.New("docx: nil editable document")
	}
	body, err := d.parseBody()
	if err != nil {
		return nil, err
	}
	headings := make([]Heading, 0)
	for _, child := range body.children {
		if child.heading != nil {
			headings = append(headings, *child.heading)
		}
	}
	return headings, nil
}

// InsertHeadingUnderIndex 在指定标题目录项管辖的章节末尾插入同级标题。
func (d *EditableDocument) InsertHeadingUnderIndex(parentIndex int, text string) error {
	parent, insertAt, err := d.sectionInsertPoint(parentIndex)
	if err != nil {
		return err
	}
	styleID := d.headingStyleID(parent.Level)
	return d.insertBodyXML(insertAt, headingParagraphXML(text, styleID))
}

// InsertHeadingUnder 在第一个文本完全匹配的目标标题下插入同级标题。
func (d *EditableDocument) InsertHeadingUnder(parentText, text string) error {
	headings, err := d.Headings()
	if err != nil {
		return err
	}
	for _, h := range headings {
		if h.Text == parentText {
			return d.InsertHeadingUnderIndex(h.Index, text)
		}
	}
	return fmt.Errorf("docx: heading %q not found", parentText)
}

// InsertContentUnderIndex 把内容块追加到指定标题目录项管辖的章节末尾。
func (d *EditableDocument) InsertContentUnderIndex(parentIndex int, block *ContentBlock) error {
	_, insertAt, err := d.sectionInsertPoint(parentIndex)
	if err != nil {
		return err
	}
	xml, err := d.importContentBlock(block)
	if err != nil {
		return err
	}
	return d.insertBodyXML(insertAt, xml)
}

// InsertContentUnder 把内容块追加到第一个文本完全匹配的目标标题章节末尾。
func (d *EditableDocument) InsertContentUnder(parentText string, block *ContentBlock) error {
	headings, err := d.Headings()
	if err != nil {
		return err
	}
	for _, h := range headings {
		if h.Text == parentText {
			return d.InsertContentUnderIndex(h.Index, block)
		}
	}
	return fmt.Errorf("docx: heading %q not found", parentText)
}

// InsertParagraphUnderIndex 向指定标题章节末尾追加普通文字段落。
func (d *EditableDocument) InsertParagraphUnderIndex(parentIndex int, text string) error {
	block := NewContentBlock()
	block.AddParagraph(text)
	return d.InsertContentUnderIndex(parentIndex, block)
}

// InsertRichParagraphUnderIndex 向指定标题章节末尾追加富文本段落。
func (d *EditableDocument) InsertRichParagraphUnderIndex(parentIndex int, p RichParagraph) error {
	block := NewContentBlock()
	block.AddRichParagraph(p)
	return d.InsertContentUnderIndex(parentIndex, block)
}

// InsertChartUnderIndex 向指定标题章节末尾追加一个图表。
func (d *EditableDocument) InsertChartUnderIndex(parentIndex int, opt *option.Option, ins ChartInsertOption, renderOpts ...chart.RenderOption) error {
	block := NewContentBlock()
	if err := block.AddChart(opt, ins, renderOpts...); err != nil {
		return err
	}
	return d.InsertContentUnderIndex(parentIndex, block)
}

// InsertTableUnderIndex 向指定标题章节末尾追加一个表格。
func (d *EditableDocument) InsertTableUnderIndex(parentIndex int, t Table) error {
	block := NewContentBlock()
	if err := block.AddTable(t); err != nil {
		return err
	}
	return d.InsertContentUnderIndex(parentIndex, block)
}

// parseBody 把当前文档的 body 解析为 parsedBody，附带读取 styles.xml 的标题级别映射。
func (d *EditableDocument) parseBody() (parsedBody, error) {
	docXML, ok := d.files["word/document.xml"]
	if !ok {
		return parsedBody{}, errors.New("docx: missing word/document.xml")
	}
	styleLevels := headingStyleLevels(d.files["word/styles.xml"])
	return parseBodyXML(docXML, styleLevels)
}

// sectionInsertPoint 计算指定标题章节末尾的字节偏移。
// 章节边界定义：遇到第一个同级或更高级的标题、或 <w:sectPr> 立即停下；
// 都没有则用 </w:body> 的起点。返回的 Heading 是目标父标题本身，供调用方读级别。
func (d *EditableDocument) sectionInsertPoint(parentIndex int) (Heading, int, error) {
	if d == nil {
		return Heading{}, 0, errors.New("docx: nil editable document")
	}
	body, err := d.parseBody()
	if err != nil {
		return Heading{}, 0, err
	}

	parentChild := -1
	var parent Heading
	for i, child := range body.children {
		if child.heading != nil && child.heading.Index == parentIndex {
			parentChild = i
			parent = *child.heading
			break
		}
	}
	if parentChild < 0 {
		return Heading{}, 0, fmt.Errorf("docx: heading index %d not found", parentIndex)
	}

	insertAt := body.bodyEndStart
	for i := parentChild + 1; i < len(body.children); i++ {
		child := body.children[i]
		if child.heading != nil && child.heading.Level <= parent.Level {
			insertAt = child.start
			break
		}
		if child.name.Local == "sectPr" {
			insertAt = child.start
			break
		}
	}
	return parent, insertAt, nil
}

// insertBodyXML 在 word/document.xml 的指定字节偏移处插入一段 body XML 片段。
// 不做语法校验，调用方需保证 bodyXML 是合法的 OOXML 子树。
func (d *EditableDocument) insertBodyXML(insertAt int, bodyXML string) error {
	docXML := d.files["word/document.xml"]
	if insertAt < 0 || insertAt > len(docXML) {
		return fmt.Errorf("docx: invalid insert offset %d", insertAt)
	}
	updated := make([]byte, 0, len(docXML)+len(bodyXML))
	updated = append(updated, docXML[:insertAt]...)
	updated = append(updated, bodyXML...)
	updated = append(updated, docXML[insertAt:]...)
	d.files["word/document.xml"] = updated
	return nil
}

// importContentBlock 把内容块中的图片和图表搬入当前 docx 包：
// 写入 word/media/ 与 word/charts/ 部件、注册 Relationship、补齐 Content-Types，
// 并把 body XML 中的临时 rid 占位替换为真正分配的 Relationship Id，返回可直接插入的 body XML。
func (d *EditableDocument) importContentBlock(block *ContentBlock) (string, error) {
	if block == nil {
		return "", errors.New("docx: nil content block")
	}
	bodyXML := block.xml()
	usedRelIDs := documentRelIDs(d.files["word/_rels/document.xml.rels"])

	for _, img := range block.doc.images {
		filename := nextPartFilename(d.files, "word/media", "image", img.ext)
		rid := nextRelID(usedRelIDs, "rIdImage")
		usedRelIDs[rid] = true

		bodyXML = strings.ReplaceAll(bodyXML, img.rid, rid)
		d.files["word/media/"+filename] = img.data
		addDocumentRelationship(d.files, ooxml.Relationship{
			ID:     rid,
			Type:   ooxml.RelImage,
			Target: "media/" + filename,
		})
		ensureDefaultContentType(d.files, img.ext, img.mime)
	}

	for _, ch := range block.doc.charts {
		filename := nextPartFilename(d.files, "word/charts", "chart", "xml")
		rid := nextRelID(usedRelIDs, "rIdChart")
		usedRelIDs[rid] = true

		bodyXML = strings.ReplaceAll(bodyXML, ch.rid, rid)
		d.files["word/charts/"+filename] = ch.data
		addDocumentRelationship(d.files, ooxml.Relationship{
			ID:     rid,
			Type:   ooxml.RelChart,
			Target: "charts/" + filename,
		})
		ensureOverrideContentType(d.files, "/word/charts/"+filename, "application/vnd.openxmlformats-officedocument.drawingml.chart+xml")
	}

	return bodyXML, nil
}

// headingStyleID 给指定级别挑一个样式 id 用于新插入的标题段落。
// 优先用标准的 "HeadingN"；若 styles.xml 未声明该样式（或级别对不上），
// 退而求其次：从同级别的其它样式中按 id 字典序取第一个；仍然没有时返回 "HeadingN" 作为默认值。
func (d *EditableDocument) headingStyleID(level int) string {
	want := fmt.Sprintf("Heading%d", level)
	actual := parseStyleHeadingLevels(d.files["word/styles.xml"])
	if actual[want] == level {
		return want
	}
	ids := make([]string, 0, len(actual))
	for id, styleLevel := range actual {
		if styleLevel == level {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if len(ids) > 0 {
		return ids[0]
	}
	return want
}

// headingParagraphXML 生成一个带 pStyle 的标题段落 XML，文本和样式 id 都做 XML 转义。
func headingParagraphXML(text, styleID string) string {
	return `<w:p><w:pPr><w:pStyle w:val="` + xmlEscape(styleID) + `"/></w:pPr><w:r><w:t xml:space="preserve">` + xmlEscape(text) + `</w:t></w:r></w:p>`
}
