// Package docx 提供把 zcharts 图表插入 Word docx 文件的能力。
//
// 高层用法：
//
//	doc := docx.New()
//	doc.AddParagraph("月度报告")
//	doc.AddChart(opt, docx.AsImage(docx.PNG, 600, 400))
//	doc.Save("report.docx")
//
// 当前实现支持两条路径：
//   - docx.AsImage(docx.PNG, width, height)：把 zcharts 渲染结果作为图片嵌入，支持所有已实现图表类型。
//   - docx.AsNativeChart(width, height)：优先生成 Word 原生 chart XML；部分非标准 chart 类型用 Word 形状绘制，其他图表自动退回 PNG 嵌入。
package docx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/chart"
	"github.com/zzhtl/zcharts/common/errs"
	nativechart "github.com/zzhtl/zcharts/docx/chart"
	"github.com/zzhtl/zcharts/docx/ooxml"
	"github.com/zzhtl/zcharts/option"
)

// ImageFormat 是嵌入到 docx 中的图片格式。
type ImageFormat string

const (
	PNG ImageFormat = "png"
	SVG ImageFormat = "svg"
)

// Document 是一个待生成的 docx 文档。
type Document struct {
	body      []bodyElement
	images    []embedImage
	charts    []embedChart
	drawingID int
}

type bodyElement interface {
	xml(*Document) string
}

type embedImage struct {
	rid      string
	filename string
	mime     string
	ext      string
	data     []byte
}

type embedChart struct {
	rid      string
	filename string
	data     []byte
}

// New 创建一个空文档。
func New() *Document { return &Document{} }

// ---- body elements ----

type paragraph struct{ text string }

func (p paragraph) xml(_ *Document) string {
	return `<w:p><w:r><w:t xml:space="preserve">` + xmlEscape(p.text) + `</w:t></w:r></w:p>`
}

type chartImage struct {
	rid       string
	widthEMU  int64
	heightEMU int64
	docPrID   int
}

func (c chartImage) xml(d *Document) string {
	return drawingXML(c.rid, c.docPrID, c.widthEMU, c.heightEMU)
}

type chartNative struct {
	rid       string
	widthEMU  int64
	heightEMU int64
	docPrID   int
}

func (c chartNative) xml(d *Document) string {
	return chartDrawingXML(c.rid, c.docPrID, c.widthEMU, c.heightEMU)
}

type chartShape struct {
	body string
}

func (c chartShape) xml(d *Document) string {
	return c.body
}

// AddParagraph 在文档末尾追加一段普通文字段落。
func (d *Document) AddParagraph(text string) {
	d.body = append(d.body, paragraph{text: text})
}

// ChartInsertOption 控制 AddChart 的插入方式。
type ChartInsertOption struct {
	Format ImageFormat
	Width  int // 像素，用于渲染图表 + 决定 docx 中显示宽度
	Height int
	// 当 native=true 时按"原生 chart XML"路径处理。
	native bool
}

// AsImage 让 AddChart 以图片形式插入（默认路径）。
func AsImage(format ImageFormat, width, height int) ChartInsertOption {
	return ChartInsertOption{Format: format, Width: width, Height: height}
}

// AsNativeChart 选用 OOXML 原生 chart XML 路径。可选传入 width、height 像素尺寸。
// 原生 chart XML 暂不支持但可用 Word 形状表达的 series 会走形状绘制，其余会自动退回 PNG 图片嵌入。
func AsNativeChart(size ...int) ChartInsertOption {
	ins := ChartInsertOption{native: true}
	if len(size) > 0 {
		ins.Width = size[0]
	}
	if len(size) > 1 {
		ins.Height = size[1]
	}
	return ins
}

// AddChart 在文档末尾插入一个图表。
// 默认以 PNG 图片嵌入；ChartInsertOption.native=true 时优先生成 Word 原生 chart XML。
func (d *Document) AddChart(opt *option.Option, ins ChartInsertOption, renderOpts ...chart.RenderOption) error {
	if ins.Width <= 0 {
		ins.Width = 600
	}
	if ins.Height <= 0 {
		ins.Height = 400
	}
	if ins.native {
		return d.addNativeChart(opt, ins, renderOpts...)
	}
	return d.addImageChart(opt, ins, renderOpts...)
}

func (d *Document) addImageChart(opt *option.Option, ins ChartInsertOption, renderOpts ...chart.RenderOption) error {
	format := canvas.FormatPNG
	mime, ext := "image/png", "png"
	switch ins.Format {
	case SVG:
		format, mime, ext = canvas.FormatSVG, "image/svg+xml", "svg"
	case "", PNG:
		// 默认 png
	default:
		return fmt.Errorf("docx: unsupported image format %q", ins.Format)
	}
	if format == canvas.FormatSVG {
		// Word 对 SVG 的支持依赖版本（Office 2019+ 支持 a:svgBlip 扩展）。
		// 第一阶段为最大兼容性，仍走 PNG 路径，但保留 SVG 选项以后扩展。
		format = canvas.FormatPNG
		mime, ext = "image/png", "png"
	}

	// 在内存中渲染图片
	var buf bytes.Buffer
	allOpts := append([]chart.RenderOption{chart.WithSize(float64(ins.Width), float64(ins.Height))}, renderOpts...)
	if err := chart.Render(opt, format, &buf, allOpts...); err != nil {
		return fmt.Errorf("docx: render chart: %w", err)
	}

	// 注册图片
	idx := len(d.images) + 1
	rid := fmt.Sprintf("rIdImage%d", idx)
	d.images = append(d.images, embedImage{
		rid:      rid,
		filename: fmt.Sprintf("image%d.%s", idx, ext),
		mime:     mime,
		ext:      ext,
		data:     buf.Bytes(),
	})

	// 转 EMU：1 px = 9525 EMU（96 DPI 标准）
	const emuPerPx = 9525
	d.body = append(d.body, chartImage{
		rid:       rid,
		widthEMU:  int64(ins.Width) * emuPerPx,
		heightEMU: int64(ins.Height) * emuPerPx,
		docPrID:   d.nextDrawingID(),
	})
	return nil
}

func (d *Document) addNativeChart(opt *option.Option, ins ChartInsertOption, renderOpts ...chart.RenderOption) error {
	data, err := nativechart.BuildChartXML(opt)
	if err != nil {
		if errors.Is(err, errs.ErrUnsupportedSeries) || errors.Is(err, errs.ErrNotImplemented) {
			body, shapeErr := nativechart.BuildShapeXML(opt, ins.Width, ins.Height)
			if shapeErr == nil {
				d.body = append(d.body, chartShape{body: body})
				return nil
			}
			if !errors.Is(shapeErr, errs.ErrUnsupportedSeries) && !errors.Is(shapeErr, errs.ErrNotImplemented) {
				return fmt.Errorf("docx: build native shape: %w", shapeErr)
			}
			ins.native = false
			ins.Format = PNG
			return d.addImageChart(opt, ins, renderOpts...)
		}
		return fmt.Errorf("docx: build native chart: %w", err)
	}

	idx := len(d.charts) + 1
	rid := fmt.Sprintf("rIdChart%d", idx)
	d.charts = append(d.charts, embedChart{
		rid:      rid,
		filename: fmt.Sprintf("chart%d.xml", idx),
		data:     data,
	})

	const emuPerPx = 9525
	d.body = append(d.body, chartNative{
		rid:       rid,
		widthEMU:  int64(ins.Width) * emuPerPx,
		heightEMU: int64(ins.Height) * emuPerPx,
		docPrID:   d.nextDrawingID(),
	})
	return nil
}

func (d *Document) nextDrawingID() int {
	d.drawingID++
	return d.drawingID
}

// Write 把文档写到 io.Writer。
func (d *Document) Write(w io.Writer) error {
	if w == nil {
		return errors.New("docx: nil writer")
	}
	pkg := ooxml.New()

	// 根 rels：指向 word/document.xml
	rootRels := ooxml.NewRels()
	rootRels.Add(ooxml.Relationship{
		ID:     "rId1",
		Type:   ooxml.RelOfficeDocument,
		Target: "word/document.xml",
	})
	pkg.AddPart(ooxml.Part{
		Name: "_rels/.rels",
		Data: rootRels.XML(),
	})

	// 文档 rels：图片关联
	docRels := ooxml.NewRels()
	imageExtensions := map[string]string{}
	for _, img := range d.images {
		docRels.Add(ooxml.Relationship{
			ID:     img.rid,
			Type:   ooxml.RelImage,
			Target: "media/" + img.filename,
		})
		imageExtensions[img.ext] = img.mime
	}
	for _, ch := range d.charts {
		docRels.Add(ooxml.Relationship{
			ID:     ch.rid,
			Type:   ooxml.RelChart,
			Target: "charts/" + ch.filename,
		})
	}
	pkg.AddPart(ooxml.Part{
		Name: "word/_rels/document.xml.rels",
		Data: docRels.XML(),
	})

	// 各图片 part
	for _, img := range d.images {
		pkg.AddPart(ooxml.Part{
			Name: "word/media/" + img.filename,
			Data: img.data,
		})
	}
	for _, ch := range d.charts {
		pkg.AddPart(ooxml.Part{
			Name:        "word/charts/" + ch.filename,
			ContentType: "application/vnd.openxmlformats-officedocument.drawingml.chart+xml",
			Data:        ch.data,
		})
	}
	for ext, mime := range imageExtensions {
		pkg.AddDefault(ext, mime)
	}

	// document.xml
	pkg.AddPart(ooxml.Part{
		Name:        "word/document.xml",
		ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml",
		Data:        d.documentXML(),
	})

	return pkg.Write(w)
}

// Save 把文档写出到本地文件。
func (d *Document) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.Write(f)
}

func (d *Document) documentXML() []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<w:document `)
	buf.WriteString(`xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" `)
	buf.WriteString(`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" `)
	buf.WriteString(`xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" `)
	buf.WriteString(`xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" `)
	buf.WriteString(`xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture" `)
	buf.WriteString(`xmlns:v="urn:schemas-microsoft-com:vml" `)
	buf.WriteString(`xmlns:o="urn:schemas-microsoft-com:office:office">`)
	buf.WriteString(`<w:body>`)
	for _, el := range d.body {
		buf.WriteString(el.xml(d))
	}
	// 必需的 sectPr 节点（Word 打开时需要）
	buf.WriteString(`<w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720" w:gutter="0"/></w:sectPr>`)
	buf.WriteString(`</w:body></w:document>`)
	return buf.Bytes()
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '&':
			b.WriteString("&amp;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
