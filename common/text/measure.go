// Package text 提供文本测量与简单换行能力。
//
// 这里不依赖具体的字体加载实现，而是接受 golang.org/x/image/font.Face 接口，
// 由 zcharts/font 包负责创建 Face。
package text

import (
	"github.com/zzhtl/zcharts/common/geom"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Metrics 描述一段文本在指定 Face 下的渲染尺寸。
type Metrics struct {
	Width   float64 // 文本宽度（像素）
	Ascent  float64 // 基线到顶部
	Descent float64 // 基线到底部（正值）
	Height  float64 // Ascent + Descent
}

// Measure 计算 s 在 face 下的尺寸。
func Measure(face font.Face, s string) Metrics {
	if face == nil {
		return Metrics{}
	}
	advance := font.MeasureString(face, s)
	m := face.Metrics()
	return Metrics{
		Width:   fixedToFloat(advance),
		Ascent:  fixedToFloat(m.Ascent),
		Descent: fixedToFloat(m.Descent),
		Height:  fixedToFloat(m.Ascent + m.Descent),
	}
}

// MeasureBounds 返回贴合的 bounding rect（以基线为 Y=0，X 起点为 0）。
func MeasureBounds(face font.Face, s string) geom.Rect {
	m := Measure(face, s)
	return geom.Rect{X: 0, Y: -m.Ascent, W: m.Width, H: m.Height}
}

func fixedToFloat(x fixed.Int26_6) float64 {
	return float64(x) / 64.0
}
