package theme

import (
	"fmt"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = map[string]func() *Theme{
		"default": Default,
		"light":   Default,
		"dark":    Dark,
	}
)

// Register 注册一个主题工厂函数。每次 Get 都会调用该函数生成新实例。
func Register(name string, factory func() *Theme) {
	registryMu.Lock()
	registry[name] = factory
	registryMu.Unlock()
}

// Get 按名称获取主题。未注册时返回 (nil, error)。
func Get(name string) (*Theme, error) {
	registryMu.RLock()
	f, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("zcharts/theme: unknown theme %q", name)
	}
	return f(), nil
}

// MustGet 同 Get，未注册时 panic。
func MustGet(name string) *Theme {
	t, err := Get(name)
	if err != nil {
		panic(err)
	}
	return t
}

// Names 返回当前已注册的主题名列表。
func Names() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}
