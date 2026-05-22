# zcharts

基于 Go 的图表生成库，兼容 Apache ECharts 风格的 JSON 配置，**不依赖浏览器**，纯 Go 编译运行。

支持 PNG / SVG / PDF 三种输出格式，并提供把图表直接嵌入 Word docx 文档的能力。

## 特性

- ✅ 兼容 ECharts 5.x JSON option，迁移成本极低
- ✅ 纯 Go 实现，无 CGO/无浏览器/无 headless Chrome 依赖
- ✅ 一阶段支持 7 种核心图表：**折线、柱状、饼/环、散点、雷达、热力、仪表盘**
- ✅ PNG / SVG / PDF 三种输出格式
- ✅ Word docx 集成：图表可直接插入到 docx 段落中（中文段落、多图混排）
- ✅ 内置 default / dark 两套主题，与 ECharts 主题机制对齐
- ✅ 字体可定制：内置 fallback + 用户可注册中文字体覆盖
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

## 7 种核心图表

完整示例位于 `examples/`：

| 类型 | 目录 |
|---|---|
| 折线图 / 区域图 | `examples/line/` |
| 柱状图（多 series 并排） | `examples/bar/` |
| 饼图 / 环图 | `examples/pie/` |
| 散点图 | `examples/scatter/` |
| 雷达图 | `examples/radar/` |
| 热力图 | `examples/heatmap/` |
| 仪表盘 | `examples/gauge/` |
| 多格式输出（PNG/SVG/PDF） | `examples/multi_format/` |
| 暗黑主题 | `examples/theme_dark/` |
| 嵌入 Word docx | `examples/docx_insert/` |

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
    "github.com/zzhtl/zcharts/docx"
    "github.com/zzhtl/zcharts/jsonopt"
)

doc := docx.New()
doc.AddParagraph("月度业务报告")
doc.AddParagraph("以下为本月销售折线图：")

opt, _ := jsonopt.ParseString(optionJSON)
_ = doc.AddChart(opt, docx.AsImage(docx.PNG, 720, 400))

_ = doc.Save("report.docx")
```

`docx.AsImage` 把图表渲染为 PNG 后嵌入到 docx 段落里。
`docx.AsNativeChart()` 是预留的"OOXML 原生 chart XML"路径，第一阶段未实现，调用时返回 `errs.ErrNotImplemented`。

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
  ├── scale/   线性 / 分类 / 时间 / 对数刻度
  ├── layout/  标题 / 图例 / 网格 / 坐标轴
  └── series/  各 series 类型渲染器
common/        通用工具：color / geom / number / text / errs
docx/          Word docx 集成（与 chart 解耦的独立子领域）
  ├── ooxml/   docx zip 包结构 + relationships
  └── chart/   原生 chart XML 占位（未实现）
```

依赖方向严格单向：`common ← font/theme/option ← canvas ← render ← chart ← docx`。

## 当前阶段不支持

- 动画 / 交互 / tooltip 实时显示（静态图表场景）
- ECharts 扩展生态（GL / Map / Tree / Sankey 等高级图形）
- Word docx 中的"原生 chart XML"路径（仅留接口）
- stack 堆叠 series（已规划，未实现）

## 运行示例

```bash
go run ./examples/line       # → line.png
go run ./examples/bar        # → bar.png
go run ./examples/pie        # → pie.png
go run ./examples/scatter    # → scatter.png
go run ./examples/radar      # → radar.png
go run ./examples/heatmap    # → heatmap.png
go run ./examples/gauge      # → gauge.png
go run ./examples/multi_format    # → multi.{png,svg,pdf}
go run ./examples/theme_dark      # → dark.png
go run ./examples/docx_insert     # → report.docx
```

## 许可证

Apache-2.0
