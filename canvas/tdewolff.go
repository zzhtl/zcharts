package canvas

import (
	"fmt"
	"image/png"
	"io"
	"math"
	"sync"

	zcolor "github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/font"

	tdcanvas "github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/pdf"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"github.com/tdewolff/canvas/renderers/svg"
)

// New 创建一个像素尺寸为 w×h 的画布。fonts 可为 nil（此时使用一个新建的默认 Manager）。
//
// 内部约定 1 canvas 单位 = 1 px：tdewolff 的 mm 单位被复用为 px，输出 PNG 时使用 1 dot/mm
// 的分辨率，保证 1:1 像素映射。
func New(w, h float64, fonts *font.Manager) Canvas {
	if fonts == nil {
		fonts = font.New()
	}
	c := tdcanvas.New(w, h)
	ctx := tdcanvas.NewContext(c)
	// 屏幕坐标系：原点在左上，Y 向下
	ctx.SetCoordSystem(tdcanvas.CartesianIV)
	return &tdAdapter{
		w:        w,
		h:        h,
		canvas:   c,
		ctx:      ctx,
		fonts:    fonts,
		families: map[string]*tdcanvas.FontFamily{},
	}
}

type tdAdapter struct {
	w, h   float64
	canvas *tdcanvas.Canvas
	ctx    *tdcanvas.Context
	fonts  *font.Manager

	mu       sync.Mutex
	families map[string]*tdcanvas.FontFamily

	hasFill, hasStroke bool
}

func (c *tdAdapter) Size() (float64, float64) { return c.w, c.h }

// ---- 样式 ----

func (c *tdAdapter) SetFill(col zcolor.Color) {
	c.ctx.SetFillColor(col.NRGBA())
	c.hasFill = col.A != 0
}

func (c *tdAdapter) SetStroke(col zcolor.Color) {
	c.ctx.SetStrokeColor(col.NRGBA())
	c.hasStroke = col.A != 0
}

func (c *tdAdapter) SetStrokeWidth(w float64) { c.ctx.SetStrokeWidth(w) }

func (c *tdAdapter) SetDash(offset float64, pattern []float64) {
	c.ctx.SetDashes(offset, pattern...)
}

func (c *tdAdapter) SetLineCap(cap LineCap) {
	switch cap {
	case CapRound:
		c.ctx.SetStrokeCapper(tdcanvas.RoundCap)
	case CapSquare:
		c.ctx.SetStrokeCapper(tdcanvas.SquareCap)
	default:
		c.ctx.SetStrokeCapper(tdcanvas.ButtCap)
	}
}

func (c *tdAdapter) NoFill() {
	c.ctx.SetFill(tdcanvas.Paint{})
	c.hasFill = false
}

func (c *tdAdapter) NoStroke() {
	c.ctx.SetStroke(tdcanvas.Paint{})
	c.hasStroke = false
}

// ---- 状态栈与变换 ----

func (c *tdAdapter) Save()                       { c.ctx.Push() }
func (c *tdAdapter) Restore()                    { c.ctx.Pop() }
func (c *tdAdapter) Translate(x, y float64)      { c.ctx.Translate(x, y) }
func (c *tdAdapter) Rotate(deg float64)          { c.ctx.Rotate(-deg) } // CartesianIV 翻 Y → 旋转方向反转
func (c *tdAdapter) Scale(sx, sy float64)        { c.ctx.Scale(sx, sy) }
func (c *tdAdapter) Clip(x, y, w, h float64)     { /* tdewolff 当前未在 Context 暴露 Clip；保留无操作占位 */ }
func (c *tdAdapter) ClearClip()                  {}

// ---- 路径 ----

func (c *tdAdapter) BeginPath()                          {}
func (c *tdAdapter) MoveTo(x, y float64)                 { c.ctx.MoveTo(x, y) }
func (c *tdAdapter) LineTo(x, y float64)                 { c.ctx.LineTo(x, y) }
func (c *tdAdapter) QuadTo(cx, cy, x, y float64)         { c.ctx.QuadTo(cx, cy, x, y) }
func (c *tdAdapter) CubicTo(c1x, c1y, c2x, c2y, x, y float64) {
	c.ctx.CubeTo(c1x, c1y, c2x, c2y, x, y)
}
func (c *tdAdapter) ClosePath()                          { c.ctx.Close() }
func (c *tdAdapter) Fill()                               { c.ctx.Fill() }
func (c *tdAdapter) Stroke()                             { c.ctx.Stroke() }
func (c *tdAdapter) FillStroke()                         { c.ctx.FillStroke() }

// ---- 高阶图元 ----

func (c *tdAdapter) DrawRect(x, y, w, h float64) {
	c.ctx.MoveTo(x, y)
	c.ctx.LineTo(x+w, y)
	c.ctx.LineTo(x+w, y+h)
	c.ctx.LineTo(x, y+h)
	c.ctx.Close()
	c.flushPath()
}

func (c *tdAdapter) DrawCircle(cx, cy, r float64) {
	c.ctx.MoveTo(cx+r, cy)
	c.ctx.Arc(r, r, 0, 0, 360)
	c.ctx.Close()
	c.flushPath()
}

func (c *tdAdapter) DrawLine(x1, y1, x2, y2 float64) {
	c.ctx.MoveTo(x1, y1)
	c.ctx.LineTo(x2, y2)
	c.ctx.Stroke()
}

// DrawSector 绘制环形扇形（饼图/仪表盘用）。
// 角度单位：度，0° 指向正右方，顺时针为正方向。
func (c *tdAdapter) DrawSector(cx, cy, rIn, rOut, startDeg, endDeg float64) {
	if endDeg < startDeg {
		startDeg, endDeg = endDeg, startDeg
	}
	if rOut <= 0 {
		return
	}
	if rIn < 0 {
		rIn = 0
	}
	// 角度（屏幕坐标系，Y 向下，顺时针为正）→ 笛卡尔（Y 向上，逆时针为正）：
	// 我们在 CartesianIV 下：Context 的 Arc 接受 CCW 角度（按内部 CartesianI 解释），
	// 但反射后视觉上变 CW。为了直观给上层使用顺时针角度，这里直接传入即可，
	// tdewolff 会按当前坐标系翻转处理。
	x0 := cx + rOut*cosDeg(startDeg)
	y0 := cy + rOut*sinDeg(startDeg)
	c.ctx.MoveTo(x0, y0)
	c.ctx.Arc(rOut, rOut, 0, startDeg, endDeg)
	if rIn > 0 {
		xi := cx + rIn*cosDeg(endDeg)
		yi := cy + rIn*sinDeg(endDeg)
		c.ctx.LineTo(xi, yi)
		c.ctx.Arc(rIn, rIn, 0, endDeg, startDeg)
		c.ctx.Close()
	} else {
		c.ctx.LineTo(cx, cy)
		c.ctx.Close()
	}
	c.flushPath()
}

func (c *tdAdapter) flushPath() {
	switch {
	case c.hasFill && c.hasStroke:
		c.ctx.FillStroke()
	case c.hasFill:
		c.ctx.Fill()
	case c.hasStroke:
		c.ctx.Stroke()
	default:
		// 没有填充也没有描边，丢弃当前 path
		c.ctx.Fill()
	}
}

// ---- 文字 ----

func (c *tdAdapter) DrawText(x, y float64, s string, style TextStyle) {
	if s == "" {
		return
	}
	face, err := c.face(style)
	if err != nil {
		return
	}
	w, h, ascent := c.measure(s, style)

	// VAlign 偏移：tdewolff 的 DrawText 在 CartesianIV 下，y 表示文字框 top（顶部）。
	// 我们把 (x,y) 视为 baseline，按用户的 VAlign 调整。
	var dy float64
	switch style.VAlign {
	case AlignBaseline:
		dy = -ascent
	case AlignTop:
		dy = 0
	case AlignMiddle:
		dy = -h / 2
	case AlignBottom:
		dy = -h
	}

	// Anchor 偏移
	var dx float64
	switch style.Anchor {
	case AnchorMiddle:
		dx = -w / 2
	case AnchorEnd:
		dx = -w
	}

	if style.Rotation != 0 {
		c.ctx.Push()
		c.ctx.Translate(x, y)
		c.ctx.Rotate(-style.Rotation)
		c.ctx.DrawText(dx, dy, tdcanvas.NewTextLine(face, s, tdcanvas.Left))
		c.ctx.Pop()
		return
	}
	c.ctx.DrawText(x+dx, y+dy, tdcanvas.NewTextLine(face, s, tdcanvas.Left))
}

func (c *tdAdapter) MeasureText(s string, style TextStyle) (float64, float64, float64) {
	return c.measure(s, style)
}

func (c *tdAdapter) measure(s string, style TextStyle) (w, h, ascent float64) {
	face, err := c.face(style)
	if err != nil {
		return 0, 0, 0
	}
	line := tdcanvas.NewTextLine(face, s, tdcanvas.Left)
	b := line.Bounds()
	w = b.W()
	h = b.H()
	// FontFace 的 Metrics 提供精确 ascent。
	m := face.Metrics()
	ascent = m.Ascent
	if ascent == 0 {
		ascent = h * 0.8
	}
	return w, h, ascent
}

// face 取（或惰性创建）一个 tdewolff FontFace。
//
// tdewolff 的 Face 接受 pt 单位（内部转 mm 时 size_mm = size_pt * 25.4/72）。
// 我们的 canvas 约定 1 单位 = 1 px = 1 mm，希望 style.Size 直接等于像素高，
// 因此需把 px → pt 反向预乘 ptPerMm。
const ptPerMm = 72.0 / 25.4

func (c *tdAdapter) face(style TextStyle) (*tdcanvas.FontFace, error) {
	family := style.Family
	if family == "" {
		family = font.DefaultFamily
	}
	size := style.Size
	if size <= 0 {
		size = 12
	}
	fam, err := c.fontFamily(family)
	if err != nil {
		return nil, err
	}
	fontStyle := tdcanvas.FontRegular
	if style.Weight == "bold" {
		fontStyle = tdcanvas.FontBold
	}
	return fam.Face(size*ptPerMm, style.Color.NRGBA(), fontStyle), nil
}

func (c *tdAdapter) fontFamily(family string) (*tdcanvas.FontFamily, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if fam, ok := c.families[family]; ok {
		return fam, nil
	}
	entry := c.fonts.Lookup(family)
	if entry == nil {
		return nil, fmt.Errorf("zcharts/canvas: font family %q not found", family)
	}
	fam := tdcanvas.NewFontFamily(family)
	if err := fam.LoadFont(entry.Data, 0, tdcanvas.FontRegular); err != nil {
		return nil, fmt.Errorf("zcharts/canvas: load font %q: %w", family, err)
	}
	c.families[family] = fam
	return fam, nil
}

// ---- 输出 ----

func (c *tdAdapter) Write(w io.Writer, format Format) error {
	switch format {
	case FormatPNG:
		// resolution=DPMM(1) → 1 canvas 单位 = 1 像素
		img := rasterizer.Draw(c.canvas, tdcanvas.DPMM(1), tdcanvas.DefaultColorSpace)
		return png.Encode(w, img)
	case FormatSVG:
		r := svg.New(w, c.w, c.h, nil)
		c.canvas.RenderTo(r)
		return r.Close()
	case FormatPDF:
		r := pdf.New(w, c.w, c.h, nil)
		c.canvas.RenderTo(r)
		return r.Close()
	default:
		return fmt.Errorf("zcharts/canvas: unsupported format %q", format)
	}
}

// ---- 小工具 ----

func cosDeg(d float64) float64 { return math.Cos(d * math.Pi / 180) }
func sinDeg(d float64) float64 { return math.Sin(d * math.Pi / 180) }
