package container

import (
	"fmt"
	"reflect"
	"tiny-di/utility"
)

type itemInterface interface {
	Life() Lifetime
	Deps() []string
	Name() string
	TypeOf() reflect.Type
	ActivatorAny() (any, error)
	Init(c *Scope) (any, error)
}

type itemFor[T any] struct {
	NameType     string
	Dependencies []string
	Lifetime     Lifetime
	Type         reflect.Type
	Activator    func() (*T, error)
}

func activatorFor[T any]() (*T, error)           { return new(T), nil }
func (i *itemFor[T]) Life() Lifetime             { return i.Lifetime }
func (i *itemFor[T]) Deps() []string             { return i.Dependencies }
func (i *itemFor[T]) Name() string               { return i.NameType }
func (i *itemFor[T]) TypeOf() reflect.Type       { return i.Type }
func (i *itemFor[T]) ActivatorAny() (any, error) { return i.Activator() }
func (i *itemFor[T]) Init(s *Scope) (any, error) {
	var ok bool
	var dep any
	var err error
	var initWasInvoked bool
	switch i.Lifetime {
	case HostedService:
	case Singleton:
		dep, ok = s.global.deps[i.NameType]
		if ok && !utility.IsNilOrDefault(dep) {
			initWasInvoked = true
			break
		}
		dep, err = i.Activator()
	case Scoped:
		dep, ok = s.deps[i.NameType]
		if s.isGlobal {
			return new(T), fmt.Errorf("dependency %s is Scope dependency. You should BuildScope at first", i.NameType)
		}
		if ok && !utility.IsNilOrDefault(dep) {
			initWasInvoked = true
			break
		}
		dep, err = i.Activator()
	case Transient:
		dep, err = i.Activator()
		initWasInvoked = false
	}

	if err != nil {
		return new(T), err
	}
	if initWasInvoked {
		return unwrap[T](dep)
	}

	args := make([]reflect.Value, 0)
	for _, v := range i.Dependencies {
		itemVal, ok := s.depsTree[v[1:]]
		if !ok {
			return new(T), fmt.Errorf("dependency %s not found", v)
		}
		item, _ := itemVal.(itemInterface)
		dep, err := item.Init(s)
		if err != nil {
			return new(T), err
		}
		args = append(args, reflect.ValueOf(dep))
	}

	depValue := reflect.ValueOf(dep)
	depValue.MethodByName("Init").Call(args)
	if i.Lifetime == Scoped {
		s.deps[i.NameType] = dep
	}
	if i.Lifetime == Singleton {
		s.global.deps[i.NameType] = dep
	}
	return unwrap[T](dep)
}
