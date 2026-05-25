package jsonopt

import (
	"testing"

	"github.com/zzhtl/zcharts/option"
)

func TestParseLineBasic(t *testing.T) {
	src := `{
		"title": {"text": "Sales"},
		"xAxis": {"type": "category", "data": ["Mon","Tue","Wed"]},
		"yAxis": {"type": "value"},
		"series": [{"type": "line", "name": "A", "data": [120, 200, 150]}]
	}`
	opt, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if opt.Title == nil || opt.Title.Text != "Sales" {
		t.Fatalf("title=%v", opt.Title)
	}
	if len(opt.XAxis) != 1 || opt.XAxis[0].Type != option.AxisCategory {
		t.Fatalf("xAxis=%v", opt.XAxis)
	}
	if len(opt.XAxis[0].Data) != 3 || opt.XAxis[0].Data[0] != "Mon" {
		t.Fatalf("xAxis.data=%v", opt.XAxis[0].Data)
	}
	if len(opt.Series) != 1 {
		t.Fatalf("series len=%d", len(opt.Series))
	}
	line, ok := opt.Series[0].(*option.LineSeries)
	if !ok {
		t.Fatalf("series[0] type=%T", opt.Series[0])
	}
	if got := line.Data.AsValues(); got[0] != 120 || got[2] != 150 {
		t.Fatalf("data=%v", got)
	}
}

func TestParsePieRadius(t *testing.T) {
	src := `{
		"series": [{"type": "pie", "radius": ["40%", "70%"], "center": ["50%","50%"],
			"data":[{"name":"A","value":40},{"name":"B","value":60}]}]
	}`
	opt, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	p := opt.Series[0].(*option.PieSeries)
	if len(p.Radius) != 2 || !p.Radius[0].IsPercent() || p.Radius[0].Value != 40 {
		t.Fatalf("radius=%v", p.Radius)
	}
	if p.Data[0].Name != "A" || p.Data[0].Number() != 40 {
		t.Fatalf("data[0]=%+v", p.Data[0])
	}
}

func TestParseBarMixed(t *testing.T) {
	src := `{
		"xAxis":{"type":"category","data":["a","b"]},
		"yAxis":{"type":"value"},
		"series":[{"type":"bar","data":[10,20]},{"type":"line","data":[5,15]}]
	}`
	opt, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := opt.Series[0].(*option.BarSeries); !ok {
		t.Fatalf("series[0]=%T", opt.Series[0])
	}
	if _, ok := opt.Series[1].(*option.LineSeries); !ok {
		t.Fatalf("series[1]=%T", opt.Series[1])
	}
}

func TestParseAdditionalSeriesTypes(t *testing.T) {
	src := `{
		"series":[
			{"type":"funnel","data":[{"name":"A","value":10}]},
			{"type":"wordCloud","sizeRange":[12,48],"data":[{"name":"稳定","value":30}]},
			{"type":"timeline","data":[{"time":"2026-05-01","title":"启动","content":"项目启动"}]},
			{"type":"metricCard","coordinateSystem":"none","data":[{"name":"KPI","value":1}]}
		]
	}`
	opt, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := opt.Series[0].(*option.FunnelSeries); !ok {
		t.Fatalf("series[0]=%T", opt.Series[0])
	}
	if wc, ok := opt.Series[1].(*option.WordCloudSeries); !ok || len(wc.SizeRange) != 2 {
		t.Fatalf("series[1]=%T", opt.Series[1])
	}
	timeline, ok := opt.Series[2].(*option.TimelineSeries)
	if !ok || len(timeline.Data) != 1 || timeline.Data[0].Title != "启动" {
		t.Fatalf("series[2]=%T data=%+v", opt.Series[2], timeline)
	}
	custom, ok := opt.Series[3].(*option.CustomSeries)
	if !ok {
		t.Fatalf("series[3]=%T", opt.Series[3])
	}
	if custom.Kind() != "metricCard" || len(custom.Raw) == 0 {
		t.Fatalf("custom=%+v raw=%s", custom, string(custom.Raw))
	}
}

func TestParseTimePair(t *testing.T) {
	src := `{
		"xAxis":{"type":"time"},
		"series":[{"type":"line","data":[["2026-05-01",120],["2026-05-02",140]]}]
	}`
	opt, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	line := opt.Series[0].(*option.LineSeries)
	if len(line.Data[0].Values) != 2 {
		t.Fatalf("values=%v", line.Data[0].Values)
	}
	if line.Data[0].Values[0] <= 0 || line.Data[0].Values[1] != 120 {
		t.Fatalf("values=%v", line.Data[0].Values)
	}
}

func TestParseScatterPair(t *testing.T) {
	src := `{"series":[{"type":"scatter","data":[[1,2],[3,4],[5,6]]}]}`
	opt, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	s := opt.Series[0].(*option.ScatterSeries)
	x, y := s.Data[0].Pair()
	if x != 1 || y != 2 {
		t.Fatalf("pair[0]=(%v,%v)", x, y)
	}
}

func TestParseFlex(t *testing.T) {
	src := `{"grid":{"left":30,"right":"10%","top":"center"}}`
	opt, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if opt.Grid.Left.Value != 30 || opt.Grid.Left.Unit != "" {
		t.Fatalf("left=%+v", opt.Grid.Left)
	}
	if opt.Grid.Right.Value != 10 || opt.Grid.Right.Unit != "%" {
		t.Fatalf("right=%+v", opt.Grid.Right)
	}
	if opt.Grid.Top.Keyword != "center" {
		t.Fatalf("top=%+v", opt.Grid.Top)
	}
}

func TestParseUnknownIgnored(t *testing.T) {
	src := `{"animation":false,"madeUpField":123,"series":[]}`
	_, err := ParseString(src)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}
