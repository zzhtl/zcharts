package docx_test

import (
	"bytes"
	"log"

	"github.com/zzhtl/zcharts/docx"
	"github.com/zzhtl/zcharts/jsonopt"
)

func ExampleDocument_AddChart() {
	const optionJSON = `{
	  "title": {"text": "每周销量"},
	  "xAxis": {"type": "category", "data": ["周一","周二","周三"]},
	  "yAxis": {"type": "value"},
	  "series": [{"name": "销量", "type": "bar", "data": [120, 200, 150]}]
	}`

	opt, err := jsonopt.ParseString(optionJSON)
	if err != nil {
		log.Fatal(err)
	}
	doc := docx.New()
	doc.AddParagraph("销售报告")
	if err := doc.AddChart(opt, docx.AsImage(docx.PNG, 640, 360)); err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	if err := doc.Write(&out); err != nil {
		log.Fatal(err)
	}
	// Output:
}
