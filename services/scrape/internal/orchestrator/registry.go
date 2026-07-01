package orchestrator

// Registry maps platform IDs to their adapters for easy lookup.
type Registry struct {
	adapters map[int16]PlatformAdapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[int16]PlatformAdapter)}
}

func (r *Registry) Register(a PlatformAdapter) {
	r.adapters[a.PlatformID()] = a
}

func (r *Registry) Get(platformID int16) (PlatformAdapter, bool) {
	a, ok := r.adapters[platformID]
	return a, ok
}

func (r *Registry) All() map[int16]PlatformAdapter {
	return r.adapters
}
