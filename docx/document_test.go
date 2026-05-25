package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/zzhtl/zcharts/jsonopt"
)

const testLineJSON = `{
  "title": {"text": "Word 图表"},
  "xAxis": {"type": "category", "data": ["Mon","Tue"]},
  "yAxis": {"type": "value"},
  "series": [{"name": "sales", "type": "line", "data": [12, 18]}]
}`

func TestDocumentWriteChartPackage(t *testing.T) {
	opt, err := jsonopt.ParseString(testLineJSON)
	if err != nil {
		t.Fatal(err)
	}
	doc := New()
	doc.AddParagraph(`报告 <Q2> & "charts"`)
	if err := doc.AddChart(opt, AsImage(PNG, 320, 180)); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatal(err)
	}

	files := unzipDocx(t, out.Bytes())
	for _, name := range []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
		"word/_rels/document.xml.rels",
		"word/media/image1.png",
	} {
		if _, ok := files[name]; !ok {
			t.Fatalf("missing docx part %s", name)
		}
	}

	assertXML(t, files["[Content_Types].xml"])
	assertXML(t, files["_rels/.rels"])
	assertXML(t, files["word/document.xml"])
	assertXML(t, files["word/_rels/document.xml.rels"])

	contentTypes := string(files["[Content_Types].xml"])
	if !strings.Contains(contentTypes, `Extension="png" ContentType="image/png"`) {
		t.Fatalf("missing png content type: %s", contentTypes)
	}
	if !strings.Contains(contentTypes, `PartName="/word/document.xml"`) {
		t.Fatalf("missing document override: %s", contentTypes)
	}

	documentXML := string(files["word/document.xml"])
	if !strings.Contains(documentXML, `报告 &lt;Q2&gt; &amp; &quot;charts&quot;`) {
		t.Fatalf("paragraph text was not escaped: %s", documentXML)
	}
	if !strings.Contains(documentXML, `r:embed="rIdImage1"`) {
		t.Fatalf("missing image relationship reference: %s", documentXML)
	}
	if !strings.Contains(documentXML, `cx="3048000" cy="1714500"`) {
		t.Fatalf("unexpected image extent: %s", documentXML)
	}

	relsXML := string(files["word/_rels/document.xml.rels"])
	if !strings.Contains(relsXML, `Id="rIdImage1"`) ||
		!strings.Contains(relsXML, `Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"`) ||
		!strings.Contains(relsXML, `Target="media/image1.png"`) {
		t.Fatalf("unexpected document rels: %s", relsXML)
	}

	img, err := png.Decode(bytes.NewReader(files["word/media/image1.png"]))
	if err != nil {
		t.Fatal(err)
	}
	if got := img.Bounds().Dx(); got != 320 {
		t.Fatalf("image width=%d", got)
	}
	if got := img.Bounds().Dy(); got != 180 {
		t.Fatalf("image height=%d", got)
	}
}

func TestDocumentWriteRejectsNilWriter(t *testing.T) {
	if err := New().Write(nil); err == nil {
		t.Fatal("expected nil writer error")
	}
}

func TestDocumentWriteNativeChartPackage(t *testing.T) {
	opt, err := jsonopt.ParseString(testLineJSON)
	if err != nil {
		t.Fatal(err)
	}
	doc := New()
	doc.AddParagraph("原生 Word 图表")
	if err := doc.AddChart(opt, AsNativeChart()); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatal(err)
	}

	files := unzipDocx(t, out.Bytes())
	for _, name := range []string{
		"word/document.xml",
		"word/_rels/document.xml.rels",
		"word/charts/chart1.xml",
	} {
		if _, ok := files[name]; !ok {
			t.Fatalf("missing docx part %s", name)
		}
	}
	assertXML(t, files["word/charts/chart1.xml"])

	documentXML := string(files["word/document.xml"])
	if !strings.Contains(documentXML, `r:id="rIdChart1"`) {
		t.Fatalf("missing chart relationship reference: %s", documentXML)
	}
	relsXML := string(files["word/_rels/document.xml.rels"])
	if !strings.Contains(relsXML, `Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart"`) ||
		!strings.Contains(relsXML, `Target="charts/chart1.xml"`) {
		t.Fatalf("unexpected document rels: %s", relsXML)
	}
	contentTypes := string(files["[Content_Types].xml"])
	if !strings.Contains(contentTypes, `PartName="/word/charts/chart1.xml" ContentType="application/vnd.openxmlformats-officedocument.drawingml.chart+xml"`) {
		t.Fatalf("missing chart content type: %s", contentTypes)
	}
	chartXML := string(files["word/charts/chart1.xml"])
	if !strings.Contains(chartXML, `<c:lineChart>`) ||
		!strings.Contains(chartXML, `<c:v>Mon</c:v>`) ||
		!strings.Contains(chartXML, `<c:v>18</c:v>`) {
		t.Fatalf("unexpected chart xml: %s", chartXML)
	}
}

func TestDocumentWriteNativeScatterChartPackage(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "xAxis": {"type": "value"},
	  "yAxis": {"type": "value"},
	  "series": [{"name": "样本", "type": "scatter", "data": [[1,2], [3,4]]}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	doc := New()
	if err := doc.AddChart(opt, AsNativeChart(320, 180)); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatal(err)
	}

	files := unzipDocx(t, out.Bytes())
	if _, ok := files["word/media/image1.png"]; ok {
		t.Fatal("native scatter chart should not fall back to PNG image")
	}
	chartXML := string(files["word/charts/chart1.xml"])
	if !strings.Contains(chartXML, `<c:scatterChart>`) ||
		!strings.Contains(chartXML, `<c:xVal><c:numLit>`) ||
		!strings.Contains(chartXML, `<c:yVal><c:numLit>`) {
		t.Fatalf("unexpected scatter chart xml: %s", chartXML)
	}
}

func TestDocumentAddNativeChartUsesShapeDrawingForGauge(t *testing.T) {
	opt, err := jsonopt.ParseString(`{"series":[{"type":"gauge","data":[{"value":66,"name":"完成率"}]}]}`)
	if err != nil {
		t.Fatal(err)
	}
	doc := New()
	if err := doc.AddChart(opt, AsNativeChart(320, 180)); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatal(err)
	}

	files := unzipDocx(t, out.Bytes())
	assertXML(t, files["word/document.xml"])
	if _, ok := files["word/charts/chart1.xml"]; ok {
		t.Fatal("gauge shape drawing should not create chart XML")
	}
	if _, ok := files["word/media/image1.png"]; ok {
		t.Fatal("gauge shape drawing should not fall back to PNG image")
	}
	documentXML := string(files["word/document.xml"])
	if !strings.Contains(documentXML, `<v:group`) ||
		!strings.Contains(documentXML, `<v:line`) ||
		!strings.Contains(documentXML, `完成率`) {
		t.Fatalf("missing native shape drawing: %s", documentXML)
	}
}

func TestDocumentAddNativeChartFallsBackToImageForUnsupportedSeries(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "series": [
	    {"name":"A","type":"pie","data":[{"name":"x","value":1}]},
	    {"name":"B","type":"pie","data":[{"name":"y","value":2}]}
	  ]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	doc := New()
	if err := doc.AddChart(opt, AsNativeChart(320, 180)); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		t.Fatal(err)
	}

	files := unzipDocx(t, out.Bytes())
	if _, ok := files["word/charts/chart1.xml"]; ok {
		t.Fatal("unsupported native chart should not create chart XML")
	}
	if _, ok := files["word/media/image1.png"]; !ok {
		t.Fatal("unsupported native chart should fall back to PNG image")
	}
	documentXML := string(files["word/document.xml"])
	if !strings.Contains(documentXML, `r:embed="rIdImage1"`) {
		t.Fatalf("missing image fallback relationship reference: %s", documentXML)
	}
}

func unzipDocx(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, readErr := io.ReadAll(rc)
		closeErr := rc.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		files[f.Name] = body
	}
	return files
}

func assertXML(t *testing.T, data []byte) {
	t.Helper()
	var v any
	if err := xml.Unmarshal(data, &v); err != nil {
		t.Fatalf("invalid xml: %v\n%s", err, data)
	}
}
