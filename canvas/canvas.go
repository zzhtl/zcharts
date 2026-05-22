// Package canvas 是 zcharts 的绘制抽象层。
//
// 它向上层提供与具体图形库无关的绘图接口（路径、文字、变换、状态栈），
// 向下封装 github.com/tdewolff/canvas 完成实际的栅格化与多格式输出。
//
// 单位约定
//
//	所有坐标、宽高、字体 size 均以"像素"为内部单位（适配器内部直接当 mm 传给 tdewolff，
//	输出 PNG 时用 1 dot/mm 的分辨率，从而保证 px:像素 = 1:1）。
package canvas

import (
	"io"

	"github.com/zzhtl/zcharts/common/color"
)

// Format 是输出图片格式。
type Format string

const (
	FormatPNG Format = "png"
	FormatSVG Format = "svg"
	FormatPDF Format = "pdf"
)

// TextAnchor 决定 DrawText 的 x 坐标作为文字框的哪个水平位置。
type TextAnchor uint8

const (
	AnchorStart  TextAnchor = iota // 左对齐
	AnchorMiddle                    // 水平居中
	AnchorEnd                       // 右对齐
)

// VerticalAlign 决定 DrawText 的 y 坐标作为文字框的哪个垂直位置。
type VerticalAlign uint8

const (
	AlignBaseline VerticalAlign = iota // 基线对齐（默认）
	AlignTop                            // 文字顶部
	AlignMiddle                         // 文字中央
	AlignBottom                         // 文字底部
)

// TextStyle 单次绘制文字的样式。
type TextStyle struct {
	Family   string
	Size     float64
	Color    color.Color
	Weight   string // "normal" / "bold"
	Anchor   TextAnchor
	VAlign   VerticalAlign
	Rotation float64 // 顺时针角度，0 为水平
}

// LineCap 描述线端样式。
type LineCap uint8

const (
	CapButt LineCap = iota
	CapRound
	CapSquare
)

// Canvas 是上层渲染所依赖的绘图接口。
type Canvas interface {
	// Size 返回画布像素尺寸。
	Size() (w, h float64)

	// 样式
	SetFill(c color.Color)
	SetStroke(c color.Color)
	SetStrokeWidth(w float64)
	SetDash(offset float64, pattern []float64)
	SetLineCap(cap LineCap)
	NoFill()
	NoStroke()

	// 状态栈
	Save()
	Restore()
	Translate(x, y float64)
	Rotate(deg float64)
	Scale(sx, sy float64)
	Clip(x, y, w, h float64)
	ClearClip()

	// 路径
	BeginPath()
	MoveTo(x, y float64)
	LineTo(x, y float64)
	QuadTo(cx, cy, x, y float64)
	CubicTo(c1x, c1y, c2x, c2y, x, y float64)
	ClosePath()
	Fill()
	Stroke()
	FillStroke()

	// 高阶图元
	DrawRect(x, y, w, h float64)
	DrawCircle(cx, cy, r float64)
	DrawLine(x1, y1, x2, y2 float64)
	// DrawSector 绘制以 (cx,cy) 为中心、内径 rIn、外径 rOut，角度区间
	// [startDeg, endDeg]（顺时针，0° 指向正右方）的环形扇形。
	// rIn=0 时退化为饼图扇形。
	DrawSector(cx, cy, rIn, rOut, startDeg, endDeg float64)

	// 文字
	DrawText(x, y float64, text string, style TextStyle)
	// MeasureText 返回文字宽度、总高度以及基线之上的 ascent。
	MeasureText(text string, style TextStyle) (w, h, ascent float64)

	// 输出
	Write(w io.Writer, format Format) error
}
