// Package font 管理 zcharts 渲染所需的字体资源。
//
// 设计目标：
//   - 内置一个体积小的 fallback 字体（goregular），保证开箱即用；
//   - 用户可通过 Load/LoadFile 注册自定义字体（如思源黑体），覆盖默认；
//   - 通过 Face(family, size) 获取 golang.org/x/image/font.Face，供渲染使用。
//
// 注意：goregular 仅含 ASCII 字符。若图表中包含中文等非 ASCII，必须由用户注册中文字体。
package font

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/zzhtl/zcharts/common/errs"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

// Entry 保存一个字体的解析对象与原始字节。
// 原始字节供 tdewolff/canvas 等下游库重新加载使用，避免重复 IO。
type Entry struct {
	Font *opentype.Font
	Data []byte
}

// Manager 管理多个 family → Entry 的映射。
type Manager struct {
	mu       sync.RWMutex
	fonts    map[string]*Entry
	fallback *Entry
}

// DefaultFamily 是 fallback 字体的逻辑名。
const DefaultFamily = "default"

// New 创建带有内置 fallback 字体的 Manager。
func New() *Manager {
	data := append([]byte(nil), goregular.TTF...)
	f, err := opentype.Parse(data)
	if err != nil {
		// goregular.TTF 由 Go 团队维护，几乎不可能解析失败；真发生时直接 panic。
		panic(fmt.Errorf("zcharts/font: parse goregular: %w", err))
	}
	entry := &Entry{Font: f, Data: data}
	m := &Manager{
		fonts:    map[string]*Entry{DefaultFamily: entry},
		fallback: entry,
	}
	return m
}

// Load 从字节切片加载字体并注册到 family。
func (m *Manager) Load(family string, data []byte) error {
	if family == "" {
		return errors.New("zcharts/font: empty family")
	}
	f, err := opentype.Parse(data)
	if err != nil {
		return fmt.Errorf("zcharts/font: parse %q: %w", family, err)
	}
	cp := append([]byte(nil), data...)
	m.mu.Lock()
	m.fonts[family] = &Entry{Font: f, Data: cp}
	m.mu.Unlock()
	return nil
}

// LoadFile 从文件加载字体。
func (m *Manager) LoadFile(family, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("zcharts/font: read %s: %w", path, err)
	}
	return m.Load(family, data)
}

// Face 返回指定 family 在 size 下的 font.Face。
// family 为空或未找到时返回 fallback 字体的 Face。
func (m *Manager) Face(family string, size float64) (font.Face, error) {
	e := m.lookup(family)
	if e == nil {
		return nil, errs.ErrFontUnavailable
	}
	face, err := opentype.NewFace(e.Font, &opentype.FaceOptions{
		Size:    size,
		DPI:     72, // 像素 = pt 时使用 72
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("zcharts/font: new face: %w", err)
	}
	return face, nil
}

// Lookup 返回 family 对应的 Entry（含原始字节与 *opentype.Font）。
// family 未找到时返回 fallback。
func (m *Manager) Lookup(family string) *Entry { return m.lookup(family) }

// MustFace 是 Face 的 panic 版本，失败时直接 panic。
func (m *Manager) MustFace(family string, size float64) font.Face {
	face, err := m.Face(family, size)
	if err != nil {
		panic(err)
	}
	return face
}

// Has 报告 family 是否已注册。
func (m *Manager) Has(family string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.fonts[family]
	return ok
}

// Families 返回当前已注册的 family 名称列表（含 default）。
func (m *Manager) Families() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.fonts))
	for k := range m.fonts {
		out = append(out, k)
	}
	return out
}

func (m *Manager) lookup(family string) *Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if family != "" {
		if e, ok := m.fonts[family]; ok {
			return e
		}
	}
	return m.fallback
}
