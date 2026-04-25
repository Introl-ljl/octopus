package transformer

import (
	"sync"

	"github.com/bestruirui/octopus/internal/transformer/model"
)

// Registry 管理 Format 到 Inbound/Outbound 适配器工厂的映射。
// 参考 CLIProxyAPI 的 translator.Registry 设计，使用字符串 Format 作为 key。
type Registry struct {
	mu       sync.RWMutex
	inbound  map[Format]func() model.Inbound
	outbound map[Format]func() model.Outbound
}

// NewRegistry 构造一个空的注册表。
func NewRegistry() *Registry {
	return &Registry{
		inbound:  make(map[Format]func() model.Inbound),
		outbound: make(map[Format]func() model.Outbound),
	}
}

// RegisterInbound 注册一个入站适配器工厂。
func (r *Registry) RegisterInbound(format Format, factory func() model.Inbound) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inbound[format] = factory
}

// RegisterOutbound 注册一个出站适配器工厂。
func (r *Registry) RegisterOutbound(format Format, factory func() model.Outbound) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outbound[format] = factory
}

// GetInbound 获取指定格式的入站适配器实例。
// 每次调用返回一个新实例（因为适配器可能保持流式状态）。
func (r *Registry) GetInbound(format Format) model.Inbound {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if factory, ok := r.inbound[format]; ok {
		return factory()
	}
	return nil
}

// GetOutbound 获取指定格式的出站适配器实例。
func (r *Registry) GetOutbound(format Format) model.Outbound {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if factory, ok := r.outbound[format]; ok {
		return factory()
	}
	return nil
}

// NeedTransform 判断入站和出站格式是否需要进行协议转换。
// 当两者相同时返回 false，表示可以使用透明代理直接透传。
func (r *Registry) NeedTransform(inFormat, outFormat Format) bool {
	return inFormat != outFormat
}

// HasInbound 检查是否存在指定格式的入站适配器。
func (r *Registry) HasInbound(format Format) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.inbound[format]
	return ok
}

// HasOutbound 检查是否存在指定格式的出站适配器。
func (r *Registry) HasOutbound(format Format) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.outbound[format]
	return ok
}

// defaultRegistry 是全局默认注册表实例。
var defaultRegistry = NewRegistry()

// Default 返回全局默认注册表。
func Default() *Registry {
	return defaultRegistry
}
