package ooxml

import (
	"bytes"
	"fmt"
)

// Relationship 描述 OOXML 的一个 part-level relationship。
type Relationship struct {
	ID     string // "rId1"
	Type   string // 关系类型完整 URL
	Target string // 目标路径，相对当前 .rels 所属 part
	// TargetMode 留空时为内部；"External" 表示外部资源
	TargetMode string
}

// Rels 是一组 Relationship 的集合，可序列化为 OOXML .rels XML。
type Rels struct {
	items   []Relationship
	counter int // 自动分配 ID 时用
}

// NewRels 创建空集合。
func NewRels() *Rels { return &Rels{} }

// Add 添加 relationship。当 r.ID 为空时自动生成 "rIdN"。
// 返回最终生效的 ID。
func (r *Rels) Add(rel Relationship) string {
	if rel.ID == "" {
		r.counter++
		rel.ID = fmt.Sprintf("rId%d", r.counter)
	}
	r.items = append(r.items, rel)
	return rel.ID
}

// XML 序列化为 .rels 文件字节。
func (r *Rels) XML() []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for _, rel := range r.items {
		fmt.Fprintf(&buf, `<Relationship Id="%s" Type="%s" Target="%s"`,
			rel.ID, rel.Type, rel.Target)
		if rel.TargetMode != "" {
			fmt.Fprintf(&buf, ` TargetMode="%s"`, rel.TargetMode)
		}
		buf.WriteString(`/>`)
	}
	buf.WriteString(`</Relationships>`)
	return buf.Bytes()
}

// 常用 relationship type URI。
const (
	RelOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	RelImage          = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	RelChart          = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart"
)
