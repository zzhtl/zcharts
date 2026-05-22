package main

import (
	"log"
	"os"

	"github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Player Skill Profile"},
  "radar": {
    "indicator": [
      {"name": "Attack",   "max": 100},
      {"name": "Defense",  "max": 100},
      {"name": "Speed",    "max": 100},
      {"name": "Stamina",  "max": 100},
      {"name": "Skill",    "max": 100},
      {"name": "Teamwork", "max": 100}
    ],
    "center": ["50%", "55%"],
    "radius": "60%"
  },
  "series": [{
    "type": "radar",
    "data": [
      {"name": "Player A", "value": [85, 70, 90, 60, 80, 75]},
      {"name": "Player B", "value": [60, 85, 70, 90, 65, 80]}
    ]
  }]
}`

func main() {
	out, err := os.Create("radar.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, out,
		chart.WithSize(700, 600)); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote radar.png")
}
