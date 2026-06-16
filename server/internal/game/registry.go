package game

// Registry — soha kaliti (slug) bo'yicha provider tanlaydi.
// Topilmasa fallback (masalan Sample) qaytadi.
type Registry struct {
	byKey    map[string]Provider
	fallback Provider
}

func NewRegistry(fallback Provider) *Registry {
	return &Registry{byKey: make(map[string]Provider), fallback: fallback}
}

func (r *Registry) Register(key string, p Provider) { r.byKey[key] = p }

func (r *Registry) For(key string) Provider {
	if p, ok := r.byKey[key]; ok {
		return p
	}
	return r.fallback
}
