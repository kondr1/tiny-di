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
	Activator    func() (T, error)
}

func (i *itemFor[T]) Life() Lifetime             { return i.Lifetime }
func (i *itemFor[T]) Deps() []string             { return i.Dependencies }
func (i *itemFor[T]) Name() string               { return i.NameType }
func (i *itemFor[T]) TypeOf() reflect.Type       { return i.Type }
func (i *itemFor[T]) ActivatorAny() (any, error) { return i.Activator() }
func (i *itemFor[T]) Init(c *Scope) (any, error) {
	var ok bool
	var dep any
	var err error
	var initWasInvoked bool
	switch i.Lifetime {
	case Singleton:
		dep, ok = c.global.deps[i.NameType]
		if ok && !utility.IsNilOrDefault(dep) {
			initWasInvoked = true
			break
		}
		dep, err = i.Activator()
	case Scoped:
		dep, ok = c.deps[i.NameType]
		if c.id == 1 {
			return *new(T), fmt.Errorf("dependency %s is Scope dependency. You should BuildScope at first", i.NameType)
		}
		if ok && !utility.IsNilOrDefault(dep) {
			initWasInvoked = true
			break
		}
		dep, err = i.Activator()
	case Transient:
		dep, err = i.Activator()
		initWasInvoked = false
	case HostedService:
		dep, ok = c.global.deps[i.NameType]
		if ok && !utility.IsNilOrDefault(dep) {
			initWasInvoked = true
			break
		}
		dep, err = i.Activator()
	}

	if err != nil {
		return *new(T), err
	}
	if initWasInvoked {
		return unwrap[T](dep)
	}

	args := make([]reflect.Value, 0)
	for _, v := range i.Dependencies {
		itemVal, ok := c.depsTree[v]
		if !ok {
			return *new(T), fmt.Errorf("dependency %s not found", v)
		}
		item, _ := itemVal.(itemInterface)
		dep, err := item.Init(c)
		if err != nil {
			return *new(T), err
		}
		args = append(args, reflect.ValueOf(dep))
	}

	depValue := reflect.ValueOf(dep)
	depValue.MethodByName("Init").Call(args)
	if i.Lifetime == Scoped {
		c.deps[i.NameType] = dep
	}
	if i.Lifetime == Singleton {
		c.global.deps[i.NameType] = dep
	}
	return unwrap[T](dep)
}
