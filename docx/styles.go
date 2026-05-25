package docx

import (
	"bytes"
	"fmt"
)

// stylesContentType 是 word/styles.xml 的 OOXML content-type。
const stylesContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"

// headingStyle 描述一个标题样式的字号（半磅）与颜色（不带 # 的 6 位 hex）。
type headingStyle struct {
	id    string
	name  string
	sizeH int // 半磅，即磅 ×2
	color string
}

// builtinHeadings 是内置的 1~6 级标题样式定义，styleId 必须与 AddHeading 引用一致。
var builtinHeadings = []headingStyle{
	{"Heading1", "heading 1", 40, "2E5496"},
	{"Heading2", "heading 2", 32, "2E74B5"},
	{"Heading3", "heading 3", 28, "1F4E79"},
	{"Heading4", "heading 4", 24, "2E74B5"},
	{"Heading5", "heading 5", 22, "2E74B5"},
	{"Heading6", "heading 6", 20, "404040"},
}

// stylesXML 生成 word/styles.xml：包含默认正文样式、文档标题 Title 以及 1~6 级标题。
func stylesXML() []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)

	// 文档默认：正文 11pt（sz=22）。
	b.WriteString(`<w:docDefaults><w:rPrDefault><w:rPr><w:sz w:val="22"/><w:szCs w:val="22"/></w:rPr></w:rPrDefault></w:docDefaults>`)

	// Normal 正文样式。
	b.WriteString(`<w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/></w:style>`)

	// Title 文档大标题。
	b.WriteString(`<w:style w:type="paragraph" w:styleId="Title">`)
	b.WriteString(`<w:name w:val="Title"/>`)
	b.WriteString(`<w:pPr><w:spacing w:before="240" w:after="120"/><w:jc w:val="center"/></w:pPr>`)
	b.WriteString(`<w:rPr><w:b/><w:sz w:val="52"/><w:szCs w:val="52"/><w:color w:val="1F3864"/></w:rPr>`)
	b.WriteString(`</w:style>`)

	// Heading1..6。
	for _, h := range builtinHeadings {
		fmt.Fprintf(&b, `<w:style w:type="paragraph" w:styleId="%s">`, h.id)
		fmt.Fprintf(&b, `<w:name w:val="%s"/>`, h.name)
		b.WriteString(`<w:pPr><w:keepNext/><w:spacing w:before="200" w:after="80"/><w:outlineLvl w:val="`)
		fmt.Fprintf(&b, `%d`, headingOutlineLvl(h.id))
		b.WriteString(`"/></w:pPr>`)
		fmt.Fprintf(&b, `<w:rPr><w:b/><w:sz w:val="%d"/><w:szCs w:val="%d"/><w:color w:val="%s"/></w:rPr>`, h.sizeH, h.sizeH, h.color)
		b.WriteString(`</w:style>`)
	}

	b.WriteString(`</w:styles>`)
	return b.Bytes()
}

// headingOutlineLvl 把 HeadingN 转成 0 起的 outline 级别。
func headingOutlineLvl(id string) int {
	for i, h := range builtinHeadings {
		if h.id == id {
			return i
		}
	}
	return 0
}
