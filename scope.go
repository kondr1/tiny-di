package container

import (
	"fmt"
)

type Scope struct {
	*Container
	isGlobal  bool
	instances map[string]any
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
	if !s.builded {
		panic("container not builded. You should call Build() before RequireServiceFor")
	}
	nameDep := nameFor[T]()
	if nameDep == "" || nameDep == "<nil>" {
		panic("Cannt extract dependency name. Maybe you should use RequireServiceForI for interfaces?")
	}
	item, ok := s.callSites[nameDep]
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrDependencyNotFound, nameDep))
	}
	dep, err := item.Build(s)
	if err != nil {
		return nil, err
	}
	return unwrap[T](dep)
}
func RequireServiceForI[I any](s *Scope) (I, error) {
	if !s.builded {
		panic("container not builded. You should call Build() before RequireServiceForI")
	}
	nameDep := nameForI[I]()
	if nameDep == "" || nameDep == "<nil>" {
		panic("Cannt extract dependency name. Maybe you should use RequireServiceFor for struct?")
	}
	item, ok := s.callSites[nameDep]
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrDependencyNotFound, nameDep))
	}
	dep, err := item.Build(s)
	if err != nil {
		return *new(I), err
	}
	return unwrapI[I](dep)
}
