package docx

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/zzhtl/zcharts/jsonopt"
)

func TestEditableDocumentHeadings(t *testing.T) {
	path := writeEditableDocx(t)

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	headings, err := doc.Headings()
	if err != nil {
		t.Fatalf("Headings: %v", err)
	}

	want := []Heading{
		{Index: 0, Level: 1, StyleID: "Heading1", Text: "一、整体概况"},
		{Index: 1, Level: 2, StyleID: "Heading2", Text: "1.1 销售趋势"},
		{Index: 2, Level: 3, StyleID: "Heading3", Text: "1.1.1 存量章节"},
		{Index: 3, Level: 2, StyleID: "Heading2", Text: "1.2 渠道占比"},
		{Index: 4, Level: 1, StyleID: "Heading1", Text: "二、结论"},
	}
	if len(headings) != len(want) {
		t.Fatalf("headings len=%d, want %d: %#v", len(headings), len(want), headings)
	}
	for i := range want {
		if headings[i] != want[i] {
			t.Fatalf("heading[%d]=%#v, want %#v", i, headings[i], want[i])
		}
	}
}

func TestEditableDocumentHeadingsNormalizeArbitraryStyleID(t *testing.T) {
	path := writeEditableDocx(t)

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	documentXML := string(doc.files["word/document.xml"])
	documentXML = strings.Replace(documentXML, `w:val="Heading1"`, `w:val="a0"`, 1)
	doc.files["word/document.xml"] = []byte(documentXML)

	stylesXML := string(doc.files["word/styles.xml"])
	stylesXML = strings.Replace(stylesXML, "</w:styles>",
		`<w:style w:type="paragraph" w:styleId="a0"><w:name w:val="custom title"/><w:pPr><w:outlineLvl w:val="0"/></w:pPr></w:style></w:styles>`, 1)
	doc.files["word/styles.xml"] = []byte(stylesXML)

	headings, err := doc.Headings()
	if err != nil {
		t.Fatalf("Headings: %v", err)
	}
	if got := headings[0]; got.Level != 1 || got.StyleID != "Heading1" || got.Text != "一、整体概况" {
		t.Fatalf("heading not normalized: %#v", got)
	}
	for _, h := range headings {
		if h.StyleID != "Heading"+strconv.Itoa(h.Level) {
			t.Fatalf("heading style id is not normalized: %#v", h)
		}
	}
}

func TestEditableDocumentInsertHeadingUnderIndexUsesTargetLevel(t *testing.T) {
	path := writeEditableDocx(t)

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := doc.InsertHeadingUnderIndex(2, "1.1.2 新增同级章节"); err != nil {
		t.Fatalf("InsertHeadingUnderIndex: %v", err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	files := unzipDocx(t, out.Bytes())
	documentXML := string(files["word/document.xml"])
	assertXML(t, files["word/document.xml"])

	assertOrder(t, documentXML, "1.1.1 存量章节", "存量正文", "1.1.2 新增同级章节", "1.2 渠道占比")
	if !strings.Contains(documentXML, `<w:pStyle w:val="Heading3"/>`) {
		t.Fatalf("inserted heading style missing: %s", documentXML)
	}

	reopenedPath := t.TempDir() + "/updated.docx"
	if err := os.WriteFile(reopenedPath, out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(reopenedPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	headings, err := reopened.Headings()
	if err != nil {
		t.Fatalf("reopened Headings: %v", err)
	}
	if got := headings[3]; got.Level != 3 || got.StyleID != "Heading3" || got.Text != "1.1.2 新增同级章节" {
		t.Fatalf("inserted heading=%#v", got)
	}
}

func TestEditableDocumentInsertHeadingUnderText(t *testing.T) {
	path := writeEditableDocx(t)

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := doc.InsertHeadingUnder("1.1 销售趋势", "1.1 新增同级章节"); err != nil {
		t.Fatalf("InsertHeadingUnder: %v", err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	files := unzipDocx(t, out.Bytes())
	documentXML := string(files["word/document.xml"])
	assertOrder(t, documentXML, "1.1 销售趋势", "1.1.1 存量章节", "1.1 新增同级章节", "1.2 渠道占比")
}

func TestEditableDocumentInsertContentBlockUnderIndex(t *testing.T) {
	path := writeEditableDocx(t)

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	opt, err := jsonopt.ParseString(testLineJSON)
	if err != nil {
		t.Fatalf("parse option: %v", err)
	}
	block := NewContentBlock()
	block.AddParagraph("新增正文")
	if err := block.AddTable(Table{
		Rows: []Row{
			{Header: true, Cells: []Cell{{Text: "指标"}, {Text: "值"}}},
			{Cells: []Cell{{Text: "新增表格"}, {Text: "42"}}},
		},
	}); err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	if err := block.AddChart(opt, AsImage(PNG, 120, 80)); err != nil {
		t.Fatalf("AddChart: %v", err)
	}
	if err := doc.InsertContentUnderIndex(1, block); err != nil {
		t.Fatalf("InsertContentUnderIndex: %v", err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	files := unzipDocx(t, out.Bytes())
	documentXML := string(files["word/document.xml"])
	assertXML(t, files["word/document.xml"])
	assertXML(t, files["word/_rels/document.xml.rels"])
	assertXML(t, files["[Content_Types].xml"])

	assertOrder(t, documentXML, "1.1 销售趋势", "1.1.1 存量章节", "新增正文", "新增表格", "1.2 渠道占比")
	if _, ok := files["word/media/image1.png"]; !ok {
		t.Fatal("missing inserted chart image")
	}
	relsXML := string(files["word/_rels/document.xml.rels"])
	if !strings.Contains(relsXML, `Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"`) ||
		!strings.Contains(relsXML, `Target="media/image1.png"`) {
		t.Fatalf("missing inserted image relationship: %s", relsXML)
	}
	contentTypes := string(files["[Content_Types].xml"])
	if !strings.Contains(contentTypes, `Extension="png" ContentType="image/png"`) {
		t.Fatalf("missing png content type: %s", contentTypes)
	}
}

func TestEditableDocumentInsertParagraphUnderIndex(t *testing.T) {
	path := writeEditableDocx(t)

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := doc.InsertParagraphUnderIndex(3, "渠道新增正文"); err != nil {
		t.Fatalf("InsertParagraphUnderIndex: %v", err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	files := unzipDocx(t, out.Bytes())
	documentXML := string(files["word/document.xml"])
	assertOrder(t, documentXML, "1.2 渠道占比", "渠道正文", "渠道新增正文", "二、结论")
}

func TestEditableDocumentHeadingsComplexHierarchy(t *testing.T) {
	path := writeComplexEditableDocx(t)

	doc, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	headings, err := doc.Headings()
	if err != nil {
		t.Fatalf("Headings: %v", err)
	}

	want := []Heading{
		{Index: 0, Level: 1, StyleID: "Heading1", Text: "一、总览"},
		{Index: 1, Level: 2, StyleID: "Heading2", Text: "2A 销售"},
		{Index: 2, Level: 3, StyleID: "Heading3", Text: "3A-1 国内"},
		{Index: 3, Level: 4, StyleID: "Heading4", Text: "4A-1 华东"},
		{Index: 4, Level: 5, StyleID: "Heading5", Text: "5A-1 上海"},
		{Index: 5, Level: 6, StyleID: "Heading6", Text: "6A-1 黄浦"},
		{Index: 6, Level: 6, StyleID: "Heading6", Text: "6A-2 浦东"},
		{Index: 7, Level: 5, StyleID: "Heading5", Text: "5A-2 杭州"},
		{Index: 8, Level: 6, StyleID: "Heading6", Text: "6A-3 西湖"},
		{Index: 9, Level: 4, StyleID: "Heading4", Text: "4A-2 华南"},
		{Index: 10, Level: 5, StyleID: "Heading5", Text: "5A-3 广州"},
		{Index: 11, Level: 3, StyleID: "Heading3", Text: "3A-2 海外"},
		{Index: 12, Level: 4, StyleID: "Heading4", Text: "4A-3 欧洲"},
		{Index: 13, Level: 5, StyleID: "Heading5", Text: "5A-4 德国"},
		{Index: 14, Level: 2, StyleID: "Heading2", Text: "2B 运营"},
		{Index: 15, Level: 3, StyleID: "Heading3", Text: "3B-1 流量"},
		{Index: 16, Level: 4, StyleID: "Heading4", Text: "4B-1 自然流量"},
		{Index: 17, Level: 5, StyleID: "Heading5", Text: "5B-1 SEO"},
		{Index: 18, Level: 6, StyleID: "Heading6", Text: "6B-1 内容SEO"},
		{Index: 19, Level: 3, StyleID: "Heading3", Text: "3B-2 转化"},
		{Index: 20, Level: 2, StyleID: "Heading2", Text: "2C 风险"},
		{Index: 21, Level: 1, StyleID: "Heading1", Text: "二、结论"},
	}
	if len(headings) != len(want) {
		t.Fatalf("headings len=%d, want %d: %#v", len(headings), len(want), headings)
	}
	for i := range want {
		if headings[i] != want[i] {
			t.Fatalf("heading[%d]=%#v, want %#v", i, headings[i], want[i])
		}
	}
}

func TestEditableDocumentInsertContentComplexHierarchyBoundaries(t *testing.T) {
	cases := []struct {
		name     string
		target   string
		inserted string
		order    []string
	}{
		{
			name:     "level2 includes all level3-6 children",
			target:   "2A 销售",
			inserted: "插入-2A内容",
			order:    []string{"5A-4 德国正文", "插入-2A内容", "2B 运营"},
		},
		{
			name:     "level3 stops before next level3 sibling",
			target:   "3A-1 国内",
			inserted: "插入-3A1内容",
			order:    []string{"5A-3 广州正文", "插入-3A1内容", "3A-2 海外"},
		},
		{
			name:     "level4 stops before next level4 sibling",
			target:   "4A-1 华东",
			inserted: "插入-4A1内容",
			order:    []string{"6A-3 西湖正文", "插入-4A1内容", "4A-2 华南"},
		},
		{
			name:     "level5 includes level6 children and stops before level5 sibling",
			target:   "5A-1 上海",
			inserted: "插入-5A1内容",
			order:    []string{"6A-2 浦东正文", "插入-5A1内容", "5A-2 杭州"},
		},
		{
			name:     "level6 stops before next level6 sibling",
			target:   "6A-1 黄浦",
			inserted: "插入-6A1内容",
			order:    []string{"6A-1 黄浦正文", "插入-6A1内容", "6A-2 浦东"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc, heading := openComplexDocAndFindHeading(t, tc.target)

			block := NewContentBlock()
			block.AddParagraph(tc.inserted)
			if err := doc.InsertContentUnderIndex(heading.Index, block); err != nil {
				t.Fatalf("InsertContentUnderIndex: %v", err)
			}

			files := writeEditableDocumentAndUnzip(t, doc)
			assertXML(t, files["word/document.xml"])
			assertOrder(t, string(files["word/document.xml"]), tc.order...)
		})
	}
}

func TestEditableDocumentInsertHeadingComplexHierarchyUsesSameLevelAndBoundary(t *testing.T) {
	cases := []struct {
		name       string
		target     string
		inserted   string
		wantLevel  int
		wantStyle  string
		order      []string
		reopenText string
	}{
		{
			name:       "insert same level4 heading after level4 subtree",
			target:     "4A-1 华东",
			inserted:   "4A-1 新增同级",
			wantLevel:  4,
			wantStyle:  "Heading4",
			order:      []string{"6A-3 西湖正文", "4A-1 新增同级", "4A-2 华南"},
			reopenText: "4A-1 新增同级",
		},
		{
			name:       "insert same level3 heading after level3 subtree",
			target:     "3A-1 国内",
			inserted:   "3A-1 新增同级",
			wantLevel:  3,
			wantStyle:  "Heading3",
			order:      []string{"5A-3 广州正文", "3A-1 新增同级", "3A-2 海外"},
			reopenText: "3A-1 新增同级",
		},
		{
			name:       "insert same level5 heading after level6 children",
			target:     "5A-1 上海",
			inserted:   "5A-1 新增同级",
			wantLevel:  5,
			wantStyle:  "Heading5",
			order:      []string{"6A-2 浦东正文", "5A-1 新增同级", "5A-2 杭州"},
			reopenText: "5A-1 新增同级",
		},
		{
			name:       "insert same level6 heading before next level6 sibling",
			target:     "6A-1 黄浦",
			inserted:   "6A-1 新增同级",
			wantLevel:  6,
			wantStyle:  "Heading6",
			order:      []string{"6A-1 黄浦正文", "6A-1 新增同级", "6A-2 浦东"},
			reopenText: "6A-1 新增同级",
		},
		{
			name:       "insert same level2 heading after all descendants",
			target:     "2B 运营",
			inserted:   "2B 新增同级",
			wantLevel:  2,
			wantStyle:  "Heading2",
			order:      []string{"3B-2 转化正文", "2B 新增同级", "2C 风险"},
			reopenText: "2B 新增同级",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc, heading := openComplexDocAndFindHeading(t, tc.target)

			if err := doc.InsertHeadingUnderIndex(heading.Index, tc.inserted); err != nil {
				t.Fatalf("InsertHeadingUnderIndex: %v", err)
			}

			files := writeEditableDocumentAndUnzip(t, doc)
			documentXML := string(files["word/document.xml"])
			assertXML(t, files["word/document.xml"])
			assertOrder(t, documentXML, tc.order...)
			if !strings.Contains(documentXML, `<w:pStyle w:val="`+tc.wantStyle+`"/>`) {
				t.Fatalf("inserted heading style %s missing: %s", tc.wantStyle, documentXML)
			}

			reopened := reopenEditableDocumentFromBytes(t, files)
			headings, err := reopened.Headings()
			if err != nil {
				t.Fatalf("reopened Headings: %v", err)
			}
			got := findHeadingByText(t, headings, tc.reopenText)
			if got.Level != tc.wantLevel || got.StyleID != tc.wantStyle {
				t.Fatalf("inserted heading=%#v, want level=%d style=%s", got, tc.wantLevel, tc.wantStyle)
			}
		})
	}
}

func writeEditableDocx(t *testing.T) string {
	t.Helper()
	doc := New()
	doc.AddHeading("一、整体概况", 1)
	doc.AddParagraph("概况正文")
	doc.AddHeading("1.1 销售趋势", 2)
	doc.AddParagraph("趋势正文")
	doc.AddHeading("1.1.1 存量章节", 3)
	doc.AddParagraph("存量正文")
	doc.AddHeading("1.2 渠道占比", 2)
	doc.AddParagraph("渠道正文")
	doc.AddHeading("二、结论", 1)
	doc.AddParagraph("结论正文")

	path := t.TempDir() + "/source.docx"
	if err := doc.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return path
}

func writeComplexEditableDocx(t *testing.T) string {
	t.Helper()
	doc := New()
	doc.AddHeading("一、总览", 1)
	doc.AddParagraph("一、总览正文")
	doc.AddHeading("2A 销售", 2)
	doc.AddParagraph("2A 销售正文")
	doc.AddHeading("3A-1 国内", 3)
	doc.AddParagraph("3A-1 国内正文")
	doc.AddHeading("4A-1 华东", 4)
	doc.AddParagraph("4A-1 华东正文")
	doc.AddHeading("5A-1 上海", 5)
	doc.AddParagraph("5A-1 上海正文")
	doc.AddHeading("6A-1 黄浦", 6)
	doc.AddParagraph("6A-1 黄浦正文")
	doc.AddHeading("6A-2 浦东", 6)
	doc.AddParagraph("6A-2 浦东正文")
	doc.AddHeading("5A-2 杭州", 5)
	doc.AddParagraph("5A-2 杭州正文")
	doc.AddHeading("6A-3 西湖", 6)
	doc.AddParagraph("6A-3 西湖正文")
	doc.AddHeading("4A-2 华南", 4)
	doc.AddParagraph("4A-2 华南正文")
	doc.AddHeading("5A-3 广州", 5)
	doc.AddParagraph("5A-3 广州正文")
	doc.AddHeading("3A-2 海外", 3)
	doc.AddParagraph("3A-2 海外正文")
	doc.AddHeading("4A-3 欧洲", 4)
	doc.AddParagraph("4A-3 欧洲正文")
	doc.AddHeading("5A-4 德国", 5)
	doc.AddParagraph("5A-4 德国正文")
	doc.AddHeading("2B 运营", 2)
	doc.AddParagraph("2B 运营正文")
	doc.AddHeading("3B-1 流量", 3)
	doc.AddParagraph("3B-1 流量正文")
	doc.AddHeading("4B-1 自然流量", 4)
	doc.AddParagraph("4B-1 自然流量正文")
	doc.AddHeading("5B-1 SEO", 5)
	doc.AddParagraph("5B-1 SEO正文")
	doc.AddHeading("6B-1 内容SEO", 6)
	doc.AddParagraph("6B-1 内容SEO正文")
	doc.AddHeading("3B-2 转化", 3)
	doc.AddParagraph("3B-2 转化正文")
	doc.AddHeading("2C 风险", 2)
	doc.AddParagraph("2C 风险正文")
	doc.AddHeading("二、结论", 1)
	doc.AddParagraph("二、结论正文")

	path := t.TempDir() + "/complex.docx"
	if err := doc.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return path
}

func openComplexDocAndFindHeading(t *testing.T, text string) (*EditableDocument, Heading) {
	t.Helper()
	doc, err := Open(writeComplexEditableDocx(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	headings, err := doc.Headings()
	if err != nil {
		t.Fatalf("Headings: %v", err)
	}
	return doc, findHeadingByText(t, headings, text)
}

func findHeadingByText(t *testing.T, headings []Heading, text string) Heading {
	t.Helper()
	for _, h := range headings {
		if h.Text == text {
			return h
		}
	}
	t.Fatalf("heading %q not found in %#v", text, headings)
	return Heading{}
}

func writeEditableDocumentAndUnzip(t *testing.T, doc *EditableDocument) map[string][]byte {
	t.Helper()
	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return unzipDocx(t, out.Bytes())
}

func reopenEditableDocumentFromBytes(t *testing.T, files map[string][]byte) *EditableDocument {
	t.Helper()
	var out bytes.Buffer
	path := t.TempDir() + "/reopen.docx"
	doc := &EditableDocument{
		files: files,
		order: make([]string, 0, len(files)),
	}
	for name := range files {
		doc.order = append(doc.order, name)
	}
	if err := doc.Write(&out); err != nil {
		t.Fatalf("Write reopen source: %v", err)
	}
	if err := os.WriteFile(path, out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	return reopened
}

func assertOrder(t *testing.T, s string, items ...string) {
	t.Helper()
	last := -1
	for _, item := range items {
		pos := strings.Index(s, item)
		if pos < 0 {
			t.Fatalf("missing %q in %s", item, s)
		}
		if pos <= last {
			t.Fatalf("%q appears out of order in %s", item, s)
		}
		last = pos
	}
}
