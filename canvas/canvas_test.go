package canvas

import (
	"bytes"
	"testing"

	"github.com/zzhtl/zcharts/common/color"
)

func TestSmokeRender(t *testing.T) {
	for _, f := range []Format{FormatPNG, FormatSVG, FormatPDF} {
		t.Run(string(f), func(t *testing.T) {
			c := New(200, 100, nil)
			c.SetFill(color.MustParse("#5470c6"))
			c.SetStroke(color.MustParse("#333"))
			c.SetStrokeWidth(2)
			c.DrawRect(10, 10, 80, 80)

			c.SetFill(color.MustParse("#fac858"))
			c.NoStroke()
			c.DrawCircle(150, 50, 30)

			c.SetStroke(color.MustParse("#ee6666"))
			c.SetStrokeWidth(1)
			c.DrawLine(100, 10, 100, 90)

			c.DrawText(100, 95, "中文 zcharts", TextStyle{
				Size:   14,
				Color:  color.MustParse("#333"),
				Anchor: AnchorMiddle,
				VAlign: AlignBottom,
			})

			var buf bytes.Buffer
			if err := c.Write(&buf, f); err != nil {
				t.Fatalf("write %s: %v", f, err)
			}
			if buf.Len() == 0 {
				t.Fatalf("%s: empty output", f)
			}
		})
	}
}
