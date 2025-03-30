package container

import (
	"fmt"
	"reflect"
	"slices"
)

func activatorFor[T any]() (T, error) {
	t := reflect.TypeFor[T]()
	realType := t
	if t.Kind() == reflect.Ptr {
		realType = t.Elem()
	}
	value := reflect.New(realType).Interface()
	return unwrap[T](value)
}

type Lifetime int

const (
	Singleton Lifetime = iota
	Transient
	Scoped
	HostedService
)

type Container struct {
	depsTree     map[string]any // map[string]item[T]
	global       *Scope
	scopeCounter int
}

func AddHostedService[T any](c *Container) { add[T](c, HostedService) }

// Transient lifetime services are created each time they're requested from the service container
func AddTransient[T any](c *Container) { add[T](c, Transient) }

// Singleton lifetime services are created the first time they're requested
func AddSingleton[T any](c *Container) { add[T](c, Singleton) }

// A scoped lifetime indicates that services are created once per client request (connection).
func AddScoped[T any](c *Container) { add[T](c, Scoped) }

func nameFor[T any]() string {
	t := reflect.TypeFor[T]()
	name := t.Name()
	if t.Kind() == reflect.Ptr {
		name = "*"
		pointerTo := t.Elem()
		name = name + nameOf(pointerTo)
	}
	return name
}

func nameOf(t reflect.Type) string {
	name := t.Name()
	if t.Kind() == reflect.Ptr {
		name = "*"
		pointerTo := t.Elem()
		name = name + nameOf(pointerTo)
	}
	return name
}

func (c *Container) createScope() *Scope {
	c.scopeCounter = c.scopeCounter + 1
	return &Scope{
		Container: c,
		id:        c.scopeCounter,
		deps:      make(map[string]any),
	}
}
func iii(ii itemInterface) any { return nil }
func add[T any](c *Container, lifetime Lifetime) {
	t := reflect.TypeFor[T]()
	name := nameFor[T]()
	name2 := fmt.Sprintf("%T", t)
	fmt.Println("Name: ", name, "= Name2: ", name2)

	dep := &itemFor[T]{
		NameType:     name,
		Dependencies: []string{},
		Lifetime:     lifetime,
		Type:         t,
		Activator:    activatorFor[T],
	}
	iii(dep)
	if c.depsTree == nil {
		c.depsTree = make(map[string]any)
	}
	if c.global == nil {
		c.global = c.createScope()
	}

	initFunc, ok := dep.Type.MethodByName("Init")
	if !ok {
		panic("Init method not found for " + dep.NameType + " dependency. Maybe you need you pointer symbol in type?")
	}

	for i := range initFunc.Type.NumIn() {
		if i == 0 {
			continue // for any method zero argument would be "this" argument
		}
		arg := initFunc.Type.In(i)
		name := nameOf(arg)
		fmt.Println("Arg kind: ", arg.Kind(), " name: ", name)
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
func RequireService[T any](c *Container) (T, error) {
	return RequireServiceFor[T](c.global)
}
