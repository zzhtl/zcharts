package docx

import "encoding/xml"

// attrValue 按 local name（忽略命名空间前缀）取属性值，缺失返回空串。
func attrValue(attrs []xml.Attr, name string) string {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

// tagStart 给定 xml.Decoder.InputOffset() 返回的 token 末尾偏移，
// 向左回溯到对应起始标签的 '<' 处，得到该标签在原始字节流中的起点。
func tagStart(data []byte, tokenEnd int) int {
	for i := tokenEnd - 1; i >= 0; i-- {
		if data[i] == '<' {
			return i
		}
	}
	return 0
}

// endTagStart 给定 token 末尾偏移，向左回溯到对应结束标签的 '</' 处。
// 用于精确定位 </w:body> 之类闭合标签的起点。
func endTagStart(data []byte, tokenEnd int) int {
	for i := tokenEnd - 2; i >= 0; i-- {
		if data[i] == '<' && data[i+1] == '/' {
			return i
		}
	}
	return tokenEnd
}
