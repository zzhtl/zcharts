package docx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Heading 是已有 Word 文档中的一个标题目录项。
type Heading struct {
	Index   int
	Level   int
	StyleID string // 标准化样式 id，固定输出为 Heading1..Heading6。
	Text    string
}

// styleInfo 是 word/styles.xml 中一个段落样式的解析中间结果。
// 用 basedOn 链 + outlineLvl/名称回退，把任意自定义样式归一回 1~6 级标题。
type styleInfo struct {
	id      string
	name    string
	basedOn string
	level   int
}

// parseHeadingParagraph 判断一个 <w:p> 段落是否是标题，并提取出级别与文本。
// 级别识别优先级：outlineLvl 直读 > pStyle 命中的样式级别。
func parseHeadingParagraph(data []byte, styleLevels map[string]int) (Heading, bool, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	styleID := ""
	directLevel := 0
	var text strings.Builder
	inText := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Heading{}, false, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "pStyle":
				styleID = attrValue(t.Attr, "val")
			case "outlineLvl":
				if val := attrValue(t.Attr, "val"); val != "" {
					n, err := strconv.Atoi(val)
					if err == nil && n >= 0 && n < len(builtinHeadings) {
						directLevel = n + 1
					}
				}
			case "t":
				inText = true
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				text.Write([]byte(t))
			}
		}
	}

	level := directLevel
	if level == 0 {
		level = styleLevels[styleID]
	}
	if level < 1 || level > len(builtinHeadings) {
		return Heading{}, false, nil
	}
	return Heading{
		Level:   level,
		StyleID: fmt.Sprintf("Heading%d", level),
		Text:    text.String(),
	}, true, nil
}

// headingStyleLevels 返回「样式 id → 标题级别」映射：内置 HeadingN 打底，
// styles.xml 中解析出的自定义样式可以覆盖同名条目。
func headingStyleLevels(stylesXML []byte) map[string]int {
	levels := builtinHeadingStyleLevels()
	for id, level := range parseStyleHeadingLevels(stylesXML) {
		levels[id] = level
	}
	return levels
}

// builtinHeadingStyleLevels 仅返回内置 Heading1..HeadingN 的样式 id → 级别映射。
func builtinHeadingStyleLevels() map[string]int {
	levels := make(map[string]int, len(builtinHeadings))
	for i, h := range builtinHeadings {
		levels[h.id] = i + 1
	}
	return levels
}

// parseStyleHeadingLevels 扫描 word/styles.xml，把所有段落样式归一到 1~6 级。
// outlineLvl 缺失时，会用样式 id / 显示名按 "Heading{n}" / "标题{n}" 回退，
// 再沿 basedOn 链向上追溯，使得自定义样式也能被识别成标题。
func parseStyleHeadingLevels(stylesXML []byte) map[string]int {
	if len(stylesXML) == 0 {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(stylesXML))
	styles := map[string]*styleInfo{}
	var current *styleInfo
	styleDepth := 0

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if current != nil {
				styleDepth++
				switch t.Name.Local {
				case "name":
					current.name = attrValue(t.Attr, "val")
				case "basedOn":
					current.basedOn = attrValue(t.Attr, "val")
				case "outlineLvl":
					if val := attrValue(t.Attr, "val"); val != "" {
						n, err := strconv.Atoi(val)
						if err == nil && n >= 0 && n < len(builtinHeadings) {
							current.level = n + 1
						}
					}
				}
				continue
			}
			if t.Name.Local != "style" || attrValue(t.Attr, "type") != "paragraph" {
				continue
			}
			id := attrValue(t.Attr, "styleId")
			if id == "" {
				continue
			}
			current = &styleInfo{id: id}
			styleDepth = 1
		case xml.EndElement:
			if current == nil {
				continue
			}
			styleDepth--
			if styleDepth == 0 && t.Name.Local == "style" {
				styles[current.id] = current
				current = nil
			}
		}
	}

	for _, st := range styles {
		if st.level == 0 {
			st.level = headingLevelFromName(st.id)
		}
		if st.level == 0 {
			st.level = headingLevelFromName(st.name)
		}
	}
	levels := make(map[string]int, len(styles))
	for id, st := range styles {
		if level := resolveStyleLevel(styles, st, 0); level > 0 {
			levels[id] = level
		}
	}
	return levels
}

// resolveStyleLevel 沿 basedOn 链回溯样式所属的标题级别。
// depth 守门防御 styles.xml 中可能存在的循环引用。
func resolveStyleLevel(styles map[string]*styleInfo, st *styleInfo, depth int) int {
	if st == nil || depth > len(styles) {
		return 0
	}
	if st.level > 0 {
		return st.level
	}
	if st.basedOn == "" {
		return 0
	}
	return resolveStyleLevel(styles, styles[st.basedOn], depth+1)
}

// headingLevelFromName 从样式 id 或显示名中识别 "Heading{n}" / "标题{n}"，
// 用作 outlineLvl 与 basedOn 都缺失时的最后回退手段。
func headingLevelFromName(s string) int {
	compact := strings.ToLower(strings.Join(strings.Fields(s), ""))
	for i := 1; i <= len(builtinHeadings); i++ {
		n := strconv.Itoa(i)
		if compact == "heading"+n || compact == "标题"+n {
			return i
		}
	}
	return 0
}
