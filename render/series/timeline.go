package series

import (
	"strings"

	zcanvas "github.com/zzhtl/zcharts/canvas"
	"github.com/zzhtl/zcharts/common/color"
	"github.com/zzhtl/zcharts/common/geom"
	"github.com/zzhtl/zcharts/option"
)

// DrawTimelineArgs Timeline 渲染参数。
type DrawTimelineArgs struct {
	Canvas    zcanvas.Canvas
	Series    *option.TimelineSeries
	Bounds    geom.Rect
	Palette   color.Palette
	Family    string
	TextColor color.Color
	MutedText color.Color
	LineColor color.Color
}

// DrawTimeline 绘制从上到下的事件时间线。
func DrawTimeline(a DrawTimelineArgs) {
	if a.Series == nil || len(a.Series.Data) == 0 {
		return
	}
	rect := timelineRect(a.Bounds, a.Series)
	if rect.IsZero() {
		return
	}

	events := a.Series.Data
	lineX := timelineLineX(rect)
	contentX := lineX + 28
	contentW := rect.Right() - contentX
	if contentW <= 40 {
		return
	}

	ys := timelineEventPositions(rect, len(events))
	lineColor := timelineLineColor(a)
	lineWidth := a.Series.LineStyle.Width
	if lineWidth <= 0 {
		lineWidth = 2
	}
	if len(ys) > 1 {
		a.Canvas.SetStroke(lineColor)
		a.Canvas.SetStrokeWidth(lineWidth)
		a.Canvas.DrawLine(lineX, ys[0], lineX, ys[len(ys)-1])
	}

	rowH := rect.H / float64(len(events))
	for i, event := range events {
		y := ys[i]
		dotColor := timelineEventColor(a, event, i)
		a.Canvas.SetStroke(dotColor.WithAlpha(58))
		a.Canvas.SetStrokeWidth(6)
		a.Canvas.NoFill()
		a.Canvas.DrawCircle(lineX, y, 8)
		a.Canvas.SetFill(dotColor)
		a.Canvas.SetStroke(color.RGB(255, 255, 255))
		a.Canvas.SetStrokeWidth(2)
		a.Canvas.DrawCircle(lineX, y, 6)
		a.Canvas.SetStroke(dotColor.WithAlpha(110))
		a.Canvas.SetStrokeWidth(1)
		a.Canvas.DrawLine(lineX+10, y, contentX-8, y)

		drawTimelineEventText(a, event, y, rowH, lineX, contentX, contentW)
	}
}

func timelineRect(bounds geom.Rect, s *option.TimelineSeries) geom.Rect {
	x := bounds.X + bounds.W*0.08
	y := bounds.Y + bounds.H*0.06
	w := bounds.W * 0.84
	h := bounds.H * 0.86
	if s.Width.Set {
		w = s.Width.Resolve(bounds.W, w)
	}
	if s.Height.Set {
		h = s.Height.Resolve(bounds.H, h)
	}
	if s.Left.Set {
		if s.Left.Keyword == "center" {
			x = bounds.X + (bounds.W-w)/2
		} else {
			x = bounds.X + s.Left.Resolve(bounds.W, x-bounds.X)
		}
	}
	if s.Top.Set {
		if s.Top.Keyword == "middle" || s.Top.Keyword == "center" {
			y = bounds.Y + (bounds.H-h)/2
		} else {
			y = bounds.Y + s.Top.Resolve(bounds.H, y-bounds.Y)
		}
	}
	return geom.Rect{X: x, Y: y, W: w, H: h}
}

func timelineLineX(rect geom.Rect) float64 {
	timeW := rect.W * 0.22
	if timeW < 82 {
		timeW = 82
	}
	if timeW > 180 {
		timeW = 180
	}
	return rect.X + timeW
}

func timelineEventPositions(rect geom.Rect, n int) []float64 {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []float64{rect.CenterY()}
	}
	top := rect.Y + 22
	bottom := rect.Bottom() - 22
	if bottom <= top {
		top, bottom = rect.Y, rect.Bottom()
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = top + (bottom-top)*float64(i)/float64(n-1)
	}
	return out
}

func drawTimelineEventText(a DrawTimelineArgs, event option.TimelineEvent, y, rowH, lineX, contentX, contentW float64) {
	timeStyle := zcanvas.TextStyle{
		Family: pickTimelineFamily(a.Family),
		Size:   12,
		Color:  timelineMutedTextColor(a),
		Anchor: zcanvas.AnchorEnd,
		VAlign: zcanvas.AlignMiddle,
	}
	if event.Time != "" {
		// AlignMiddle 已让文字垂直居中于 y，与圆点中心同一水平线。
		a.Canvas.DrawText(lineX-22, y, event.Time, timeStyle)
	}

	title := firstTimelineText(event.Title, event.Name, event.Content)
	content := event.Content
	if title == event.Content {
		content = ""
	}
	titleStyle := zcanvas.TextStyle{
		Family: pickTimelineFamily(a.Family),
		Size:   15,
		Weight: "bold",
		Color:  timelineTextColor(a),
		Anchor: zcanvas.AnchorStart,
		VAlign: zcanvas.AlignMiddle,
	}
	contentStyle := zcanvas.TextStyle{
		Family: pickTimelineFamily(a.Family),
		Size:   timelineLabelSize(a),
		Color:  timelineMutedTextColor(a),
		Anchor: zcanvas.AnchorStart,
		VAlign: zcanvas.AlignMiddle,
	}

	titleLines := limitTimelineLines(a.Canvas, wrapTimelineText(a.Canvas, title, titleStyle, contentW), 2, titleStyle, contentW)
	maxContentLines := int((rowH - 28) / 16)
	if maxContentLines > 3 {
		maxContentLines = 3
	}
	contentLines := limitTimelineLines(a.Canvas, wrapTimelineText(a.Canvas, content, contentStyle, contentW), maxContentLines, contentStyle, contentW)

	titleLineH := 18.0
	contentLineH := 16.0
	blockH := float64(len(titleLines)) * titleLineH
	if len(contentLines) > 0 {
		blockH += 4 + float64(len(contentLines))*contentLineH
	}
	textY := y - blockH/2 + titleLineH/2
	for _, line := range titleLines {
		a.Canvas.DrawText(contentX, textY, line, titleStyle)
		textY += titleLineH
	}
	if len(contentLines) > 0 {
		textY += 4
		for _, line := range contentLines {
			a.Canvas.DrawText(contentX, textY, line, contentStyle)
			textY += contentLineH
		}
	}
}

func timelineLineColor(a DrawTimelineArgs) color.Color {
	out := a.LineColor
	if a.Series.LineStyle.Color != "" {
		if c, err := color.Parse(string(a.Series.LineStyle.Color)); err == nil {
			out = c
		}
	}
	if a.Series.LineStyle.Opacity != nil {
		out = out.WithAlpha(uint8(*a.Series.LineStyle.Opacity * 255))
	}
	return out
}

func timelineEventColor(a DrawTimelineArgs, event option.TimelineEvent, index int) color.Color {
	out := a.Palette.At(index)
	if a.Series.Color != "" {
		if c, err := color.Parse(string(a.Series.Color)); err == nil {
			out = c
		}
	}
	if a.Series.ItemStyle.Color != "" {
		if c, err := color.Parse(string(a.Series.ItemStyle.Color)); err == nil {
			out = c
		}
	}
	if event.ItemStyle.Color != "" {
		if c, err := color.Parse(string(event.ItemStyle.Color)); err == nil {
			out = c
		}
	}
	return out
}

func timelineTextColor(a DrawTimelineArgs) color.Color {
	out := a.TextColor
	if a.Series.Label.Color != "" {
		if c, err := color.Parse(string(a.Series.Label.Color)); err == nil {
			out = c
		}
	}
	return out
}

func timelineMutedTextColor(a DrawTimelineArgs) color.Color {
	if !a.MutedText.IsZero() {
		return a.MutedText
	}
	return timelineTextColor(a).WithAlpha(185)
}

func timelineLabelSize(a DrawTimelineArgs) float64 {
	if a.Series.Label.FontSize > 0 {
		return a.Series.Label.FontSize
	}
	return 12
}

func pickTimelineFamily(fallback string) string {
	if fallback != "" {
		return fallback
	}
	return "default"
}

func wrapTimelineText(c zcanvas.Canvas, s string, style zcanvas.TextStyle, maxWidth float64) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if maxWidth <= 0 {
		return []string{s}
	}
	var lines []string
	var cur []rune
	for _, r := range []rune(s) {
		candidate := string(append(cur, r))
		w, _, _ := c.MeasureText(candidate, style)
		if w <= maxWidth || len(cur) == 0 {
			cur = append(cur, r)
			continue
		}
		lines = append(lines, strings.TrimSpace(string(cur)))
		cur = []rune{r}
	}
	if len(cur) > 0 {
		lines = append(lines, strings.TrimSpace(string(cur)))
	}
	return lines
}

func limitTimelineLines(c zcanvas.Canvas, lines []string, maxLines int, style zcanvas.TextStyle, maxWidth float64) []string {
	if maxLines <= 0 || len(lines) == 0 {
		return nil
	}
	if len(lines) <= maxLines {
		return lines
	}
	out := append([]string(nil), lines[:maxLines]...)
	out[len(out)-1] = fitTimelineEllipsis(c, out[len(out)-1], style, maxWidth)
	return out
}

func fitTimelineEllipsis(c zcanvas.Canvas, s string, style zcanvas.TextStyle, maxWidth float64) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "..."
	}
	for {
		candidate := s + "..."
		w, _, _ := c.MeasureText(candidate, style)
		if w <= maxWidth || len([]rune(s)) == 1 {
			return candidate
		}
		rs := []rune(s)
		s = strings.TrimSpace(string(rs[:len(rs)-1]))
	}
}

func firstTimelineText(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
