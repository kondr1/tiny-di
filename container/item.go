package container

import (
	"fmt"
	"reflect"
	"tiny-di/utility"
)

type itemInterface interface {
	Life() Lifetime
	Deps() []string
	Ints() []string
	Name() string
	TypeOf() reflect.Type
	ActivatorAny() any
	Init(c *Scope) (any, error)
}

type itemFor[T any] struct {
	NameType     string
	Interfaces   []string
	Dependencies []string
	Lifetime     Lifetime
	Type         reflect.Type
	Activator    func() *T
}

func activatorFor[T any]() *T              { return new(T) }
func (i *itemFor[T]) Life() Lifetime       { return i.Lifetime }
func (i *itemFor[T]) Deps() []string       { return i.Dependencies }
func (i *itemFor[T]) Ints() []string       { return i.Interfaces }
func (i *itemFor[T]) Name() string         { return i.NameType }
func (i *itemFor[T]) TypeOf() reflect.Type { return i.Type }
func (i *itemFor[T]) ActivatorAny() any    { return i.Activator() }
func (i *itemFor[T]) Init(s *Scope) (any, error) {
	var ok bool
	var resolved any
	var constructorWasInvoked bool
	switch i.Lifetime {
	case HostedService:
	case Singleton:
		resolved, ok = s.global.deps[i.NameType]
		if ok && !utility.IsNilOrDefault(resolved) {
			constructorWasInvoked = true
			break
		}
		resolved = i.Activator()
	case Scoped:
		resolved, ok = s.deps[i.NameType]
		if s.isGlobal {
			return nil, fmt.Errorf("dependency %s is Scope dependency. You should BuildScope at first", i.NameType)
		}
		if ok && !utility.IsNilOrDefault(resolved) {
			constructorWasInvoked = true
			break
		}
		resolved = i.Activator()
	case Transient:
		resolved = i.Activator()
		constructorWasInvoked = false
	}

	if constructorWasInvoked {
		return unwrap[T](resolved)
	}

	args := make([]reflect.Value, 0)
	for _, v := range i.Dependencies {
		itemVal, ok := s.depsTree[v[1:]]
		if !ok {
			return nil, fmt.Errorf("dependency %s not found", v)
		}
		item, _ := itemVal.(itemInterface)
		dep, err := item.Init(s)
		if err != nil {
			return nil, err
		}
		args = append(args, reflect.ValueOf(dep))
	}

	depValue := reflect.ValueOf(resolved)
	depValue.MethodByName("Init").Call(args)
	if i.Lifetime == Scoped {
		s.deps[i.NameType] = resolved
	}
	if i.Lifetime == Singleton {
		s.global.deps[i.NameType] = resolved
	}
	return unwrap[T](resolved)
}
