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
func AddTransientWithoutInterface[T any](c *Container) { add[T](c, Transient) }

// Singleton lifetime services are created the first time they're requested
func AddSingletonWithoutInterface[T any](c *Container) { add[T](c, Singleton) }

// A scoped lifetime indicates that services are created once per client request (connection).
func AddScopedWithoutInterface[T any](c *Container) { add[T](c, Scoped) }

// Transient lifetime services are created each time they're requested from the service container
func AddTransient[I any, T any](c *Container) { addI[I, T](c, Transient) }

// Singleton lifetime services are created the first time they're requested
func AddSingleton[I any, T any](c *Container) { addI[I, T](c, Singleton) }

// A scoped lifetime indicates that services are created once per client request (connection).
func AddScoped[I any, T any](c *Container) { addI[I, T](c, Scoped) }

func nameFor[T any]() string  { return fmt.Sprintf("%T", *new(T)) }
func nameForI[T any]() string { return fmt.Sprintf("%T", new(T)) }

func (c *Container) CreateScope() *Scope {
	return &Scope{
		Container: c,
		deps:      make(map[string]any),
	}
}
func addI[I any, T any](c *Container, lifetime Lifetime) {
	nameT := nameFor[T]()
	nameI := nameForI[I]()

	interfaceType := reflect.TypeFor[I]()
	if interfaceType.Kind() != reflect.Interface {
		panic("First type " + nameI + " argument should be interface type")
	}

	structType := reflect.TypeFor[T]()
	if structType.Kind() != reflect.Struct {
		panic("Second type " + nameT + " argument should be struct type")
	}
	pointerType := reflect.PointerTo(structType)
	fmt.Printf("%v: %v", pointerType, interfaceType)
	if interfaceType.Implements(reflect.PointerTo(structType)) {
		panic("Second type argument " + nameT + " should implement interface first type argument " + nameI)
	}

	depByType, okByType := c.depsTree[nameT]
	_, okByInterface := c.depsTree[nameI]
	if okByType && !okByInterface {
		item := depByType.(*itemFor[T])
		item.Interfaces = append(item.Interfaces, nameI)
		c.depsTree[nameI] = depByType
		return
	}
	if okByType && okByInterface {
		panic("Dependency " + nameT + " implementation of " + nameI + " already exists in container")
	}
	if !okByType && okByInterface {
		panic("Dependency " + nameI + " already exists in container")
	}

	dep := add[T](c, lifetime)
	dep.Interfaces = append(dep.Interfaces, nameI)
	c.depsTree[nameI] = dep
}
func add[T any](c *Container, lifetime Lifetime) *itemFor[T] {
	name := nameFor[T]()
	dep := &itemFor[T]{
		NameType:     name,
		Dependencies: []string{},
		Interfaces:   []string{},
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
	return dep
}
func RequireService[T any](c *Container) (*T, error) {
	return RequireServiceFor[T](c.global)
}
func RequireServiceI[I any](c *Container) (I, error) {
	return RequireServiceForI[I](c.global)
}
