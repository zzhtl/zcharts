package docx

import (
	"bytes"
	"fmt"
)

// numberingContentType 是 word/numbering.xml 的 OOXML content-type。
const numberingContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"

// 列表的 numId：无序列表用 numBulletID，有序列表用 numOrderedID。
// abstractNumId 与之一一对应。
const (
	numBulletID  = 1
	numOrderedID = 2
)

// numberingXML 生成 word/numbering.xml：一个无序（项目符号）和一个有序（十进制）列表定义，
// 各支持 0~3 级缩进。
func numberingXML() []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)

	// abstractNum 0：无序（项目符号）。
	b.WriteString(`<w:abstractNum w:abstractNumId="0">`)
	for lvl := 0; lvl < 4; lvl++ {
		fmt.Fprintf(&b, `<w:lvl w:ilvl="%d"><w:start w:val="1"/><w:numFmt w:val="bullet"/>`, lvl)
		fmt.Fprintf(&b, `<w:lvlText w:val="%s"/><w:lvlJc w:val="left"/>`, bulletGlyph(lvl))
		fmt.Fprintf(&b, `<w:pPr><w:ind w:left="%d" w:hanging="360"/></w:pPr></w:lvl>`, 720+lvl*420)
	}
	b.WriteString(`</w:abstractNum>`)

	// abstractNum 1：有序（十进制）。
	b.WriteString(`<w:abstractNum w:abstractNumId="1">`)
	for lvl := 0; lvl < 4; lvl++ {
		fmt.Fprintf(&b, `<w:lvl w:ilvl="%d"><w:start w:val="1"/><w:numFmt w:val="decimal"/>`, lvl)
		fmt.Fprintf(&b, `<w:lvlText w:val="%%%d."/><w:lvlJc w:val="left"/>`, lvl+1)
		fmt.Fprintf(&b, `<w:pPr><w:ind w:left="%d" w:hanging="360"/></w:pPr></w:lvl>`, 720+lvl*420)
	}
	b.WriteString(`</w:abstractNum>`)

	// num 实例。
	fmt.Fprintf(&b, `<w:num w:numId="%d"><w:abstractNumId w:val="0"/></w:num>`, numBulletID)
	fmt.Fprintf(&b, `<w:num w:numId="%d"><w:abstractNumId w:val="1"/></w:num>`, numOrderedID)

	b.WriteString(`</w:numbering>`)
	return b.Bytes()
}

// bulletGlyph 返回各级无序列表的项目符号字形（直接用 Unicode 字符，避免 Symbol 字体依赖）。
func bulletGlyph(lvl int) string {
	switch lvl % 3 {
	case 0:
		return "•" // •
	case 1:
		return "◦" // ◦
	default:
		return "▪" // ▪
	}
}
