// Package chart 是 zcharts 的对外入口。
//
// 它把 option 配置（或 ECharts JSON）渲染到 io.Writer，支持 PNG/SVG/PDF 三种格式。
package chart

import (
	"fmt"
	"io"

	"github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/font"
	"github.com/zzhtl/zcharts/jsonopt"
	"github.com/zzhtl/zcharts/option"
	"github.com/zzhtl/zcharts/render"
	"github.com/zzhtl/zcharts/theme"
)

// Format 是输出图片格式。
type Format = canvas.Format

const (
	FormatPNG = canvas.FormatPNG
	FormatSVG = canvas.FormatSVG
	FormatPDF = canvas.FormatPDF
)

// RenderOptions 控制渲染过程的参数。
type RenderOptions struct {
	Width           float64 // 像素，默认 800
	Height          float64 // 像素，默认 500
	Theme           string  // 主题名，默认 "default"
	Fonts           *font.Manager
	SeriesRenderers map[option.SeriesKind]render.SeriesRenderer
}

// RenderOption 是函数式选项。
type RenderOption func(*RenderOptions)

// WithSize 指定输出图像的像素尺寸。
func WithSize(w, h float64) RenderOption {
	return func(o *RenderOptions) { o.Width, o.Height = w, h }
}

// WithTheme 指定主题名称（需在 theme 包中已注册）。
func WithTheme(name string) RenderOption {
	return func(o *RenderOptions) { o.Theme = name }
}

// WithFonts 指定字体管理器。
func WithFonts(m *font.Manager) RenderOption {
	return func(o *RenderOptions) { o.Fonts = m }
}

// WithSeriesRenderer 为指定 series.type 注册本次渲染使用的自定义渲染器。
func WithSeriesRenderer(seriesType string, renderer render.SeriesRenderer) RenderOption {
	return func(o *RenderOptions) {
		if o.SeriesRenderers == nil {
			o.SeriesRenderers = make(map[option.SeriesKind]render.SeriesRenderer)
		}
		o.SeriesRenderers[option.SeriesKind(seriesType)] = renderer
	}
}

func defaults(opts []RenderOption) RenderOptions {
	out := RenderOptions{Width: 800, Height: 500, Theme: "default"}
	for _, o := range opts {
		o(&out)
	}
	return out
}

// Render 把 opt 渲染到 w，按 format 决定输出格式。
func Render(opt *option.Option, format Format, w io.Writer, opts ...RenderOption) error {
	o := defaults(opts)
	th, err := theme.Get(o.Theme)
	if err != nil {
		return fmt.Errorf("chart: %w", err)
	}
	c := canvas.New(o.Width, o.Height, o.Fonts)
	if err := render.RenderWithOptions(c, opt, th, render.RenderOptions{
		SeriesRenderers: o.SeriesRenderers,
	}); err != nil {
		return err
	}
	return c.Write(w, format)
}

// RenderFromJSON 解析 ECharts JSON 后渲染。
func RenderFromJSON(data []byte, format Format, w io.Writer, opts ...RenderOption) error {
	opt, err := jsonopt.Parse(data)
	if err != nil {
		return err
	}
	return Render(opt, format, w, opts...)
}
