package docx

import "fmt"

// drawingXML 生成 OOXML 中"段落内嵌图片"的 <w:drawing> 节点。
//
// rid: 图片在 document.xml.rels 中的 ID
// docPrID: 文档级别的 picture id（用于辅助识别）
// widthEMU/heightEMU: 图片在 Word 中显示的尺寸（English Metric Unit）
func drawingXML(rid string, docPrID int, widthEMU, heightEMU int64) string {
	return fmt.Sprintf(`<w:p><w:r><w:drawing>`+
		`<wp:inline distT="0" distB="0" distL="0" distR="0">`+
		`<wp:extent cx="%d" cy="%d"/>`+
		`<wp:effectExtent l="0" t="0" r="0" b="0"/>`+
		`<wp:docPr id="%d" name="Picture %d"/>`+
		`<wp:cNvGraphicFramePr><a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/></wp:cNvGraphicFramePr>`+
		`<a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">`+
		`<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
		`<pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
		`<pic:nvPicPr>`+
		`<pic:cNvPr id="%d" name="Picture %d"/>`+
		`<pic:cNvPicPr/>`+
		`</pic:nvPicPr>`+
		`<pic:blipFill>`+
		`<a:blip xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" r:embed="%s"/>`+
		`<a:stretch><a:fillRect/></a:stretch>`+
		`</pic:blipFill>`+
		`<pic:spPr>`+
		`<a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm>`+
		`<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`+
		`</pic:spPr>`+
		`</pic:pic>`+
		`</a:graphicData>`+
		`</a:graphic>`+
		`</wp:inline>`+
		`</w:drawing></w:r></w:p>`,
		widthEMU, heightEMU,
		docPrID, docPrID,
		docPrID, docPrID,
		rid,
		widthEMU, heightEMU,
	)
}
