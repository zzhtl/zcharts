package docx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/zzhtl/zcharts/docx/ooxml"
)

// nextPartFilename 在 files 中挑下一个未占用的部件文件名，命名形如 image3.png、chart2.xml。
// 避免追加新部件时与原 docx 包内已有文件冲突。
func nextPartFilename(files map[string][]byte, dir, prefix, ext string) string {
	for i := 1; ; i++ {
		filename := fmt.Sprintf("%s%d.%s", prefix, i, ext)
		if _, ok := files[dir+"/"+filename]; !ok {
			return filename
		}
	}
}

// nextRelID 在 used 集合外挑下一个未占用的 Relationship Id，形如 rIdImage3。
// 调用方需要在拿到 id 后主动放入 used，否则会反复返回同一个。
func nextRelID(used map[string]bool, prefix string) string {
	for i := 1; ; i++ {
		id := fmt.Sprintf("%s%d", prefix, i)
		if !used[id] {
			return id
		}
	}
}

// documentRelIDs 收集 word/_rels/document.xml.rels 中已存在的所有 Relationship Id，
// 作为 nextRelID 的初始占用集合，避免新 id 与原有 id 冲突。
func documentRelIDs(data []byte) map[string]bool {
	ids := map[string]bool{}
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return ids
		}
		if err != nil {
			return ids
		}
		t, ok := tok.(xml.StartElement)
		if !ok || t.Name.Local != "Relationship" {
			continue
		}
		if id := attrValue(t.Attr, "Id"); id != "" {
			ids[id] = true
		}
	}
}

// addDocumentRelationship 向 word/_rels/document.xml.rels 追加一条 Relationship。
// rels 文件缺失时直接以最小骨架新建，避免破坏原文件已有内容。
func addDocumentRelationship(files map[string][]byte, rel ooxml.Relationship) {
	const relsPath = "word/_rels/document.xml.rels"
	relXML := relationshipXML(rel)
	data := files[relsPath]
	if len(data) == 0 {
		files[relsPath] = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` + relXML + `</Relationships>`)
		return
	}
	files[relsPath] = insertBeforeClosingTag(data, "Relationships", relXML)
}

// relationshipXML 序列化单条 <Relationship/> 元素。
func relationshipXML(rel ooxml.Relationship) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<Relationship Id="%s" Type="%s" Target="%s"`,
		xmlEscape(rel.ID), xmlEscape(rel.Type), xmlEscape(rel.Target))
	if rel.TargetMode != "" {
		fmt.Fprintf(&b, ` TargetMode="%s"`, xmlEscape(rel.TargetMode))
	}
	b.WriteString(`/>`)
	return b.String()
}

// ensureDefaultContentType 保证 [Content_Types].xml 中存在指定扩展名的 Default 声明。
// 已存在则不变；文件缺失时以最小骨架新建。空参数静默忽略。
func ensureDefaultContentType(files map[string][]byte, ext, contentType string) {
	if ext == "" || contentType == "" {
		return
	}
	const path = "[Content_Types].xml"
	data := files[path]
	if len(data) == 0 {
		files[path] = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` + defaultContentTypeXML(ext, contentType) + `</Types>`)
		return
	}
	if hasDefaultContentType(data, ext) {
		return
	}
	files[path] = insertBeforeClosingTag(data, "Types", defaultContentTypeXML(ext, contentType))
}

// ensureOverrideContentType 保证 [Content_Types].xml 中存在指定部件名的 Override 声明。
// 已存在则不变；文件缺失时以最小骨架新建。空参数静默忽略。
func ensureOverrideContentType(files map[string][]byte, partName, contentType string) {
	if partName == "" || contentType == "" {
		return
	}
	const path = "[Content_Types].xml"
	data := files[path]
	if len(data) == 0 {
		files[path] = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` + overrideContentTypeXML(partName, contentType) + `</Types>`)
		return
	}
	if hasOverrideContentType(data, partName) {
		return
	}
	files[path] = insertBeforeClosingTag(data, "Types", overrideContentTypeXML(partName, contentType))
}

// defaultContentTypeXML 序列化单条 <Default/> 元素。
func defaultContentTypeXML(ext, contentType string) string {
	return `<Default Extension="` + xmlEscape(ext) + `" ContentType="` + xmlEscape(contentType) + `"/>`
}

// overrideContentTypeXML 序列化单条 <Override/> 元素。
func overrideContentTypeXML(partName, contentType string) string {
	return `<Override PartName="` + xmlEscape(partName) + `" ContentType="` + xmlEscape(contentType) + `"/>`
}

// hasDefaultContentType 判定 [Content_Types].xml 中是否已经声明过该扩展名的 Default。
// 扩展名比较不区分大小写。
func hasDefaultContentType(data []byte, ext string) bool {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return false
		}
		if err != nil {
			return false
		}
		t, ok := tok.(xml.StartElement)
		if !ok || t.Name.Local != "Default" {
			continue
		}
		if strings.EqualFold(attrValue(t.Attr, "Extension"), ext) {
			return true
		}
	}
}

// hasOverrideContentType 判定 [Content_Types].xml 中是否已经声明过该部件名的 Override。
func hasOverrideContentType(data []byte, partName string) bool {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return false
		}
		if err != nil {
			return false
		}
		t, ok := tok.(xml.StartElement)
		if !ok || t.Name.Local != "Override" {
			continue
		}
		if attrValue(t.Attr, "PartName") == partName {
			return true
		}
	}
}

// insertBeforeClosingTag 在 data 中最后一个 </localName> 标签之前插入 insert 文本。
// 找不到对应结束标签时退化为在末尾补一个，保证 XML 结构闭合。
func insertBeforeClosingTag(data []byte, localName, insert string) []byte {
	needle := []byte("</" + localName + ">")
	if pos := bytes.LastIndex(data, needle); pos >= 0 {
		out := make([]byte, 0, len(data)+len(insert))
		out = append(out, data[:pos]...)
		out = append(out, insert...)
		out = append(out, data[pos:]...)
		return out
	}
	return append(append(data, insert...), []byte("</"+localName+">")...)
}
