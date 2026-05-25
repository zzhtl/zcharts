package docx

import (
	"fmt"
	"strings"

	"github.com/zzhtl/zcharts/chart"
	"github.com/zzhtl/zcharts/option"
)

// Table 是一个富表格。零值不可用，至少需要一行。
type Table struct {
	Rows         []Row
	ColumnWidths []int // 各列宽度，单位 twips（1/20 磅）；为空则按整表宽度均分
	Width        int   // 整表宽度 twips；0 表示占满页面宽度（100%）
}

// Row 是表格的一行。
type Row struct {
	Cells  []Cell
	Height int  // 行高 twips；0 表示自动
	Header bool // 表头行：跨页重复，且未指定 Style 时默认加粗
}

// Cell 是表格单元格，支持富文本、内嵌图表、合并与样式。
type Cell struct {
	// 简单文本内容（最常用）。设置 Paragraphs 后忽略 Text/Style/Align。
	Text  string
	Style RunStyle // Text 的字符样式
	Align string   // 段落对齐："left"/"center"/"right"/"both"

	// 进阶：多段落富文本。
	Paragraphs []RichParagraph

	// 单元格内嵌图表（位于文本之后）。
	Chart *CellChart

	// 单元格样式
	Shading  string // 背景填充色 "#RRGGBB"
	VAlign   string // 垂直对齐："top"/"center"/"bottom"
	GridSpan int    // 跨列数；>1 生效
	VMerge   string // "restart"（合并起点）/"continue"（被合并），空为普通单元格
}

// CellChart 描述要嵌入单元格的图表。
type CellChart struct {
	Option     *option.Option
	Insert     ChartInsertOption // 用 AsImage / AsNativeChart 构造
	RenderOpts []chart.RenderOption
}

// rawElement 持有一段已生成好的 body XML。
type rawElement struct{ body string }

func (e rawElement) xml(_ *Document) string { return e.body }

// AddTable 在文档末尾追加一个表格。
// 单元格内嵌图表会在此处完成渲染与媒体注册，渲染失败时返回错误且不追加表格。
func (d *Document) AddTable(t Table) error {
	ncols := tableColumnCount(t)
	widths := tableColumnWidths(t, ncols)

	var b strings.Builder
	b.WriteString("<w:tbl>")
	b.WriteString(tablePropsXML(t, widths))
	// tblGrid
	b.WriteString("<w:tblGrid>")
	for _, w := range widths {
		fmt.Fprintf(&b, `<w:gridCol w:w="%d"/>`, w)
	}
	b.WriteString("</w:tblGrid>")

	for _, row := range t.Rows {
		b.WriteString("<w:tr>")
		b.WriteString(rowPropsXML(row))
		col := 0
		for _, cell := range row.Cells {
			span := cell.GridSpan
			if span < 1 {
				span = 1
			}
			cw := 0
			for i := 0; i < span && col+i < len(widths); i++ {
				cw += widths[col+i]
			}
			col += span
			content, err := d.cellContentXML(cell, row, cw)
			if err != nil {
				return err
			}
			b.WriteString("<w:tc>")
			b.WriteString(cellPropsXML(cell, cw))
			b.WriteString(content)
			b.WriteString("</w:tc>")
		}
		b.WriteString("</w:tr>")
	}
	b.WriteString("</w:tbl>")
	// OOXML 要求表格后必须有一个块级元素，补一个空段落（同时分隔相邻表格）。
	b.WriteString("<w:p/>")

	d.body = append(d.body, rawElement{body: b.String()})
	return nil
}

// cellContentXML 生成单元格内部的块级内容（至少一个 <w:p>）。
func (d *Document) cellContentXML(cell Cell, row Row, widthTwips int) (string, error) {
	var parts []string
	switch {
	case len(cell.Paragraphs) > 0:
		for _, p := range cell.Paragraphs {
			parts = append(parts, p.xml(d))
		}
	case cell.Text != "":
		style := cell.Style
		if row.Header && !style.Bold {
			style.Bold = true
		}
		parts = append(parts, RichParagraph{
			Runs: []Run{{Text: cell.Text, Style: style}},
			Para: ParaStyle{Align: cell.Align},
		}.xml(d))
	}
	if cell.Chart != nil {
		insert := fitChartInsertToCell(cell.Chart.Insert, widthTwips)
		el, err := d.buildChartElement(cell.Chart.Option, insert, cell.Chart.RenderOpts...)
		if err != nil {
			return "", err
		}
		parts = append(parts, el.xml(d))
	}
	if len(parts) == 0 {
		parts = append(parts, "<w:p/>")
	}
	return strings.Join(parts, ""), nil
}

func fitChartInsertToCell(ins ChartInsertOption, widthTwips int) ChartInsertOption {
	if ins.Width <= 0 || widthTwips <= 0 {
		return ins
	}
	const (
		twipsPerPx       = 15
		cellPaddingTwips = 240
	)
	usable := widthTwips - cellPaddingTwips
	if usable <= 0 {
		return ins
	}
	maxWidth := usable / twipsPerPx
	if maxWidth <= 0 || ins.Width <= maxWidth {
		return ins
	}
	if ins.Height > 0 {
		ins.Height = maxInt(1, ins.Height*maxWidth/ins.Width)
	}
	ins.Width = maxWidth
	return ins
}

// tableColumnCount 计算表格列数（取各行 gridSpan 之和的最大值）。
func tableColumnCount(t Table) int {
	max := len(t.ColumnWidths)
	for _, row := range t.Rows {
		n := 0
		for _, c := range row.Cells {
			s := c.GridSpan
			if s < 1 {
				s = 1
			}
			n += s
		}
		if n > max {
			max = n
		}
	}
	if max < 1 {
		max = 1
	}
	return max
}

// tableColumnWidths 返回各列宽度（twips）。
func tableColumnWidths(t Table, ncols int) []int {
	if len(t.ColumnWidths) == ncols {
		return t.ColumnWidths
	}
	total := t.Width
	if total <= 0 {
		total = 9350 // 约等于 A4/Letter 默认页边距内的可打印宽度
	}
	widths := make([]int, ncols)
	each := total / ncols
	for i := range widths {
		widths[i] = each
	}
	return widths
}

// tablePropsXML 生成 <w:tblPr>，含整表宽度与统一的单线边框。
func tablePropsXML(t Table, widths []int) string {
	var b strings.Builder
	b.WriteString("<w:tblPr>")
	if t.Width > 0 {
		fmt.Fprintf(&b, `<w:tblW w:w="%d" w:type="dxa"/>`, t.Width)
	} else {
		b.WriteString(`<w:tblW w:w="5000" w:type="pct"/>`)
	}
	// 统一单线灰色边框。
	b.WriteString(`<w:tblBorders>`)
	for _, edge := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		fmt.Fprintf(&b, `<w:%s w:val="single" w:sz="4" w:space="0" w:color="BFBFBF"/>`, edge)
	}
	b.WriteString(`</w:tblBorders>`)
	b.WriteString(`<w:tblLook w:val="04A0" w:firstRow="1" w:lastRow="0" w:firstColumn="1" w:lastColumn="0" w:noHBand="0" w:noVBand="1"/>`)
	b.WriteString("</w:tblPr>")
	return b.String()
}

// rowPropsXML 生成 <w:trPr>；无属性时返回空串。
func rowPropsXML(row Row) string {
	var b strings.Builder
	if row.Height > 0 {
		fmt.Fprintf(&b, `<w:trHeight w:val="%d"/>`, row.Height)
	}
	if row.Header {
		b.WriteString(`<w:tblHeader/>`)
	}
	if b.Len() == 0 {
		return ""
	}
	return "<w:trPr>" + b.String() + "</w:trPr>"
}

// cellPropsXML 生成 <w:tcPr>，字段顺序遵循 OOXML CT_TcPr。
func cellPropsXML(cell Cell, width int) string {
	var b strings.Builder
	b.WriteString("<w:tcPr>")
	if width > 0 {
		fmt.Fprintf(&b, `<w:tcW w:w="%d" w:type="dxa"/>`, width)
	}
	if cell.GridSpan > 1 {
		fmt.Fprintf(&b, `<w:gridSpan w:val="%d"/>`, cell.GridSpan)
	}
	switch strings.ToLower(cell.VMerge) {
	case "restart":
		b.WriteString(`<w:vMerge w:val="restart"/>`)
	case "continue":
		b.WriteString(`<w:vMerge/>`)
	}
	if fill := normalizeColor(cell.Shading); fill != "" {
		fmt.Fprintf(&b, `<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, fill)
	}
	if v := cellVAlign(cell.VAlign); v != "" {
		fmt.Fprintf(&b, `<w:vAlign w:val="%s"/>`, v)
	}
	b.WriteString("</w:tcPr>")
	return b.String()
}

func cellVAlign(v string) string {
	switch strings.ToLower(v) {
	case "top":
		return "top"
	case "center", "middle":
		return "center"
	case "bottom":
		return "bottom"
	default:
		return ""
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
