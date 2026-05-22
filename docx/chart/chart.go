// Package chart 是 Word docx 中"原生 OOXML chart XML"的占位包。
//
// 第一阶段不实现：当用户调用 docx.AddChart(..., docx.AsNativeChart()) 时，
// 直接返回 errs.ErrNotImplemented。
//
// 后续要补充的内容（指引）：
//
//	- 生成 word/charts/chartN.xml（参考 ECMA-376 Part 1，DrawingML chart）
//	- 在 word/_rels/document.xml.rels 中加 chart relationship
//	- 在 [Content_Types].xml 中 Override chart MIME
//	- 在 document.xml 段落中插入 <w:drawing> 引用 chart
//
// 接口签名预留在此，方便未来填充实现。
package chart

import "github.com/zzhtl/zcharts/common/errs"

// BuildChartXML 接受一个 zcharts option，返回 OOXML chart 部件的 XML 字节。
// 当前总是返回 errs.ErrNotImplemented。
func BuildChartXML(_ any) ([]byte, error) {
	return nil, errs.ErrNotImplemented
}
