package docx

import (
	"fmt"
	"strings"
)

// RunStyle 描述一段文字（run）的字符级格式。零值表示继承默认样式。
type RunStyle struct {
	Font      string  // 字体名（中英文同时设置）
	Size      float64 // 磅；0 表示继承
	Bold      bool
	Italic    bool
	Underline bool
	Strike    bool   // 删除线
	Color     string // "#RRGGBB" 或 "RRGGBB"；空表示继承
}

// Run 是富文本段落中的一段连续同样式文字。
type Run struct {
	Text  string
	Style RunStyle
}

// ParaStyle 描述段落级格式。
type ParaStyle struct {
	StyleID       string // 段落样式 id，如 "Heading1"/"Title"；空表示正文
	Align         string // "left"/"center"/"right"/"both"；空表示默认
	IndentLeft    int    // 左缩进，单位 twips（1/20 磅）
	SpacingBefore int    // 段前间距，twips
	SpacingAfter  int    // 段后间距，twips
}

// RichParagraph 是带字符/段落格式的段落，由多个 Run 组成。
type RichParagraph struct {
	Runs []Run
	Para ParaStyle
}

func (p RichParagraph) xml(_ *Document) string {
	var b strings.Builder
	b.WriteString("<w:p>")
	b.WriteString(paraPropsXML(p.Para))
	for _, r := range p.Runs {
		b.WriteString(runXML(r))
	}
	b.WriteString("</w:p>")
	return b.String()
}

// AddRichParagraph 追加一个富文本段落。
func (d *Document) AddRichParagraph(p RichParagraph) {
	d.body = append(d.body, p)
}

// AddHeading 追加一个标题段落，level 取 1~6，超出范围会被截断到该区间。
func (d *Document) AddHeading(text string, level int) {
	if level < 1 {
		level = 1
	}
	if level > len(builtinHeadings) {
		level = len(builtinHeadings)
	}
	d.body = append(d.body, RichParagraph{
		Runs: []Run{{Text: text}},
		Para: ParaStyle{StyleID: fmt.Sprintf("Heading%d", level)},
	})
}

// AddTitle 追加文档大标题（居中、加粗、大字号）。
func (d *Document) AddTitle(text string) {
	d.body = append(d.body, RichParagraph{
		Runs: []Run{{Text: text}},
		Para: ParaStyle{StyleID: "Title"},
	})
}

// AddStyledParagraph 追加一个统一字符样式的正文段落（便捷封装）。
func (d *Document) AddStyledParagraph(text string, style RunStyle, align string) {
	d.body = append(d.body, RichParagraph{
		Runs: []Run{{Text: text, Style: style}},
		Para: ParaStyle{Align: align},
	})
}

// runXML 生成单个 <w:r>。
func runXML(r Run) string {
	return "<w:r>" + runPropsXML(r.Style) + `<w:t xml:space="preserve">` + xmlEscape(r.Text) + "</w:t></w:r>"
}

// runPropsXML 生成 <w:rPr>；无任何样式时返回空串。
func runPropsXML(s RunStyle) string {
	var b strings.Builder
	if s.Font != "" {
		f := xmlEscape(s.Font)
		fmt.Fprintf(&b, `<w:rFonts w:ascii="%s" w:hAnsi="%s" w:eastAsia="%s" w:cs="%s"/>`, f, f, f, f)
	}
	if s.Bold {
		b.WriteString("<w:b/>")
	}
	if s.Italic {
		b.WriteString("<w:i/>")
	}
	if s.Strike {
		b.WriteString("<w:strike/>")
	}
	if s.Underline {
		b.WriteString(`<w:u w:val="single"/>`)
	}
	if c := normalizeColor(s.Color); c != "" {
		fmt.Fprintf(&b, `<w:color w:val="%s"/>`, c)
	}
	if s.Size > 0 {
		half := int(s.Size * 2)
		fmt.Fprintf(&b, `<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, half, half)
	}
	if b.Len() == 0 {
		return ""
	}
	return "<w:rPr>" + b.String() + "</w:rPr>"
}

// paraPropsXML 生成 <w:pPr>；无任何属性时返回空串。
func paraPropsXML(p ParaStyle) string {
	var b strings.Builder
	if p.StyleID != "" {
		fmt.Fprintf(&b, `<w:pStyle w:val="%s"/>`, xmlEscape(p.StyleID))
	}
	if p.SpacingBefore > 0 || p.SpacingAfter > 0 {
		fmt.Fprintf(&b, `<w:spacing w:before="%d" w:after="%d"/>`, p.SpacingBefore, p.SpacingAfter)
	}
	if p.IndentLeft > 0 {
		fmt.Fprintf(&b, `<w:ind w:left="%d"/>`, p.IndentLeft)
	}
	if jc := alignToJc(p.Align); jc != "" {
		fmt.Fprintf(&b, `<w:jc w:val="%s"/>`, jc)
	}
	if b.Len() == 0 {
		return ""
	}
	return "<w:pPr>" + b.String() + "</w:pPr>"
}

// alignToJc 把对外的对齐名转成 OOXML jc 值。
func alignToJc(align string) string {
	switch strings.ToLower(align) {
	case "center":
		return "center"
	case "right", "end":
		return "right"
	case "both", "justify":
		return "both"
	case "left", "start":
		return "left"
	default:
		return ""
	}
}

// normalizeColor 把 "#RRGGBB"/"RRGGBB" 规整为 6 位大写 hex（不带 #）；非法或空时返回空串。
func normalizeColor(c string) string {
	c = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(c), "#"))
	if len(c) != 6 {
		return ""
	}
	return strings.ToUpper(c)
}
