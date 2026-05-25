package chart

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"github.com/zzhtl/zcharts/common/errs"
	"github.com/zzhtl/zcharts/jsonopt"
)

func TestBuildChartXMLLineAndBar(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "title": {"text": "销售 <Q2>"},
	  "legend": {},
	  "xAxis": {"type": "category", "data": ["Mon","Tue"]},
	  "series": [
	    {"name": "订单", "type": "bar", "data": [12, 18]},
	    {"name": "增长", "type": "line", "smooth": true, "data": [3, 5]}
	  ]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := BuildChartXML(opt)
	if err != nil {
		t.Fatal(err)
	}
	assertValidXML(t, out)
	s := string(out)
	for _, want := range []string{
		`<c:barChart>`,
		`<c:lineChart>`,
		`销售 &lt;Q2&gt;`,
		`<c:v>Mon</c:v>`,
		`<a:srgbClr val="5470C6"/>`,
		`<a:srgbClr val="91CC75"/>`,
		`<c:smooth val="1"/>`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %s in %s", want, s)
		}
	}
}

func TestBuildChartXMLPie(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "series": [{
	    "name": "来源",
	    "type": "pie",
	    "data": [{"name":"搜索","value":20},{"name":"直接","value":10}]
	  }]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := BuildChartXML(opt)
	if err != nil {
		t.Fatal(err)
	}
	assertValidXML(t, out)
	s := string(out)
	if !strings.Contains(s, `<c:pieChart>`) ||
		!strings.Contains(s, `<c:v>搜索</c:v>`) ||
		!strings.Contains(s, `<c:dPt><c:idx val="0"/><c:spPr>`) {
		t.Fatalf("unexpected pie chart xml: %s", s)
	}
}

func TestBuildChartXMLDoughnut(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "series": [{
	    "name": "渠道",
	    "type": "pie",
	    "radius": ["40%", "70%"],
	    "data": [{"name":"搜索","value":20},{"name":"直接","value":10}]
	  }]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := BuildChartXML(opt)
	if err != nil {
		t.Fatal(err)
	}
	assertValidXML(t, out)
	s := string(out)
	if !strings.Contains(s, `<c:doughnutChart>`) ||
		!strings.Contains(s, `<c:holeSize val="57"/>`) {
		t.Fatalf("unexpected doughnut chart xml: %s", s)
	}
}

func TestBuildChartXMLScatter(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "title": {"text": "转化关系"},
	  "xAxis": {"type": "value"},
	  "yAxis": {"type": "value"},
	  "series": [
	    {"name": "样本 A", "type": "scatter", "symbolSize": 8, "data": [[10, 42], [18, 68]]},
	    {"name": "样本 B", "type": "scatter", "data": [[12, 36], [28, 90]]}
	  ]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := BuildChartXML(opt)
	if err != nil {
		t.Fatal(err)
	}
	assertValidXML(t, out)
	s := string(out)
	for _, want := range []string{
		`<c:scatterChart>`,
		`<c:scatterStyle val="marker"/>`,
		`<c:xVal><c:numLit>`,
		`<c:yVal><c:numLit>`,
		`<c:v>10</c:v>`,
		`<c:v>68</c:v>`,
		`<c:axPos val="b"/>`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %s in %s", want, s)
		}
	}
}

func TestBuildChartXMLRadar(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "title": {"text": "能力雷达"},
	  "radar": {
	    "indicator": [
	      {"name": "质量"},
	      {"name": "效率"},
	      {"name": "稳定"}
	    ]
	  },
	  "series": [{
	    "name": "团队",
	    "type": "radar",
	    "data": [
	      {"name": "当前", "value": [80, 65, 90]},
	      {"name": "目标", "value": [90, 85, 95]}
	    ]
	  }]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := BuildChartXML(opt)
	if err != nil {
		t.Fatal(err)
	}
	assertValidXML(t, out)
	s := string(out)
	for _, want := range []string{
		`<c:radarChart>`,
		`<c:radarStyle val="marker"/>`,
		`<c:v>质量</c:v>`,
		`<c:v>当前</c:v>`,
		`<c:v>95</c:v>`,
		`<c:catAx>`,
		`<c:valAx>`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %s in %s", want, s)
		}
	}
}

func TestBuildChartXMLUnsupportedSeries(t *testing.T) {
	opt, err := jsonopt.ParseString(`{"series":[{"type":"gauge","data":[{"value":66,"name":"完成率"}]}]}`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = BuildChartXML(opt)
	if !errors.Is(err, errs.ErrUnsupportedSeries) {
		t.Fatalf("err=%v", err)
	}
}

func TestBuildShapeXMLGauge(t *testing.T) {
	opt, err := jsonopt.ParseString(`{"series":[{"type":"gauge","data":[{"value":66,"name":"完成率"}]}]}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := BuildShapeXML(opt, 320, 180)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<v:group`,
		`<v:line`,
		`<v:oval`,
		`完成率`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
}

func TestBuildShapeXMLRejectsMixedSeries(t *testing.T) {
	opt, err := jsonopt.ParseString(`{
	  "series": [
	    {"type":"gauge","data":[{"value":66}]},
	    {"type":"gauge","data":[{"value":88}]}
	  ]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = BuildShapeXML(opt, 320, 180)
	if !errors.Is(err, errs.ErrUnsupportedSeries) {
		t.Fatalf("err=%v", err)
	}
}

func assertValidXML(t *testing.T, data []byte) {
	t.Helper()
	var v any
	if err := xml.NewDecoder(bytes.NewReader(data)).Decode(&v); err != nil {
		t.Fatalf("invalid xml: %v\n%s", err, data)
	}
}
