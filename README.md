# zcharts

基于 Go 的图表生成库，兼容 Apache ECharts 风格的 JSON 配置，**不依赖浏览器**，纯 Go 编译运行。

支持 PNG / SVG / PDF 三种输出格式，并提供把图表直接嵌入 Word docx 文档的能力。

## 特性

- ✅ 兼容 ECharts 5.x JSON option，迁移成本极低
- ✅ 纯 Go 实现，无 CGO/无浏览器/无 headless Chrome 依赖
- ✅ 支持多种常用图表：**折线、柱状/堆积柱、饼/环、散点、雷达、热力、仪表盘、漏斗、词云、时间轴**
- ✅ 美观细节：柱状圆角+渐变、折线渐变面积/虚线、数据标签、饼图引导线、中文换行
- ✅ PNG / SVG / PDF 三种输出格式
- ✅ Word docx 撰写：**标题、富文本正文、有序/无序列表、富表格（合并单元格/底色/单元格内嵌图表）与图表自由组合**成完整文档
- ✅ 内置 default / dark 两套主题，与 ECharts 主题机制对齐
- ✅ 字体可定制：内置 fallback + 用户可注册中文字体覆盖
- ✅ 自定义图表扩展：未知 `series.type` 可通过 `chart.WithSeriesRenderer` 自行绘制
- ✅ 架构清晰：图表生成 与 Word 集成 物理隔离，公共能力下沉到 `common/`

## 安装

```bash
go get github.com/zzhtl/zcharts
```

## 快速开始：用 ECharts JSON 渲染一张折线图

```go
package main

import (
    "log"
    "os"

    "github.com/zzhtl/zcharts/chart"
)

const optionJSON = `{
  "title": {"text": "Weekly Sales"},
  "xAxis": {"type": "category", "data": ["Mon","Tue","Wed","Thu","Fri","Sat","Sun"]},
  "yAxis": {"type": "value"},
  "series": [{"name": "A", "type": "line", "smooth": true,
              "data": [120, 200, 150, 80, 70, 110, 130]}]
}`

func main() {
    f, _ := os.Create("line.png")
    defer f.Close()
    if err := chart.RenderFromJSON([]byte(optionJSON), chart.FormatPNG, f,
        chart.WithSize(800, 500)); err != nil {
        log.Fatal(err)
    }
}
```

## 图表示例

完整示例位于 `examples/`：

| 类型 | 目录 |
|---|---|
| 折线图 / 区域图 | `examples/line/` |
| 柱状图（多 series 并排） | `examples/bar/` |
| 堆积柱状图 | `examples/stacked_bar/` |
| 事件时间轴 | `examples/timeline/` |
| 折线 + 柱状图 | `examples/combo/` |
| 饼图 / 环图 | `examples/pie/` |
| 散点图 | `examples/scatter/` |
| 雷达图 | `examples/radar/` |
| 热力图 | `examples/heatmap/` |
| 仪表盘 | `examples/gauge/` |
| 漏斗图 | `examples/funnel/` |
| 词云图 | `examples/wordcloud/` |
| 自定义图表 | `examples/custom/` |
| 多格式输出（PNG/SVG/PDF） | `examples/multi_format/` |
| 暗黑主题 | `examples/theme_dark/` |
| 嵌入 Word docx（图片路径） | `examples/docx_insert/` |
| Word 原生图表 docx | `examples/docx_native/` |
| Word 完整报告（标题/正文/列表/表格/图表组合） | `examples/docx_report/` |
| 打开已有 Word 并按标题追加内容 | `examples/docx_edit/` |
| Word 全图表对照（每种图表「图片样式 vs Word 图表样式」并排） | `examples/docx_gallery/` |

## 自定义图表

当 JSON 中出现未内置的 `series.type` 时，解析器会保留为 `option.CustomSeries`。
渲染时通过 `chart.WithSeriesRenderer` 注册同名渲染器即可绘制任意自定义图形：

```go
chart.RenderFromJSON(data, chart.FormatPNG, w,
    chart.WithSeriesRenderer("metricCard", func(ctx *render.Context, s option.Series, index int) error {
        ctx.Canvas.DrawRect(40, 50, 240, 100)
        return nil
    }))
```

完整用法见 `examples/custom/`。

## 主题切换

```go
chart.Render(opt, chart.FormatPNG, w, chart.WithTheme("dark"))
```

内置主题：`default`、`light`、`dark`。

加载 ECharts 主题 JSON：

```go
import "github.com/zzhtl/zcharts/theme"

_ = theme.RegisterECharts("mytheme", echartsThemeJSON)
chart.Render(opt, chart.FormatPNG, w, chart.WithTheme("mytheme"))
```

## 中文字体

`zcharts` 内置一个 ASCII fallback 字体（来自 `golang.org/x/image/font/gofont/goregular`）。
中文场景下需要注册一个 CJK 字体：

```go
import "github.com/zzhtl/zcharts/font"

mgr := font.New()
_ = mgr.LoadFile("default", "/path/to/SourceHanSansSC-Regular.ttf")
chart.Render(opt, chart.FormatPNG, w, chart.WithFonts(mgr))
```

## 嵌入 Word docx

```go
import (
    "log"

    "github.com/zzhtl/zcharts/docx"
    "github.com/zzhtl/zcharts/jsonopt"
)

doc := docx.New()
doc.AddParagraph("月度业务报告")
doc.AddParagraph("以下为本月销售折线图：")

opt, err := jsonopt.ParseString(optionJSON)
if err != nil {
    log.Fatal(err)
}
if err := doc.AddChart(opt, docx.AsImage(docx.PNG, 720, 400)); err != nil {
    log.Fatal(err)
}

if err := doc.Save("assets/tmp/report.docx"); err != nil {
    log.Fatal(err)
}
```

`docx.AsImage` 把图表渲染为 PNG 后嵌入到 docx 段落里。
`docx.AsNativeChart(width, height)` 优先生成 Word 原生 chart XML：折线图、柱状图、堆积柱状图、**面积图/堆积面积图、横向条形图**、饼/环图、散点图和雷达图走标准 chart XML（可在 Word 中继续编辑）；热力图、仪表盘、漏斗图、时间轴和词云走 Word/WPS 形状绘制；仍无法稳定表达的图表会自动退回 PNG 嵌入。

### 撰写完整 Word 文档

除了图表，`docx.Document` 还支持标题、富文本正文、列表和富表格，可与图表自由组合：

```go
doc := docx.New()
doc.AddTitle("年度业务报告")                 // 文档大标题
doc.AddHeading("一、整体概况", 1)             // 1~6 级标题
doc.AddRichParagraph(docx.RichParagraph{    // 富文本：加粗、颜色、字号、对齐
    Runs: []docx.Run{
        {Text: "营收实现 "},
        {Text: "稳健增长", Style: docx.RunStyle{Bold: true, Color: "#C00000"}},
    },
})
doc.AddBulletList("线上放量", "区域整合")      // 无序列表
doc.AddOrderedList("第一步", "第二步")         // 有序列表

doc.AddTable(docx.Table{                     // 富表格
    Rows: []docx.Row{
        {Header: true, Cells: []docx.Cell{
            {Text: "指标", Shading: "#4472C4", Style: docx.RunStyle{Color: "#FFFFFF", Bold: true}, GridSpan: 2},
        }},
        {Cells: []docx.Cell{
            {Text: "营收"},
            {Chart: &docx.CellChart{Option: opt, Insert: docx.AsImage(docx.PNG, 180, 90)}}, // 单元格内嵌图表
        }},
    },
})
_ = doc.Save("report.docx")
```

表格支持合并单元格（`GridSpan` 跨列、`VMerge` 跨行）、单元格底色、对齐、表头重复以及单元格内富文本/嵌图。完整用法见 `examples/docx_report/`。

## 架构

```
chart/         对外入口：Render / RenderFromJSON
option/        ECharts 配置数据模型（Go struct）
jsonopt/       ECharts JSON → option.Option
theme/         主题（default / dark / ECharts JSON 兼容）
font/          字体管理（内置 fallback + 用户可注册）
canvas/        绘制抽象层（封装 tdewolff/canvas，输出 PNG/SVG/PDF）
render/        渲染引擎
  ├── coord/   坐标系（直角 / 极坐标）
  ├── scale/   线性 / 分类 / 时间刻度
  ├── layout/  标题 / 图例 / 网格 / 坐标轴
  └── series/  各 series 类型渲染器
common/        通用工具：color / geom / number / text / errs
docx/          Word docx 集成（与 chart 解耦的独立子领域）
  ├── ooxml/   docx zip 包结构 + relationships
  └── chart/   Word 原生 chart XML + VML 形状绘制
```

依赖方向严格单向：`common ← font/theme/option ← canvas ← render ← chart ← docx`。

## 当前阶段不支持

- 动画 / 交互 / tooltip 实时显示（静态图表场景）
- ECharts 扩展生态（GL / Map / Tree / Sankey 等高级图形）
- Word 原生 chart XML 目前覆盖 line / bar / stacked bar / area / 横向 bar / pie / doughnut / scatter / radar；heatmap / gauge / funnel / timeline / wordCloud 通过 Word/WPS 形状绘制近似表达（形状绘制在新版 Word/WPS 兼容性较好，部分老旧渲染器对 VML 支持有限）
- ECharts 所有高级布局能力的完全等价实现

## 运行示例

```bash
go run ./examples/line        # → assets/tmp/line.png
go run ./examples/bar         # → assets/tmp/bar.png
go run ./examples/stacked_bar # → assets/tmp/stacked_bar.png
go run ./examples/timeline    # → assets/tmp/timeline.png
go run ./examples/combo       # → assets/tmp/combo.png
go run ./examples/pie         # → assets/tmp/pie.png
go run ./examples/scatter     # → assets/tmp/scatter.png
go run ./examples/radar       # → assets/tmp/radar.png
go run ./examples/heatmap     # → assets/tmp/heatmap.png
go run ./examples/gauge       # → assets/tmp/gauge.png
go run ./examples/funnel      # → assets/tmp/funnel.png
go run ./examples/wordcloud   # → assets/tmp/wordcloud.png
go run ./examples/custom      # → assets/tmp/custom.png
go run ./examples/multi_format # → assets/tmp/multi.{png,svg,pdf}
go run ./examples/theme_dark   # → assets/tmp/dark.png
go run ./examples/docx_insert  # → assets/tmp/report.docx
go run ./examples/docx_native  # → assets/tmp/native_charts.docx
go run ./examples/docx_report  # → assets/tmp/full_report.docx（标题/正文/列表/表格/图表组合）
go run ./examples/docx_edit # → assets/tmp/edit_updated.docx（打开已有 Word 并按标题追加内容）
go run ./examples/docx_gallery # → assets/tmp/gallery.docx（所有图表：图片样式 vs Word 图表样式 并排对照）
```

## 许可证

Apache-2.0
