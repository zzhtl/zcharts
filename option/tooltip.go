package option

// Tooltip 第一阶段仅保留配置，不参与渲染（静态图无交互）。
type Tooltip struct {
	Show    *bool  `json:"show,omitempty"`
	Trigger string `json:"trigger,omitempty"`
}
