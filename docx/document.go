// Package docx 提供把 zcharts 图表插入 Word docx 文件的能力。
//
// 高层用法：
//
//	doc := docx.New()
//	doc.AddParagraph("月度报告")
//	doc.AddChart(opt, docx.AsImage(docx.PNG, 600, 400))
//	doc.Save("report.docx")
//
// 当前阶段只实现"插入图片"路径。AddChart 支持 docx.AsNativeChart() 占位，
// 用户调用时返回 errs.ErrNotImplemented。
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
	body   []bodyElement
	images []embedImage
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

// AddParagraph 在文档末尾追加一段普通文字段落。
func (d *Document) AddParagraph(text string) {
	d.body = append(d.body, paragraph{text: text})
}

// ChartInsertOption 控制 AddChart 的插入方式。
type ChartInsertOption struct {
	Format ImageFormat
	Width  int // 像素，用于渲染图表 + 决定 docx 中显示宽度
	Height int
	// 当 native=true 时按"原生 chart XML"路径处理（当前阶段不支持）。
	native bool
}

// AsImage 让 AddChart 以图片形式插入（默认路径）。
func AsImage(format ImageFormat, width, height int) ChartInsertOption {
	return ChartInsertOption{Format: format, Width: width, Height: height}
}

// AsNativeChart 选用 OOXML chart XML 路径（占位，未实现）。
func AsNativeChart() ChartInsertOption {
	return ChartInsertOption{native: true}
}

// AddChart 在文档末尾插入一个图表。
// 当 ChartInsertOption.native=true 时返回 errs.ErrNotImplemented。
func (d *Document) AddChart(opt *option.Option, ins ChartInsertOption, renderOpts ...chart.RenderOption) error {
	if ins.native {
		return errs.ErrNotImplemented
	}
	if ins.Width <= 0 {
		ins.Width = 600
	}
	if ins.Height <= 0 {
		ins.Height = 400
	}
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
		docPrID:   idx,
	})
	return nil
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
	buf.WriteString(`xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">`)
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
