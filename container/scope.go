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
	castTPtr, ok := v.(*T)
	if !ok {
		return nil, fmt.Errorf("failed unwrap of type %T", *new(T))
	}
	return castTPtr, nil
}

func unwrapI[I any](v any) (I, error) {
	if v == nil {
		return *new(I), fmt.Errorf("failed unwrap of type %T. Value is nil", *new(I))
	}
	castInterfacePtr, ok := v.(I)
	if !ok {
		return *new(I), fmt.Errorf("failed unwrap of type %T", *new(I))
	}
	return castInterfacePtr, nil
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
func RequireServiceForI[I any](c *Scope) (I, error) {
	nameDep := nameFor[I]()
	itemAny, ok := c.depsTree[nameDep]
	if !ok {
		return *new(I), fmt.Errorf("dependency %s not found", nameDep)
	}
	item, _ := itemAny.(itemInterface)
	dep, err := item.Init(c)
	if err != nil {
		return *new(I), err
	}
	return unwrapI[I](dep)
}
