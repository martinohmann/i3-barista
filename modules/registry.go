package modules

import (
	"barista.run/bar"
)

type Registry struct {
	modules []bar.Module
	err     error
}

func NewRegistry() *Registry {
	return &Registry{
		modules: make([]bar.Module, 0),
	}
}

func (r *Registry) Add(modules ...bar.Module) *Registry {
	if r.err != nil {
		return r
	}

	for _, module := range modules {
		if module != nil {
			r.modules = append(r.modules, module)
		}
	}
	return r
}

func (r *Registry) Addf(factory func() (bar.Module, error)) *Registry {
	if r.err != nil {
		return r
	}

	var module bar.Module

	module, r.err = factory()

	return r.Add(module)
}

func (r *Registry) Err() error {
	return r.err
}

func (r *Registry) Modules() []bar.Module {
	return r.modules
}
