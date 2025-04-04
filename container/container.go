package container

import (
	"fmt"
	"reflect"
	"slices"
)

type Lifetime int

const (
	Singleton Lifetime = iota
	Transient
	Scoped
	HostedService
)

type Container struct {
	depsTree map[string]any // map[string]item[T]
	global   *Scope
}

func AddHostedService[T any](c *Container) { add[T](c, HostedService) }

// Transient lifetime services are created each time they're requested from the service container
func AddTransient[T any](c *Container) { add[T](c, Transient) }

// Singleton lifetime services are created the first time they're requested
func AddSingleton[T any](c *Container) { add[T](c, Singleton) }

// A scoped lifetime indicates that services are created once per client request (connection).
func AddScoped[T any](c *Container) { add[T](c, Scoped) }

func nameFor[T any]() string {
	var val T
	return fmt.Sprintf("%T", val)
}

func (c *Container) createScope() *Scope {
	return &Scope{
		Container: c,
		deps:      make(map[string]any),
	}
}
func add[T any](c *Container, lifetime Lifetime) {
	name := nameFor[T]()
	dep := &itemFor[T]{
		NameType:     name,
		Dependencies: []string{},
		Lifetime:     lifetime,
		Type:         reflect.TypeFor[T](),
		Activator:    activatorFor[T],
	}
	if c.depsTree == nil {
		c.depsTree = make(map[string]any)
	}
	if c.global == nil {
		c.global = &Scope{
			Container: c,
			isGlobal:  true,
			deps:      make(map[string]any),
		}
	}

	initFunc, ok := reflect.PointerTo(dep.Type).MethodByName("Init")
	if !ok {
		panic("Init method not found for " + dep.NameType + " dependency. Maybe you need you pointer symbol in type?")
	}

	for i := range initFunc.Type.NumIn() {
		if i == 0 {
			continue // for any method zero argument would be "this" argument
		}
		arg := initFunc.Type.In(i)
		name := arg.String()
		if arg.Kind() == reflect.Struct || arg.Kind() == reflect.Ptr || arg.Kind() == reflect.Interface {
			if slices.Contains(dep.Dependencies, name) {
				panic("Dependency " + name + " already exists for " + dep.NameType)
			}
			dep.Dependencies = append(dep.Dependencies, name)
		}
	}

	_, ok = c.depsTree[dep.NameType]
	if ok {
		panic("Dependency " + dep.NameType + " already exists in container")
	}
	c.depsTree[dep.NameType] = dep
}
func RequireService[T any](c *Container) (*T, error) {
	return RequireServiceFor[T](c.global)
}
