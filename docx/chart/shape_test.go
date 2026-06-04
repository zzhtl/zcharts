package chart

import (
	"strings"
	"testing"

	"github.com/zzhtl/zcharts/jsonopt"
)

// 标题文本框需留足行高余量并清零段落间距，否则 14pt CJK 标题在 Word/WPS 中底部被裁。
func TestBuildShapeXMLTitleNotClipped(t *testing.T) {
	cases := map[string]string{
		"funnel":    `{"title":{"text":"漏斗图"},"series":[{"type":"funnel","data":[{"name":"访问","value":100},{"name":"成交","value":18}]}]}`,
		"gauge":     `{"title":{"text":"仪表盘"},"series":[{"type":"gauge","min":0,"max":100,"data":[{"value":76}]}]}`,
		"heatmap":   `{"title":{"text":"热力图"},"xAxis":{"type":"category","data":["Mon","Tue"]},"yAxis":{"type":"category","data":["早","晚"]},"series":[{"type":"heatmap","data":[[0,0,3],[1,1,9]]}]}`,
		"timeline":  `{"title":{"text":"时间轴"},"series":[{"type":"timeline","data":[{"time":"05-01","title":"需求","content":"完成"}]}]}`,
		"wordCloud": `{"title":{"text":"词云"},"series":[{"type":"wordCloud","sizeRange":[14,34],"data":[{"name":"Word","value":80}]}]}`,
	}
	for kind, optJSON := range cases {
		t.Run(kind, func(t *testing.T) {
			opt, err := jsonopt.ParseString(optJSON)
			if err != nil {
				t.Fatal(err)
			}
			xml, err := BuildShapeXML(opt, 480, 300)
			if err != nil {
				t.Fatalf("build shape: %v", err)
			}
			// 标题框高度须 >= 34px（容纳 14pt CJK 单行）。
			if !strings.Contains(xml, `top:8px;width:440px;height:34px`) {
				t.Errorf("%s: 标题框未使用加高几何，可能被裁: %s", kind, xml)
			}
			// 每个文本框段落都应清零间距、锁单倍行距。
			if !strings.Contains(xml, vmlParaPr) {
				t.Errorf("%s: 文本框缺少段落间距清零 pPr", kind)
			}
		})
	}
}
