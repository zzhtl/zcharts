package docx

import (
	"bytes"
	"image/png"
	"strings"
	"testing"

	"github.com/zzhtl/zcharts/jsonopt"
)

// buildAndUnzip 写出文档并解压，返回各部件字节。
func buildAndUnzip(t *testing.T, d *Document) map[string][]byte {
	t.Helper()
	var buf bytes.Buffer
	if err := d.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	return unzipDocx(t, buf.Bytes())
}

func TestStylesAndNumberingParts(t *testing.T) {
	d := New()
	d.AddParagraph("正文")
	files := buildAndUnzip(t, d)

	for _, name := range []string{"word/styles.xml", "word/numbering.xml"} {
		data, ok := files[name]
		if !ok {
			t.Fatalf("缺少部件 %s", name)
		}
		assertXML(t, data)
	}
	styles := string(files["word/styles.xml"])
	for _, id := range []string{"Title", "Heading1", "Heading6"} {
		if !strings.Contains(styles, `w:styleId="`+id+`"`) {
			t.Errorf("styles.xml 缺少样式 %s", id)
		}
	}

	rels := string(files["word/_rels/document.xml.rels"])
	if !strings.Contains(rels, "styles.xml") || !strings.Contains(rels, "numbering.xml") {
		t.Errorf("document.xml.rels 缺少 styles/numbering 关系:\n%s", rels)
	}
}

func TestHeadingAndRichParagraph(t *testing.T) {
	d := New()
	d.AddTitle("报告标题")
	d.AddHeading("一级标题", 1)
	d.AddHeading("超界标题", 99) // 应被截断到 Heading6
	d.AddRichParagraph(RichParagraph{
		Runs: []Run{
			{Text: "加粗红字", Style: RunStyle{Bold: true, Color: "#FF0000", Size: 14}},
			{Text: " 普通", Style: RunStyle{}},
		},
		Para: ParaStyle{Align: "center"},
	})
	files := buildAndUnzip(t, d)
	doc := string(files["word/document.xml"])
	assertXML(t, files["word/document.xml"])

	for _, frag := range []string{
		`<w:pStyle w:val="Title"/>`,
		`<w:pStyle w:val="Heading1"/>`,
		`<w:pStyle w:val="Heading6"/>`,
		`<w:b/>`,
		`<w:color w:val="FF0000"/>`,
		`<w:sz w:val="28"/>`, // 14pt → 28 半磅
		`<w:jc w:val="center"/>`,
	} {
		if !strings.Contains(doc, frag) {
			t.Errorf("document.xml 缺少片段 %q", frag)
		}
	}
}

func TestLists(t *testing.T) {
	d := New()
	d.AddBulletList("苹果", "香蕉")
	d.AddOrderedList("第一", "第二")
	d.AddListItem("二级项", false, 1, RunStyle{Italic: true})
	files := buildAndUnzip(t, d)
	doc := string(files["word/document.xml"])
	assertXML(t, files["word/document.xml"])

	if !strings.Contains(doc, `<w:numId w:val="1"/>`) {
		t.Error("缺少无序列表 numId=1")
	}
	if !strings.Contains(doc, `<w:numId w:val="2"/>`) {
		t.Error("缺少有序列表 numId=2")
	}
	if !strings.Contains(doc, `<w:ilvl w:val="1"/>`) {
		t.Error("缺少二级列表缩进 ilvl=1")
	}
}

func TestTableRichWithMergeAndShading(t *testing.T) {
	d := New()
	err := d.AddTable(Table{
		Rows: []Row{
			{
				Header: true,
				Cells: []Cell{
					{Text: "合并表头", GridSpan: 2, Align: "center", Shading: "#4472C4", Style: RunStyle{Color: "#FFFFFF"}},
				},
			},
			{
				Cells: []Cell{
					{Text: "左上", VMerge: "restart"},
					{Text: "右上"},
				},
			},
			{
				Cells: []Cell{
					{VMerge: "continue"},
					{Text: "右下"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	files := buildAndUnzip(t, d)
	doc := string(files["word/document.xml"])
	assertXML(t, files["word/document.xml"])

	for _, frag := range []string{
		"<w:tbl>",
		"<w:tblGrid>",
		`<w:gridSpan w:val="2"/>`,
		`<w:vMerge w:val="restart"/>`,
		`<w:vMerge/>`,
		`<w:shd w:val="clear" w:color="auto" w:fill="4472C4"/>`,
		`<w:tblHeader/>`,
	} {
		if !strings.Contains(doc, frag) {
			t.Errorf("document.xml 缺少表格片段 %q", frag)
		}
	}
	// 表格后应补空段落。
	if !strings.Contains(doc, "</w:tbl><w:p/>") {
		t.Error("表格后缺少分隔空段落")
	}
}

func TestTableWithEmbeddedChart(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
		"xAxis": {"type":"category","data":["A","B","C"]},
		"yAxis": {"type":"value"},
		"series": [{"type":"bar","data":[1,2,3]}]
	}`)
	if err != nil {
		t.Fatalf("parse option: %v", err)
	}
	d := New()
	if err := d.AddTable(Table{
		Rows: []Row{
			{Cells: []Cell{
				{Text: "指标"},
				{Chart: &CellChart{Option: opt, Insert: AsImage(PNG, 240, 160)}},
			}},
		},
	}); err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	files := buildAndUnzip(t, d)
	assertXML(t, files["word/document.xml"])

	// 单元格内嵌图表应注册一张图片并在文档中引用。
	if _, ok := files["word/media/image1.png"]; !ok {
		t.Error("缺少单元格内嵌图表渲染的图片 media/image1.png")
	}
	doc := string(files["word/document.xml"])
	if !strings.Contains(doc, "<w:tc>") || !strings.Contains(doc, "<w:drawing>") {
		t.Error("表格单元格内未嵌入 drawing")
	}
}

func TestTableEmbeddedChartFitsColumnWidth(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
		"xAxis": {"type":"category","data":["A","B","C"]},
		"yAxis": {"type":"value"},
		"series": [{"type":"bar","data":[1,2,3]}]
	}`)
	if err != nil {
		t.Fatalf("parse option: %v", err)
	}
	d := New()
	if err := d.AddTable(Table{
		Rows: []Row{
			{Cells: []Cell{
				{Text: "指标"},
				{Chart: &CellChart{Option: opt, Insert: AsImage(PNG, 320, 180)}},
			}},
		},
	}); err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	files := buildAndUnzip(t, d)
	doc := string(files["word/document.xml"])
	if strings.Contains(doc, `cx="3048000"`) {
		t.Fatalf("表格内图表不应保留超出列宽的 320px 宽度:\n%s", doc)
	}
	if !strings.Contains(doc, `cx="2809875" cy="1571625"`) {
		t.Fatalf("表格内图表未按列宽缩放:\n%s", doc)
	}

	img, err := png.Decode(bytes.NewReader(files["word/media/image1.png"]))
	if err != nil {
		t.Fatal(err)
	}
	if got := img.Bounds().Dx(); got != 295 {
		t.Fatalf("image width=%d", got)
	}
	if got := img.Bounds().Dy(); got != 165 {
		t.Fatalf("image height=%d", got)
	}
}
