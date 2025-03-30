package container

import (
	"fmt"
)

type Scope struct {
	*Container
	id   int
	deps map[string]any
}

func (c *Scope) DisposeScope() { c.scopeCounter = c.scopeCounter - 1 }
func unwrap[T any](v any) (T, error) {
	if v == nil {
		return *new(T), fmt.Errorf("failed unwrap of type %T. Value is nil", *new(T))
	}
	castTDep, okDep := v.(T)
	castTPtr, okPtr := v.(*T)
	if !okPtr && !okDep { // WTF?
		return *new(T), fmt.Errorf("failed unwrap of type %T", *new(T))
	}
	if okPtr {
		return *castTPtr, nil
	}
	return castTDep, nil
}
func RequireServiceFor[T any](s *Scope) (T, error) {
	nameDep := nameFor[T]()
	itemAny, ok := s.depsTree[nameDep]
	if !ok {
		return *new(T), fmt.Errorf("dependency %s not found", nameDep)
	}
	item, _ := itemAny.(itemInterface)
	dep, err := item.Init(s)
	if err != nil {
		return *new(T), err
	}
	return unwrap[T](dep)
}
