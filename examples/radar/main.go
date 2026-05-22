package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "球员能力画像"},
  "legend": {},
  "radar": {
    "indicator": [
      {"name": "进攻", "max": 100},
      {"name": "防守", "max": 100},
      {"name": "速度", "max": 100},
      {"name": "体能", "max": 100},
      {"name": "技术", "max": 100},
      {"name": "配合", "max": 100}
    ],
    "center": ["50%", "55%"],
    "radius": "60%"
  },
  "series": [{
    "type": "radar",
    "data": [
      {"name": "球员甲", "value": [85, 70, 90, 60, 80, 75]},
      {"name": "球员乙", "value": [60, 85, 70, 90, 65, 80]}
    ]
  }]
}`

func main() {
	if err := os.MkdirAll("assets/tmp", 0755); err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("assets/tmp/radar.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(700, 600)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/tmp/radar.png")
}
