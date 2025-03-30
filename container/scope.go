package container

import (
	"fmt"
)

type Scope struct {
	*Container
	isGlobal bool
	deps     map[string]any
}

func unwrap[T any](v any) (*T, error) {
	if v == nil {
		return nil, fmt.Errorf("failed unwrap of type %T. Value is nil", *new(T))
	}
	castTPtr, okPtr := v.(*T)
	if !okPtr {
		return nil, fmt.Errorf("failed unwrap of type %T", *new(T))
	}
	return castTPtr, nil
}
func RequireServiceFor[T any](s *Scope) (*T, error) {
	nameDep := nameFor[T]()
	itemAny, ok := s.depsTree[nameDep]
	if !ok {
		return nil, fmt.Errorf("dependency %s not found", nameDep)
	}
	item, _ := itemAny.(itemInterface)
	dep, err := item.Init(s)
	if err != nil {
		return nil, err
	}
	return unwrap[T](dep)
}
