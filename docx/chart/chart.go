// Package chart 生成 Word docx 中的原生 OOXML chart XML 和形状绘制 XML。
//
// 该包面向 docx.AsNativeChart() 路径，覆盖 Word 原生图表里最常用的
// line/bar/pie/scatter/radar；部分非标准 chart 类型用 Word/WPS 形状绘制。
package chart

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/zzhtl/zcharts/common/errs"
	"github.com/zzhtl/zcharts/common/number"
	"github.com/zzhtl/zcharts/option"
)

const (
	catAxisID      = 12345678
	valAxisID      = 12345679
	scatterXAxisID = 12345680
	scatterYAxisID = 12345681
)

var defaultPalette = []string{"5470C6", "91CC75", "FAC858", "EE6666", "73C0DE", "3BA272", "FC8452", "9A60B4", "EA7CCC"}

// BuildChartXML 接受一个 zcharts option，返回 OOXML chart 部件的 XML 字节。
//
// 当前支持 line/bar/pie/scatter/radar；热力、仪表盘、漏斗、词云等图表请走图片嵌入路径。
func BuildChartXML(v any) ([]byte, error) {
	opt, ok := v.(*option.Option)
	if !ok || opt == nil {
		return nil, fmt.Errorf("docx/chart: %w", errs.ErrInvalidOption)
	}
	if len(opt.Series) == 0 {
		return nil, fmt.Errorf("docx/chart: %w: empty series", errs.ErrInvalidOption)
	}

	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<c:chartSpace xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart" `)
	b.WriteString(`xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" `)
	b.WriteString(`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	b.WriteString(`<c:date1904 val="0"/><c:lang val="zh-CN"/><c:roundedCorners val="0"/>`)
	b.WriteString(`<c:chart>`)
	if opt.Title != nil && opt.Title.Text != "" {
		writeTitle(&b, opt.Title.Text)
	}
	b.WriteString(`<c:plotArea><c:layout/>`)

	if err := writePlotArea(&b, opt); err != nil {
		return nil, err
	}

	b.WriteString(`</c:plotArea>`)
	if legendVisible(opt) {
		fmt.Fprintf(&b, `<c:legend><c:legendPos val="%s"/><c:layout/><c:overlay val="0"/></c:legend>`, legendPos(opt))
	}
	b.WriteString(`<c:plotVisOnly val="1"/><c:dispBlanksAs val="gap"/>`)
	b.WriteString(`</c:chart>`)
	b.WriteString(`<c:printSettings><c:headerFooter/><c:pageMargins b="0.75" l="0.7" r="0.7" t="0.75" header="0.3" footer="0.3"/><c:pageSetup/></c:printSettings>`)
	b.WriteString(`</c:chartSpace>`)
	return b.Bytes(), nil
}

// legendVisible 决定 Word 原生图表是否显示图例。
// 与图片渲染路径保持一致：默认显示，仅在显式 show:false 时隐藏。
func legendVisible(opt *option.Option) bool {
	if opt.Legend == nil {
		return true
	}
	return opt.Legend.IsShown()
}

// legendPos 把 legend 配置映射为 OOXML 图例位置（t/b/l/r），默认顶部，
// 与图片路径的默认顶部居中一致。
func legendPos(opt *option.Option) string {
	l := opt.Legend
	if l == nil {
		return "t"
	}
	if strings.EqualFold(l.Orient, "vertical") {
		if l.Left.Set || l.Left.Keyword == "left" {
			return "l"
		}
		return "r"
	}
	if l.Bottom.Set || l.Top.Keyword == "bottom" {
		return "b"
	}
	return "t"
}

func writePlotArea(b *bytes.Buffer, opt *option.Option) error {
	lines := make([]*option.LineSeries, 0)
	bars := make([]*option.BarSeries, 0)
	pies := make([]*option.PieSeries, 0)
	scatters := make([]*option.ScatterSeries, 0)
	radars := make([]*option.RadarSeries, 0)
	for _, s := range opt.Series {
		switch v := s.(type) {
		case *option.LineSeries:
			lines = append(lines, v)
		case *option.BarSeries:
			bars = append(bars, v)
		case *option.PieSeries:
			pies = append(pies, v)
		case *option.ScatterSeries:
			scatters = append(scatters, v)
		case *option.RadarSeries:
			radars = append(radars, v)
		default:
			return fmt.Errorf("docx/chart: %w: %s", errs.ErrUnsupportedSeries, s.Kind())
		}
	}

	palette := paletteFor(opt)
	if len(pies) > 0 {
		if len(pies) != len(opt.Series) || len(pies) > 1 {
			return fmt.Errorf("docx/chart: %w: mixed or multiple pie series", errs.ErrUnsupportedSeries)
		}
		writePieChart(b, pies[0], palette)
		return nil
	}
	if len(scatters) > 0 {
		if len(scatters) != len(opt.Series) {
			return fmt.Errorf("docx/chart: %w: mixed scatter chart", errs.ErrUnsupportedSeries)
		}
		writeScatterChart(b, scatters, palette)
		writeValueAxisWithID(b, scatterXAxisID, scatterYAxisID, "b")
		writeValueAxisWithID(b, scatterYAxisID, scatterXAxisID, "l")
		return nil
	}
	if len(radars) > 0 {
		if len(radars) != len(opt.Series) {
			return fmt.Errorf("docx/chart: %w: mixed radar chart", errs.ErrUnsupportedSeries)
		}
		if err := writeRadarChart(b, radars, opt.Radar, palette); err != nil {
			return err
		}
		writeCategoryAxis(b)
		writeValueAxis(b)
		return nil
	}
	if len(bars) == 0 && len(lines) == 0 {
		return fmt.Errorf("docx/chart: %w: empty native chart", errs.ErrInvalidOption)
	}
	// 折线再分为"面积图"（带 areaStyle）与普通折线。
	var areaLines, plainLines []*option.LineSeries
	for _, l := range lines {
		if l.AreaStyle != nil {
			areaLines = append(areaLines, l)
		} else {
			plainLines = append(plainLines, l)
		}
	}
	// 横向条形图：纯柱状且 Y 轴为 category 时，按 ECharts 习惯横向绘制。
	horizontal := len(bars) > 0 && len(lines) == 0 &&
		len(opt.YAxis) > 0 && strings.EqualFold(opt.YAxis[0].Type, option.AxisCategory)
	categories := categoriesFor(opt)
	if horizontal {
		categories = yCategoriesFor(opt)
	}

	order := 0
	if len(bars) > 0 {
		barDir := "col"
		if horizontal {
			barDir = "bar"
		}
		writeBarChart(b, bars, categories, order, palette, barDir)
		order += len(bars)
	}
	if len(areaLines) > 0 {
		writeAreaChart(b, areaLines, categories, order, palette)
		order += len(areaLines)
	}
	if len(plainLines) > 0 {
		writeLineChart(b, plainLines, categories, order, palette)
	}

	catPos, valPos := "b", "l"
	if horizontal {
		catPos, valPos = "l", "b"
	}
	writeCategoryAxisAt(b, catPos)
	writeValueAxisWithID(b, valAxisID, catAxisID, valPos)
	return nil
}

func writeTitle(b *bytes.Buffer, text string) {
	b.WriteString(`<c:title><c:tx><c:rich><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr lang="zh-CN"/>`)
	fmt.Fprintf(b, `<a:t>%s</a:t>`, escape(text))
	b.WriteString(`</a:r></a:p></c:rich></c:tx><c:layout/><c:overlay val="0"/></c:title>`)
}

func writeChartRichText(b *bytes.Buffer, tag, text string, size int) {
	if size <= 0 {
		size = 900
	}
	fmt.Fprintf(b, `<c:%s><c:rich><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr lang="zh-CN" sz="%d"/>`, tag, size)
	fmt.Fprintf(b, `<a:t>%s</a:t>`, escape(text))
	b.WriteString(`</a:r></a:p></c:rich></c:` + tag + `>`)
}

func writeBarChart(b *bytes.Buffer, series []*option.BarSeries, categories []string, orderStart int, palette []string, barDir string) {
	if barDir == "" {
		barDir = "col"
	}
	grouping := "clustered"
	for _, s := range series {
		if s.Stack != "" {
			grouping = "stacked"
			break
		}
	}
	fmt.Fprintf(b, `<c:barChart><c:barDir val="%s"/>`, barDir)
	fmt.Fprintf(b, `<c:grouping val="%s"/>`, grouping)
	b.WriteString(`<c:varyColors val="0"/>`)
	for i, s := range series {
		writeCartesianSeries(b, s.Name, s.Data, categories, orderStart+i, false, false, paletteAt(palette, orderStart+i))
	}
	if grouping == "stacked" {
		b.WriteString(`<c:overlap val="100"/>`)
	}
	writeAxisRefs(b)
	b.WriteString(`</c:barChart>`)
}

// writeAreaChart 生成面积图（带 areaStyle 的折线 series）。stack 非空时为堆积面积。
func writeAreaChart(b *bytes.Buffer, series []*option.LineSeries, categories []string, orderStart int, palette []string) {
	grouping := "standard"
	for _, s := range series {
		if s.Stack != "" {
			grouping = "stacked"
			break
		}
	}
	b.WriteString(`<c:areaChart>`)
	fmt.Fprintf(b, `<c:grouping val="%s"/>`, grouping)
	b.WriteString(`<c:varyColors val="0"/>`)
	for i, s := range series {
		writeAreaSeries(b, s.Name, s.Data, categories, orderStart+i, paletteAt(palette, orderStart+i))
	}
	writeAxisRefs(b)
	b.WriteString(`</c:areaChart>`)
}

func writeAreaSeries(b *bytes.Buffer, name string, data []option.DataValue, categories []string, order int, color string) {
	b.WriteString(`<c:ser>`)
	fmt.Fprintf(b, `<c:idx val="%d"/><c:order val="%d"/>`, order, order)
	writeSeriesText(b, name)
	// 半透明填充 + 同色描边。
	fmt.Fprintf(b, `<c:spPr><a:solidFill><a:srgbClr val="%s"><a:alpha val="55000"/></a:srgbClr></a:solidFill><a:ln w="19050"><a:solidFill><a:srgbClr val="%s"/></a:solidFill></a:ln></c:spPr>`, color, color)
	writeStringLiteral(b, "cat", alignCategories(categories, data))
	writeNumberLiteral(b, "val", valuesOf(data))
	b.WriteString(`</c:ser>`)
}

func writeLineChart(b *bytes.Buffer, series []*option.LineSeries, categories []string, orderStart int, palette []string) {
	b.WriteString(`<c:lineChart><c:grouping val="standard"/><c:varyColors val="0"/>`)
	for i, s := range series {
		writeCartesianSeries(b, s.Name, s.Data, categories, orderStart+i, true, s.Smooth, paletteAt(palette, orderStart+i))
	}
	writeAxisRefs(b)
	b.WriteString(`</c:lineChart>`)
}

func writePieChart(b *bytes.Buffer, s *option.PieSeries, palette []string) {
	chartTag := "pieChart"
	if isDoughnutPie(s) {
		chartTag = "doughnutChart"
	}
	fmt.Fprintf(b, `<c:%s><c:varyColors val="1"/>`, chartTag)
	b.WriteString(`<c:ser>`)
	b.WriteString(`<c:idx val="0"/><c:order val="0"/>`)
	writeSeriesText(b, s.Name)
	for i := range s.Data {
		writeDataPointShape(b, i, paletteAt(palette, i))
	}
	writePieDataLabels(b, s, chartTag)
	writeStringLiteral(b, "cat", pieCategories(s.Data))
	writeNumberLiteral(b, "val", valuesOf(s.Data))
	b.WriteString(`</c:ser><c:firstSliceAng val="0"/>`)
	if chartTag == "doughnutChart" {
		fmt.Fprintf(b, `<c:holeSize val="%d"/>`, doughnutHoleSize(s))
	}
	fmt.Fprintf(b, `</c:%s>`, chartTag)
}

func writePieDataLabels(b *bytes.Buffer, s *option.PieSeries, chartTag string) {
	// 显式写入每个标签文本，避免 Word/WPS 仅靠 showPercent 自行计算时不显示。
	// 标签放在外侧并开启引导线，给「名称: 百分比」保留完整排版空间。
	dLblPos := "outEnd"
	if chartTag == "doughnutChart" {
		dLblPos = "ctr"
	}
	total := 0.0
	for _, d := range s.Data {
		if v := d.Number(); v > 0 {
			total += v
		}
	}
	b.WriteString(`<c:dLbls>`)
	compact := chartTag == "doughnutChart"
	labelSize := 900
	if compact {
		labelSize = 700
	}
	for i, d := range s.Data {
		label := pieDataLabelText(s, d, total, compact)
		if label == "" {
			continue
		}
		fmt.Fprintf(b, `<c:dLbl><c:idx val="%d"/>`, i)
		writeChartRichText(b, "tx", label, labelSize)
		fmt.Fprintf(b, `<c:dLblPos val="%s"/>`, dLblPos)
		b.WriteString(`<c:showLegendKey val="0"/><c:showVal val="0"/><c:showCatName val="0"/><c:showSerName val="0"/><c:showPercent val="0"/><c:showBubbleSize val="0"/></c:dLbl>`)
	}
	fmt.Fprintf(b, `<c:dLblPos val="%s"/>`, dLblPos)
	b.WriteString(`<c:showLegendKey val="0"/><c:showVal val="0"/><c:showCatName val="0"/><c:showSerName val="0"/><c:showPercent val="0"/><c:showBubbleSize val="0"/><c:showLeaderLines val="1"/></c:dLbls>`)
}

func pieDataLabelText(s *option.PieSeries, d option.DataValue, total float64, compact bool) string {
	value := d.Number()
	percent := 0.0
	if total > 0 && value > 0 {
		percent = value / total
	}
	if compact {
		return number.FormatFloat(percent*100, 1) + "%"
	}
	if s.Label.Formatter != "" {
		r := s.Label.Formatter
		r = strings.ReplaceAll(r, "{a}", s.Name)
		r = strings.ReplaceAll(r, "{b}", dataName(d))
		r = strings.ReplaceAll(r, "{c}", number.FormatAuto(value))
		r = strings.ReplaceAll(r, "{d}", number.FormatFloat(percent*100, 2))
		return r
	}
	return number.FormatFloat(percent*100, 1) + "%"
}

func writeScatterChart(b *bytes.Buffer, series []*option.ScatterSeries, palette []string) {
	b.WriteString(`<c:scatterChart><c:scatterStyle val="marker"/><c:varyColors val="0"/>`)
	for i, s := range series {
		writeScatterSeries(b, s, i, paletteAt(palette, i))
	}
	fmt.Fprintf(b, `<c:axId val="%d"/><c:axId val="%d"/>`, scatterXAxisID, scatterYAxisID)
	b.WriteString(`</c:scatterChart>`)
}

func writeScatterSeries(b *bytes.Buffer, s *option.ScatterSeries, order int, color string) {
	b.WriteString(`<c:ser>`)
	fmt.Fprintf(b, `<c:idx val="%d"/><c:order val="%d"/>`, order, order)
	writeSeriesText(b, s.Name)
	b.WriteString(`<c:spPr><a:ln><a:noFill/></a:ln></c:spPr>`)
	writeMarker(b, color, s.SymbolSize)
	writeNumberLiteral(b, "xVal", xValuesOf(s.Data))
	writeNumberLiteral(b, "yVal", yValuesOf(s.Data))
	b.WriteString(`<c:smooth val="0"/>`)
	b.WriteString(`</c:ser>`)
}

func writeRadarChart(b *bytes.Buffer, series []*option.RadarSeries, radar *option.Radar, palette []string) error {
	categories := radarCategories(series, radar)
	if len(categories) == 0 {
		return fmt.Errorf("docx/chart: %w: empty radar indicator", errs.ErrInvalidOption)
	}
	b.WriteString(`<c:radarChart><c:radarStyle val="marker"/><c:varyColors val="0"/>`)
	order := 0
	for _, s := range series {
		for _, item := range s.Data {
			name := item.Name
			if name == "" {
				name = s.Name
			}
			writeRadarSeries(b, name, item.Value, categories, order, paletteAt(palette, order))
			order++
		}
	}
	if order == 0 {
		return fmt.Errorf("docx/chart: %w: empty radar data", errs.ErrInvalidOption)
	}
	writeAxisRefs(b)
	b.WriteString(`</c:radarChart>`)
	return nil
}

func writeRadarSeries(b *bytes.Buffer, name string, values []float64, categories []string, order int, color string) {
	b.WriteString(`<c:ser>`)
	fmt.Fprintf(b, `<c:idx val="%d"/><c:order val="%d"/>`, order, order)
	writeSeriesText(b, name)
	writeSeriesShape(b, color, true)
	writeMarker(b, color, 5)
	writeStringLiteral(b, "cat", categories)
	writeNumberLiteral(b, "val", alignFloatValues(values, len(categories)))
	b.WriteString(`</c:ser>`)
}

func writeCartesianSeries(b *bytes.Buffer, name string, data []option.DataValue, categories []string, order int, marker bool, smooth bool, color string) {
	b.WriteString(`<c:ser>`)
	fmt.Fprintf(b, `<c:idx val="%d"/><c:order val="%d"/>`, order, order)
	writeSeriesText(b, name)
	writeSeriesShape(b, color, marker)
	if marker {
		b.WriteString(`<c:marker><c:symbol val="none"/></c:marker>`)
	}
	writeStringLiteral(b, "cat", alignCategories(categories, data))
	writeNumberLiteral(b, "val", valuesOf(data))
	if smooth {
		b.WriteString(`<c:smooth val="1"/>`)
	}
	b.WriteString(`</c:ser>`)
}

func writeSeriesShape(b *bytes.Buffer, color string, line bool) {
	if line {
		fmt.Fprintf(b, `<c:spPr><a:ln w="28575"><a:solidFill><a:srgbClr val="%s"/></a:solidFill></a:ln></c:spPr>`, color)
		return
	}
	fmt.Fprintf(b, `<c:spPr><a:solidFill><a:srgbClr val="%s"/></a:solidFill><a:ln><a:noFill/></a:ln></c:spPr>`, color)
}

func writeMarker(b *bytes.Buffer, color string, size float64) {
	if size <= 0 {
		size = 6
	}
	if size < 2 {
		size = 2
	}
	if size > 72 {
		size = 72
	}
	fmt.Fprintf(b, `<c:marker><c:symbol val="circle"/><c:size val="%s"/><c:spPr><a:solidFill><a:srgbClr val="%s"/></a:solidFill><a:ln><a:noFill/></a:ln></c:spPr></c:marker>`, strconv.FormatFloat(size, 'f', 0, 64), color)
}

func writeDataPointShape(b *bytes.Buffer, idx int, color string) {
	fmt.Fprintf(b, `<c:dPt><c:idx val="%d"/><c:spPr><a:solidFill><a:srgbClr val="%s"/></a:solidFill><a:ln><a:noFill/></a:ln></c:spPr></c:dPt>`, idx, color)
}

func writeSeriesText(b *bytes.Buffer, name string) {
	if name == "" {
		return
	}
	fmt.Fprintf(b, `<c:tx><c:v>%s</c:v></c:tx>`, escape(name))
}

func writeStringLiteral(b *bytes.Buffer, tag string, values []string) {
	fmt.Fprintf(b, `<c:%s><c:strLit><c:ptCount val="%d"/>`, tag, len(values))
	for i, v := range values {
		fmt.Fprintf(b, `<c:pt idx="%d"><c:v>%s</c:v></c:pt>`, i, escape(v))
	}
	fmt.Fprintf(b, `</c:strLit></c:%s>`, tag)
}

func writeNumberLiteral(b *bytes.Buffer, tag string, values []float64) {
	fmt.Fprintf(b, `<c:%s><c:numLit><c:formatCode>General</c:formatCode><c:ptCount val="%d"/>`, tag, len(values))
	for i, v := range values {
		fmt.Fprintf(b, `<c:pt idx="%d"><c:v>%s</c:v></c:pt>`, i, strconv.FormatFloat(v, 'f', -1, 64))
	}
	fmt.Fprintf(b, `</c:numLit></c:%s>`, tag)
}

func writeAxisRefs(b *bytes.Buffer) {
	fmt.Fprintf(b, `<c:axId val="%d"/><c:axId val="%d"/>`, catAxisID, valAxisID)
}

func writeCategoryAxis(b *bytes.Buffer) {
	writeCategoryAxisAt(b, "b")
}

func writeCategoryAxisAt(b *bytes.Buffer, pos string) {
	if pos == "" {
		pos = "b"
	}
	fmt.Fprintf(b, `<c:catAx><c:axId val="%d"/><c:scaling><c:orientation val="minMax"/></c:scaling>`, catAxisID)
	fmt.Fprintf(b, `<c:delete val="0"/><c:axPos val="%s"/>`, escape(pos))
	fmt.Fprintf(b, `<c:crossAx val="%d"/>`, valAxisID)
	b.WriteString(`<c:tickLblPos val="nextTo"/></c:catAx>`)
}

func writeValueAxis(b *bytes.Buffer) {
	writeValueAxisWithID(b, valAxisID, catAxisID, "l")
}

func writeValueAxisWithID(b *bytes.Buffer, axisID, crossAxisID int, pos string) {
	fmt.Fprintf(b, `<c:valAx><c:axId val="%d"/><c:scaling><c:orientation val="minMax"/></c:scaling>`, axisID)
	b.WriteString(`<c:delete val="0"/>`)
	if pos != "" {
		fmt.Fprintf(b, `<c:axPos val="%s"/>`, escape(pos))
	} else {
		b.WriteString(`<c:axPos val="l"/>`)
	}
	fmt.Fprintf(b, `<c:crossAx val="%d"/>`, crossAxisID)
	b.WriteString(`<c:crosses val="autoZero"/><c:tickLblPos val="nextTo"/></c:valAx>`)
}

func categoriesFor(opt *option.Option) []string {
	if len(opt.XAxis) > 0 && len(opt.XAxis[0].Data) > 0 {
		out := make([]string, len(opt.XAxis[0].Data))
		copy(out, opt.XAxis[0].Data)
		return out
	}
	for _, s := range opt.Series {
		switch v := s.(type) {
		case *option.LineSeries:
			return categoriesFromData(v.Data)
		case *option.BarSeries:
			return categoriesFromData(v.Data)
		}
	}
	return nil
}

// yCategoriesFor 取 Y 轴的 category 标签（横向条形图用）。
func yCategoriesFor(opt *option.Option) []string {
	if len(opt.YAxis) > 0 && len(opt.YAxis[0].Data) > 0 {
		out := make([]string, len(opt.YAxis[0].Data))
		copy(out, opt.YAxis[0].Data)
		return out
	}
	for _, s := range opt.Series {
		if v, ok := s.(*option.BarSeries); ok {
			return categoriesFromData(v.Data)
		}
	}
	return nil
}

func categoriesFromData(data []option.DataValue) []string {
	out := make([]string, len(data))
	for i, d := range data {
		switch {
		case d.Label != "":
			out[i] = d.Label
		case d.Name != "":
			out[i] = d.Name
		default:
			out[i] = strconv.Itoa(i + 1)
		}
	}
	return out
}

func pieCategories(data []option.DataValue) []string {
	return categoriesFromData(data)
}

func isDoughnutPie(s *option.PieSeries) bool {
	return s != nil && len(s.Radius) >= 2 && s.Radius[0].Value > 0
}

func doughnutHoleSize(s *option.PieSeries) int {
	if s == nil || len(s.Radius) < 2 {
		return 50
	}
	inner, outer := s.Radius[0].Value, s.Radius[1].Value
	if inner <= 0 || outer <= 0 {
		return 50
	}
	size := int(inner / outer * 100)
	if size < 10 {
		return 10
	}
	if size > 90 {
		return 90
	}
	return size
}

func alignCategories(categories []string, data []option.DataValue) []string {
	if len(categories) >= len(data) {
		return categories[:len(data)]
	}
	out := make([]string, len(data))
	copy(out, categories)
	for i := len(categories); i < len(data); i++ {
		d := data[i]
		switch {
		case d.Label != "":
			out[i] = d.Label
		case d.Name != "":
			out[i] = d.Name
		default:
			out[i] = strconv.Itoa(i + 1)
		}
	}
	return out
}

func valuesOf(data []option.DataValue) []float64 {
	out := make([]float64, len(data))
	for i, d := range data {
		out[i] = d.Number()
	}
	return out
}

func xValuesOf(data []option.DataValue) []float64 {
	out := make([]float64, len(data))
	for i, d := range data {
		x, _ := d.Pair()
		out[i] = x
	}
	return out
}

func yValuesOf(data []option.DataValue) []float64 {
	out := make([]float64, len(data))
	for i, d := range data {
		_, y := d.Pair()
		out[i] = y
	}
	return out
}

func radarCategories(series []*option.RadarSeries, radar *option.Radar) []string {
	if radar != nil && len(radar.Indicator) > 0 {
		out := make([]string, len(radar.Indicator))
		for i, item := range radar.Indicator {
			if item.Name != "" {
				out[i] = item.Name
			} else {
				out[i] = strconv.Itoa(i + 1)
			}
		}
		return out
	}
	maxLen := 0
	for _, s := range series {
		for _, item := range s.Data {
			if len(item.Value) > maxLen {
				maxLen = len(item.Value)
			}
		}
	}
	out := make([]string, maxLen)
	for i := range out {
		out[i] = strconv.Itoa(i + 1)
	}
	return out
}

func alignFloatValues(values []float64, n int) []float64 {
	if len(values) >= n {
		return values[:n]
	}
	out := make([]float64, n)
	copy(out, values)
	return out
}

func paletteFor(opt *option.Option) []string {
	out := make([]string, 0, len(opt.Color))
	for _, c := range opt.Color {
		if hex := normalizeHexColor(c); hex != "" {
			out = append(out, hex)
		}
	}
	if len(out) > 0 {
		return out
	}
	return defaultPalette
}

func paletteAt(palette []string, index int) string {
	if len(palette) == 0 {
		return defaultPalette[index%len(defaultPalette)]
	}
	return palette[((index%len(palette))+len(palette))%len(palette)]
}

func normalizeHexColor(c string) string {
	c = strings.TrimSpace(c)
	c = strings.TrimPrefix(c, "#")
	switch len(c) {
	case 3:
		var b strings.Builder
		for _, r := range c {
			b.WriteRune(r)
			b.WriteRune(r)
		}
		return strings.ToUpper(b.String())
	case 6:
		return strings.ToUpper(c)
	}
	return ""
}

func escape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
