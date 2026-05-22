package geom

// Rect 表示轴对齐的矩形，左上角 (X,Y) + 宽高。
// 约定屏幕坐标系：Y 向下递增。
type Rect struct {
	X, Y, W, H float64
}

// Right 返回右边界 X 坐标。
func (r Rect) Right() float64 { return r.X + r.W }

// Bottom 返回下边界 Y 坐标。
func (r Rect) Bottom() float64 { return r.Y + r.H }

// CenterX / CenterY 返回中心坐标。
func (r Rect) CenterX() float64 { return r.X + r.W/2 }
func (r Rect) CenterY() float64 { return r.Y + r.H/2 }

// Center 返回中心点。
func (r Rect) Center() Point { return Point{r.CenterX(), r.CenterY()} }

// Inset 收缩四周（正值为向内收缩）。
func (r Rect) Inset(top, right, bottom, left float64) Rect {
	return Rect{
		X: r.X + left,
		Y: r.Y + top,
		W: r.W - left - right,
		H: r.H - top - bottom,
	}
}

// InsetUniform 四周等量收缩。
func (r Rect) InsetUniform(v float64) Rect { return r.Inset(v, v, v, v) }

// Contains 测试点是否落在矩形内（含边界）。
func (r Rect) Contains(p Point) bool {
	return p.X >= r.X && p.X <= r.Right() && p.Y >= r.Y && p.Y <= r.Bottom()
}

// IsZero 表示矩形无效（宽或高 <= 0）。
func (r Rect) IsZero() bool { return r.W <= 0 || r.H <= 0 }
