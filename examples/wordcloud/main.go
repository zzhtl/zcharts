package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "产品反馈词云", "subtext": "用户评论关键词"},
  "series": [{
    "name": "关键词",
    "type": "wordCloud",
    "left": "center",
    "top": "middle",
    "width": "86%",
    "height": "78%",
    "sizeRange": [16, 58],
    "rotationRange": [0, 0],
    "gridSize": 9,
    "data": [
      {"name": "稳定", "value": 96},
      {"name": "易用", "value": 88},
      {"name": "多维报表", "value": 72},
      {"name": "实时告警", "value": 66},
      {"name": "性能稳定", "value": 62},
      {"name": "权限管理", "value": 58},
      {"name": "数据导出", "value": 52},
      {"name": "团队协作", "value": 48},
      {"name": "移动端", "value": 42},
      {"name": "可视化", "value": 38},
      {"name": "系统集成", "value": 34},
      {"name": "安全审计", "value": 30},
      {"name": "使用体验", "value": 28},
      {"name": "仪表盘", "value": 24},
      {"name": "自动化", "value": 22},
      {"name": "低延迟", "value": 20},
      {"name": "可配置", "value": 18},
      {"name": "稳定运行", "value": 16}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/wordcloud.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(760, 540)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/wordcloud.png")
}
