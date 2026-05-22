// Package errs 集中定义 zcharts 各层共享的错误类型。
package errs

import "errors"

var (
	// ErrInvalidOption 表示传入的 option 不合法或缺少必需字段。
	ErrInvalidOption = errors.New("zcharts: invalid option")
	// ErrUnsupportedSeries 表示该 series 类型当前未实现。
	ErrUnsupportedSeries = errors.New("zcharts: unsupported series type")
	// ErrUnsupportedFormat 表示请求的输出格式未实现。
	ErrUnsupportedFormat = errors.New("zcharts: unsupported output format")
	// ErrNotImplemented 表示该路径目前未实现（占位）。
	ErrNotImplemented = errors.New("zcharts: not implemented")
	// ErrFontUnavailable 表示找不到可用字体。
	ErrFontUnavailable = errors.New("zcharts: font unavailable")
)
