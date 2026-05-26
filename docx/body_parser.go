package docx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
)

// parsedBody 是 word/document.xml 中 <w:body> 的解析结果。
// 记录每个顶层子元素的字节区间，以及 </w:body> 起点；用来在原始字节流上
// 精确定位插入位置，避免重新序列化整篇文档。
type parsedBody struct {
	children     []parsedBodyChild
	bodyEndStart int
}

// parsedBodyChild 是 body 下一个顶层子元素（<w:p> / <w:tbl> / <w:sectPr> 等）的位置信息。
// heading 非 nil 表示该 <w:p> 已被识别为标题。
type parsedBodyChild struct {
	start   int
	end     int
	name    xml.Name
	heading *Heading
}

// parseBodyXML 流式扫描 document.xml，只记录 body 下顶层子元素的字节起止偏移，
// 不缓存元素内容。标题段落顺带通过 parseHeadingParagraph 识别并编号。
func parseBodyXML(docXML []byte, styleLevels map[string]int) (parsedBody, error) {
	dec := xml.NewDecoder(bytes.NewReader(docXML))
	var body parsedBody
	inBody := false
	childDepth := 0
	var current *parsedBodyChild
	headingIndex := 0

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return parsedBody{}, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			tokenEnd := int(dec.InputOffset())
			if !inBody {
				if t.Name.Local == "body" {
					inBody = true
				}
				continue
			}
			start := tagStart(docXML, tokenEnd)
			if childDepth == 0 {
				current = &parsedBodyChild{start: start, name: t.Name}
			}
			childDepth++
		case xml.EndElement:
			tokenEnd := int(dec.InputOffset())
			if inBody && childDepth == 0 && t.Name.Local == "body" {
				body.bodyEndStart = endTagStart(docXML, tokenEnd)
				return body, nil
			}
			if !inBody || childDepth == 0 {
				continue
			}
			childDepth--
			if childDepth != 0 || current == nil {
				continue
			}
			current.end = tokenEnd
			if current.name.Local == "p" {
				h, ok, err := parseHeadingParagraph(docXML[current.start:current.end], styleLevels)
				if err != nil {
					return parsedBody{}, err
				}
				if ok {
					h.Index = headingIndex
					headingIndex++
					current.heading = &h
				}
			}
			body.children = append(body.children, *current)
			current = nil
		}
	}
	return parsedBody{}, errors.New("docx: missing word/body")
}
