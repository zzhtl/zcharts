package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "项目事件时间轴", "subtext": "vertical timeline"},
  "legend": {"show": false},
  "series": [{
    "name": "项目事件",
    "type": "timeline",
    "data": [
      {"time": "2026-05-01", "title": "需求确认", "content": "完成图表导出范围和样式规范确认"},
      {"time": "2026-05-06", "title": "原型评审", "content": "确定时间线节点、事件标题和说明的展示层级"},
      {"time": "2026-05-12", "title": "开发联调", "content": "接入事件数据并验证 PNG / SVG / PDF 输出"},
      {"time": "2026-05-18", "title": "版本发布", "content": "发布静态图表生成能力并沉淀示例配置"}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/timeline.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(860, 500)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/timeline.png")
}
