// Package ooxml 是 docx 的底层 OOXML 包结构操作：把多个"part"打包成 zip，
// 并维护 ContentTypes 与 Relationships。
//
// 这里只实现 zcharts/docx 所需的最小集合，不追求完整 OOXML 规范。
package ooxml

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"sort"
)

// Part 是 docx 包内的一个文件条目。
type Part struct {
	Name        string // 包内路径，如 "word/document.xml"
	ContentType string // MIME，仅 Override 时需要；对常见扩展可走 Default
	Data        []byte
}

// Package 是一个待写出的 docx 包。
type Package struct {
	parts             []Part
	defaultExtensions map[string]string // ext → content-type
	overrides         map[string]string // path → content-type
}

// New 创建一个带有常用 default content types 的空包。
func New() *Package {
	return &Package{
		defaultExtensions: map[string]string{
			"rels": "application/vnd.openxmlformats-package.relationships+xml",
			"xml":  "application/xml",
		},
		overrides: map[string]string{},
	}
}

// AddPart 加入一个 part；当 contentType 非空时同时注册 Override。
func (p *Package) AddPart(part Part) {
	p.parts = append(p.parts, part)
	if part.ContentType != "" {
		p.overrides["/"+part.Name] = part.ContentType
	}
}

// AddDefault 注册一个扩展名 → content-type 映射，避免每个 part 都设 Override。
func (p *Package) AddDefault(ext, contentType string) {
	p.defaultExtensions[ext] = contentType
}

// Write 把包写出为 docx (zip) 流。
func (p *Package) Write(w io.Writer) error {
	zw := zip.NewWriter(w)
	// 1. 写 [Content_Types].xml（OOXML 规范要求位于包根）
	ct, err := zw.Create("[Content_Types].xml")
	if err != nil {
		return err
	}
	if _, err := ct.Write(p.contentTypesXML()); err != nil {
		return err
	}
	// 2. 写其余 parts
	for _, part := range p.parts {
		f, err := zw.Create(part.Name)
		if err != nil {
			return err
		}
		if _, err := f.Write(part.Data); err != nil {
			return err
		}
	}
	return zw.Close()
}

// Bytes 返回 docx 的字节切片。
func (p *Package) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Package) contentTypesXML() []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	defaults := make([]string, 0, len(p.defaultExtensions))
	for ext := range p.defaultExtensions {
		defaults = append(defaults, ext)
	}
	sort.Strings(defaults)
	for _, ext := range defaults {
		ct := p.defaultExtensions[ext]
		fmt.Fprintf(&buf, `<Default Extension="%s" ContentType="%s"/>`, ext, ct)
	}
	overrides := make([]string, 0, len(p.overrides))
	for path := range p.overrides {
		overrides = append(overrides, path)
	}
	sort.Strings(overrides)
	for _, path := range overrides {
		ct := p.overrides[path]
		fmt.Fprintf(&buf, `<Override PartName="%s" ContentType="%s"/>`, path, ct)
	}
	buf.WriteString(`</Types>`)
	return buf.Bytes()
}
