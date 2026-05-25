package chart

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/zzhtl/zcharts/common/errs"
	"github.com/zzhtl/zcharts/option"
)

// BuildShapeXML 使用 Word/WPS 可解析的 VML 形状绘制非标准 chart 类型。
func BuildShapeXML(v any, width, height int) (string, error) {
	opt, ok := v.(*option.Option)
	if !ok || opt == nil {
		return "", fmt.Errorf("docx/chart: %w", errs.ErrInvalidOption)
	}
	if len(opt.Series) != 1 {
		return "", fmt.Errorf("docx/chart: %w: shape drawing expects single series", errs.ErrUnsupportedSeries)
	}
	if width <= 0 {
		width = 600
	}
	if height <= 0 {
		height = 400
	}

	var (
		body string
		err  error
	)
	switch s := opt.Series[0].(type) {
	case *option.HeatmapSeries:
		body, err = buildHeatmapShape(opt, s, width, height)
	case *option.GaugeSeries:
		body, err = buildGaugeShape(opt, s, width, height)
	case *option.FunnelSeries:
		body, err = buildFunnelShape(opt, s, width, height)
	case *option.TimelineSeries:
		body, err = buildTimelineShape(opt, s, width, height)
	case *option.WordCloudSeries:
		body, err = buildWordCloudShape(opt, s, width, height)
	default:
		return "", fmt.Errorf("docx/chart: %w: %s", errs.ErrUnsupportedSeries, opt.Series[0].Kind())
	}
	if err != nil {
		return "", err
	}
	return vmlParagraph(width, height, body), nil
}

func buildHeatmapShape(opt *option.Option, s *option.HeatmapSeries, width, height int) (string, error) {
	if len(s.Data) == 0 {
		return "", fmt.Errorf("docx/chart: %w: empty heatmap data", errs.ErrInvalidOption)
	}
	var b strings.Builder
	top := writeShapeTitle(&b, opt, width)
	left, right, bottom := 72, 24, 44
	plotW := maxInt(1, width-left-right)
	plotH := maxInt(1, height-top-bottom)

	xCats := axisCategories(opt.XAxis.First().Data)
	yCats := axisCategories(opt.YAxis.First().Data)
	maxX, maxY := heatmapMaxIndex(s.Data)
	if len(xCats) == 0 {
		xCats = indexLabels(maxX + 1)
	}
	if len(yCats) == 0 {
		yCats = indexLabels(maxY + 1)
	}
	if len(xCats) == 0 || len(yCats) == 0 {
		return "", fmt.Errorf("docx/chart: %w: empty heatmap axis", errs.ErrInvalidOption)
	}

	minV, maxV := heatmapRange(opt, s.Data)
	cellW := float64(plotW) / float64(len(xCats))
	cellH := float64(plotH) / float64(len(yCats))
	for _, d := range s.Data {
		if len(d.Values) < 3 {
			continue
		}
		xi, yi := int(d.Values[0]), int(d.Values[1])
		if xi < 0 || xi >= len(xCats) || yi < 0 || yi >= len(yCats) {
			continue
		}
		t := 0.0
		if maxV > minV {
			t = (d.Values[2] - minV) / (maxV - minV)
		}
		x := left + round(float64(xi)*cellW)
		y := top + round(float64(yi)*cellH)
		w := maxInt(1, round(cellW))
		h := maxInt(1, round(cellH))
		b.WriteString(vmlRect(x, y, w, h, heatmapColor(opt, t), "#FFFFFF"))
		// 单元格足够大时显示数值；深色格用白字，浅色格用深字。
		if w >= 30 && h >= 18 {
			txtColor := "#333333"
			if t > 0.55 {
				txtColor = "#FFFFFF"
			}
			b.WriteString(vmlText(x, y+h/2-8, w, 16, strconv.FormatFloat(d.Values[2], 'f', -1, 64), txtColor, 9, false))
		}
	}
	for i, label := range xCats {
		x := left + round((float64(i)+0.5)*cellW) - 28
		b.WriteString(vmlText(x, height-bottom+8, 56, 18, label, "#555555", 9, false))
	}
	for i, label := range yCats {
		y := top + round((float64(i)+0.5)*cellH) - 9
		b.WriteString(vmlText(8, y, left-14, 18, label, "#555555", 9, false))
	}
	b.WriteString(vmlRect(left, top, plotW, plotH, "none", "#D9DDE7"))
	return b.String(), nil
}

func buildGaugeShape(opt *option.Option, s *option.GaugeSeries, width, height int) (string, error) {
	if len(s.Data) == 0 {
		return "", fmt.Errorf("docx/chart: %w: empty gauge data", errs.ErrInvalidOption)
	}
	var b strings.Builder
	top := writeShapeTitle(&b, opt, width)
	minV, maxV := s.Min, s.Max
	if maxV <= minV {
		minV, maxV = 0, 100
	}
	value := s.Data[0].Number()
	ratio := clampFloat((value-minV)/(maxV-minV), 0, 1)
	start, end := 225.0, -45.0
	if s.StartAngle != nil {
		start = *s.StartAngle
	}
	if s.EndAngle != nil {
		end = *s.EndAngle
	}

	name := s.Data[0].Name
	if name == "" {
		name = s.Name
	}
	valueText := strconv.FormatFloat(value, 'f', -1, 64)
	valueH, nameH := 26, 18
	labelBlockH := valueH
	if name != "" {
		labelBlockH += nameH + 2
	}
	labelTop := height - labelBlockH - 6
	if labelTop < top+44 {
		labelTop = height - labelBlockH
	}
	plotBottom := labelTop - 8
	if plotBottom < top+40 {
		plotBottom = top + 40
	}

	cx := width / 2
	cy := top + round(float64(plotBottom-top)*0.58)
	radius := minInt(width/2-48, int(float64(plotBottom-cy-8)/0.72))
	radius = minInt(radius, cy-top-6)
	if radius < 40 {
		radius = 40
	}
	if cy+round(float64(radius)*0.72)+8 > labelTop {
		radius = maxInt(32, int(float64(labelTop-cy-8)/0.72))
	}
	segments := 64
	activeColor := paletteShapeColor(opt, 0)
	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments-1)
		angle := start + (end-start)*t
		color := "#D8DDE8"
		if t <= ratio {
			color = activeColor
		}
		x1, y1 := polarPoint(cx, cy, radius-14, angle)
		x2, y2 := polarPoint(cx, cy, radius, angle)
		b.WriteString(vmlLine(x1, y1, x2, y2, color, 4))
	}
	needleAngle := start + (end-start)*ratio
	nx, ny := polarPoint(cx, cy, maxInt(8, radius-28), needleAngle)
	b.WriteString(vmlLine(cx, cy, nx, ny, "#333333", 2))
	b.WriteString(vmlOval(cx-5, cy-5, 10, 10, "#333333", "#333333"))

	b.WriteString(vmlText(cx-90, labelTop, 180, valueH, valueText, "#222222", 16, true))
	if name != "" {
		b.WriteString(vmlText(cx-110, labelTop+valueH+2, 220, nameH, name, "#666666", 10, false))
	}
	return b.String(), nil
}

func buildFunnelShape(opt *option.Option, s *option.FunnelSeries, width, height int) (string, error) {
	if len(s.Data) == 0 {
		return "", fmt.Errorf("docx/chart: %w: empty funnel data", errs.ErrInvalidOption)
	}
	var b strings.Builder
	top := writeShapeTitle(&b, opt, width)

	items := append([]option.DataValue(nil), s.Data...)
	switch strings.ToLower(s.Sort) {
	case "ascending":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Number() < items[j].Number() })
	case "none":
	default:
		sort.SliceStable(items, func(i, j int) bool { return items[i].Number() > items[j].Number() })
	}
	maxV, totalV := 0.0, 0.0
	for _, d := range items {
		if d.Number() > maxV {
			maxV = d.Number()
		}
		totalV += d.Number()
	}
	if maxV <= 0 {
		maxV = 1
	}
	if totalV <= 0 {
		totalV = 1
	}

	// 预先拼好每段标签（名称 + 值 + 百分比），并按最长标签预留右侧标签列宽度。
	// 所有标签统一画在漏斗右侧、深色字、完整可见，避免段内白字溢出到色块外被截断。
	const labelSize = 10
	labels := make([]string, len(items))
	maxLabelW := 0
	for i, d := range items {
		name := dataName(d)
		valueText := strconv.FormatFloat(d.Number(), 'f', -1, 64)
		pct := strconv.FormatFloat(d.Number()/totalV*100, 'f', 1, 64)
		// 统一格式：名称:数量(百分比%)
		label := valueText + "(" + pct + "%)"
		if name != "" {
			label = name + ":" + label
		}
		labels[i] = label
		if w := estimateTextWidth(label, labelSize); w > maxLabelW {
			maxLabelW = w
		}
	}

	leftMargin, gapToLabel, bottom := 16, 14, 24
	labelZoneW := minInt(maxLabelW+8, (width-leftMargin)/2)
	plotW := maxInt(60, width-leftMargin-gapToLabel-labelZoneW)
	plotH := maxInt(1, height-top-bottom)
	centerX := leftMargin + plotW/2
	labelX := centerX + plotW/2 + gapToLabel

	gap := int(s.Gap)
	if gap <= 0 {
		gap = 4
	}
	segH := maxInt(1, (plotH-gap*(len(items)-1))/len(items))
	// 用居中、宽度随数值递减的矩形条堆叠近似漏斗：跨渲染器（Word/WPS/LibreOffice）稳定，
	// 而 VML 多边形 path 在部分渲染器（LibreOffice）会被错误缩放。
	for i, d := range items {
		segW := funnelWidth(d.Number(), maxV, plotW)
		y1 := top + i*(segH+gap)
		fill := paletteShapeColor(opt, i)
		b.WriteString(vmlRect(centerX-segW/2, y1, segW, segH, fill, "#FFFFFF"))
		midY := y1 + segH/2
		// 标签统一画在漏斗右侧、深色字（不画引导线，保证 Word/WPS 的 VML 兼容性）。
		b.WriteString(vmlText(labelX, midY-9, maxInt(40, width-labelX-2), 18, labels[i], "#333333", labelSize, false))
	}
	return b.String(), nil
}

func buildTimelineShape(opt *option.Option, s *option.TimelineSeries, width, height int) (string, error) {
	if len(s.Data) == 0 {
		return "", fmt.Errorf("docx/chart: %w: empty timeline data", errs.ErrInvalidOption)
	}
	var b strings.Builder
	top := writeShapeTitle(&b, opt, width)
	lineX := maxInt(120, width/3)
	bottom := 32
	available := maxInt(1, height-top-bottom)
	step := 0
	if len(s.Data) > 1 {
		step = available / (len(s.Data) - 1)
	}
	if step <= 0 {
		step = 56
	}
	startY := top + 10
	endY := minInt(height-bottom, startY+step*(len(s.Data)-1))
	lineColor := paletteShapeColor(opt, 0)
	b.WriteString(vmlLine(lineX, startY, lineX, endY, lineColor, 2))
	for i, item := range s.Data {
		y := minInt(height-bottom, startY+i*step)
		b.WriteString(vmlOval(lineX-6, y-6, 12, 12, "#FFFFFF", lineColor))
		b.WriteString(vmlText(12, y-10, lineX-28, 20, item.Time, "#666666", 9, false))
		title := item.Title
		if title == "" {
			title = item.Name
		}
		b.WriteString(vmlText(lineX+18, y-16, width-lineX-32, 22, title, "#222222", 11, true))
		b.WriteString(vmlText(lineX+18, y+6, width-lineX-32, 34, item.Content, "#666666", 9, false))
	}
	return b.String(), nil
}

func buildWordCloudShape(opt *option.Option, s *option.WordCloudSeries, width, height int) (string, error) {
	if len(s.Data) == 0 {
		return "", fmt.Errorf("docx/chart: %w: empty wordCloud data", errs.ErrInvalidOption)
	}
	var b strings.Builder
	top := writeShapeTitle(&b, opt, width)
	items := append([]option.DataValue(nil), s.Data...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].Number() > items[j].Number() })
	minV, maxV := valueRange(items)
	minSize, maxSize := 14.0, 38.0
	if len(s.SizeRange) >= 2 {
		minSize, maxSize = s.SizeRange[0], s.SizeRange[1]
	}
	x, y := 28, top+4
	lineH := 0
	for i, d := range items {
		word := dataName(d)
		if word == "" {
			continue
		}
		t := 0.0
		if maxV > minV {
			t = (d.Number() - minV) / (maxV - minV)
		}
		size := int(minSize + (maxSize-minSize)*t)
		if size < 8 {
			size = 8
		}
		boxW := int(float64(len([]rune(word))*size)*0.9) + 18
		boxH := size + 10
		if x+boxW > width-20 {
			x = 28
			y += lineH + 8
			lineH = 0
		}
		if y+boxH > height-16 {
			break
		}
		b.WriteString(vmlWordText(x, y, boxW, boxH, word, paletteShapeColor(opt, i), size, i < 3))
		x += boxW + 10
		if boxH > lineH {
			lineH = boxH
		}
	}
	return b.String(), nil
}

func vmlParagraph(width, height int, body string) string {
	return fmt.Sprintf(`<w:p><w:r><w:pict><v:group coordorigin="0,0" coordsize="%d,%d" style="width:%dpx;height:%dpx">%s</v:group></w:pict></w:r></w:p>`, width, height, width, height, body)
}

func writeShapeTitle(b *strings.Builder, opt *option.Option, width int) int {
	if opt != nil && opt.Title != nil && opt.Title.Text != "" {
		b.WriteString(vmlText(20, 12, width-40, 24, opt.Title.Text, "#222222", 14, true))
		return 48
	}
	return 24
}

func vmlRect(x, y, width, height int, fill, stroke string) string {
	if fill == "none" {
		return fmt.Sprintf(`<v:rect style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx" filled="f" strokecolor="%s"/>`, x, y, width, height, shapeColor(stroke, "#D9DDE7"))
	}
	stroked := ` stroked="f"`
	strokeAttr := ""
	if stroke != "" {
		stroked = ""
		strokeAttr = fmt.Sprintf(` strokecolor="%s"`, shapeColor(stroke, fill))
	}
	return fmt.Sprintf(`<v:rect style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx" fillcolor="%s"%s%s/>`, x, y, width, height, shapeColor(fill, "#FFFFFF"), strokeAttr, stroked)
}

func vmlOval(x, y, width, height int, fill, stroke string) string {
	return fmt.Sprintf(`<v:oval style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx" fillcolor="%s" strokecolor="%s"/>`, x, y, width, height, shapeColor(fill, "#FFFFFF"), shapeColor(stroke, "#333333"))
}

func vmlLine(x1, y1, x2, y2 int, color string, weight int) string {
	if weight <= 0 {
		weight = 1
	}
	return fmt.Sprintf(`<v:line from="%d,%d" to="%d,%d" strokecolor="%s" strokeweight="%dpt"/>`, x1, y1, x2, y2, shapeColor(color, "#333333"), weight)
}

func vmlText(x, y, width, height int, text, color string, size int, bold bool) string {
	if text == "" || width <= 0 || height <= 0 {
		return ""
	}
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	boldXML := ""
	if bold {
		boldXML = `<w:b/>`
	}
	wordColor := strings.TrimPrefix(shapeColor(color, "#333333"), "#")
	return fmt.Sprintf(`<v:shape style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx" stroked="f" filled="f"><v:textbox inset="0,0,0,0"><w:txbxContent><w:p><w:r><w:rPr>%s<w:color w:val="%s"/><w:sz w:val="%d"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p></w:txbxContent></v:textbox></v:shape>`, x, y, width, height, boldXML, wordColor, size*2, escape(text))
}

func vmlWordText(x, y, width, height int, text, color string, size int, bold bool) string {
	if text == "" || width <= 0 || height <= 0 {
		return ""
	}
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	boldXML := ""
	if bold {
		boldXML = `<w:b/>`
	}
	wordColor := strings.TrimPrefix(shapeColor(color, "#333333"), "#")
	return fmt.Sprintf(`<v:shape style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx;mso-wrap-style:none" stroked="f" filled="f"><v:textbox inset="0,0,0,0" style="mso-fit-shape-to-text:t"><w:txbxContent><w:p><w:r><w:rPr>%s<w:noProof/><w:color w:val="%s"/><w:sz w:val="%d"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p></w:txbxContent></v:textbox></v:shape>`, x, y, width, height, boldXML, wordColor, size*2, escape(text))
}

// estimateTextWidth 估算文字像素宽度（偏大以确保 Word 文本框够宽、单行不被裁）。
// size 为字号(pt)，像素宽约为 pt 的 1.33 倍：CJK/全角按 1.5×，其余按 0.75×。
func estimateTextWidth(s string, size int) int {
	w := 0.0
	for _, r := range s {
		if r > 0x2E7F {
			w += float64(size) * 1.5
		} else {
			w += float64(size) * 0.75
		}
	}
	return int(w)
}

func polarPoint(cx, cy, radius int, angle float64) (int, int) {
	rad := angle * math.Pi / 180
	return cx + round(math.Cos(rad)*float64(radius)), cy - round(math.Sin(rad)*float64(radius))
}

func heatmapMaxIndex(data []option.DataValue) (int, int) {
	maxX, maxY := -1, -1
	for _, d := range data {
		if len(d.Values) < 2 {
			continue
		}
		if x := int(d.Values[0]); x > maxX {
			maxX = x
		}
		if y := int(d.Values[1]); y > maxY {
			maxY = y
		}
	}
	return maxX, maxY
}

func heatmapRange(opt *option.Option, data []option.DataValue) (float64, float64) {
	if opt != nil && opt.VisualMap != nil && opt.VisualMap.Max > opt.VisualMap.Min {
		return opt.VisualMap.Min, opt.VisualMap.Max
	}
	minV, maxV := math.Inf(1), math.Inf(-1)
	for _, d := range data {
		if len(d.Values) < 3 {
			continue
		}
		if d.Values[2] < minV {
			minV = d.Values[2]
		}
		if d.Values[2] > maxV {
			maxV = d.Values[2]
		}
	}
	if math.IsInf(minV, 0) || maxV <= minV {
		return 0, 1
	}
	return minV, maxV
}

func heatmapColor(opt *option.Option, t float64) string {
	start, end := "#EEF3FF", paletteShapeColor(opt, 0)
	if opt != nil && opt.VisualMap != nil && len(opt.VisualMap.InRange.Color) > 0 {
		start = shapeColor(opt.VisualMap.InRange.Color[0], start)
		end = shapeColor(opt.VisualMap.InRange.Color[len(opt.VisualMap.InRange.Color)-1], end)
	}
	return gradientColor(start, end, clampFloat(t, 0, 1))
}

func gradientColor(start, end string, t float64) string {
	sr, sg, sb := rgb(start)
	er, eg, eb := rgb(end)
	r := int(float64(sr) + (float64(er)-float64(sr))*t)
	g := int(float64(sg) + (float64(eg)-float64(sg))*t)
	b := int(float64(sb) + (float64(eb)-float64(sb))*t)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func rgb(hex string) (int, int, int) {
	h := normalizeHexColor(hex)
	if len(h) != 6 {
		return 84, 112, 198
	}
	r, _ := strconv.ParseInt(h[0:2], 16, 64)
	g, _ := strconv.ParseInt(h[2:4], 16, 64)
	b, _ := strconv.ParseInt(h[4:6], 16, 64)
	return int(r), int(g), int(b)
}

func axisCategories(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func indexLabels(n int) []string {
	if n <= 0 {
		return nil
	}
	out := make([]string, n)
	for i := range out {
		out[i] = strconv.Itoa(i + 1)
	}
	return out
}

func funnelWidth(value, maxValue float64, maxWidth int) int {
	if maxValue <= 0 {
		return maxWidth
	}
	w := int(float64(maxWidth) * clampFloat(value/maxValue, 0.08, 1))
	return maxInt(24, w)
}

func valueRange(data []option.DataValue) (float64, float64) {
	minV, maxV := math.Inf(1), math.Inf(-1)
	for _, d := range data {
		v := d.Number()
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if math.IsInf(minV, 0) {
		return 0, 1
	}
	if maxV <= minV {
		return minV, minV + 1
	}
	return minV, maxV
}

func dataName(d option.DataValue) string {
	if d.Name != "" {
		return d.Name
	}
	return d.Label
}

func paletteShapeColor(opt *option.Option, index int) string {
	return "#" + paletteAt(paletteFor(opt), index)
}

func shapeColor(value, fallback string) string {
	if hex := normalizeHexColor(value); hex != "" {
		return "#" + hex
	}
	if hex := normalizeHexColor(fallback); hex != "" {
		return "#" + hex
	}
	return "#5470C6"
}

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func round(v float64) int {
	if v >= 0 {
		return int(v + 0.5)
	}
	return int(v - 0.5)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
