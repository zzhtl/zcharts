package geom

// PathCmd 描述一个路径指令。
type PathCmd struct {
	Op   PathOp
	X, Y float64
	// 控制点（CubicTo 用 X1Y1/X2Y2，QuadTo 用 X1Y1）。
	X1, Y1, X2, Y2 float64
}

// PathOp 路径操作枚举。
type PathOp uint8

const (
	OpMoveTo PathOp = iota
	OpLineTo
	OpQuadTo
	OpCubicTo
	OpClose
)

// Path 是 PathCmd 的有序列表，描述一条复合路径。
type Path []PathCmd

func (p *Path) MoveTo(x, y float64)   { *p = append(*p, PathCmd{Op: OpMoveTo, X: x, Y: y}) }
func (p *Path) LineTo(x, y float64)   { *p = append(*p, PathCmd{Op: OpLineTo, X: x, Y: y}) }
func (p *Path) Close()                { *p = append(*p, PathCmd{Op: OpClose}) }
func (p *Path) QuadTo(x1, y1, x, y float64) {
	*p = append(*p, PathCmd{Op: OpQuadTo, X1: x1, Y1: y1, X: x, Y: y})
}
func (p *Path) CubicTo(x1, y1, x2, y2, x, y float64) {
	*p = append(*p, PathCmd{Op: OpCubicTo, X1: x1, Y1: y1, X2: x2, Y2: y2, X: x, Y: y})
}
