package container

import (
	"fmt"
	"reflect"
	"slices"
)

// func extractType(t reflect.Type, pointerCount int) (reflect.Type, int) {
// 	if t.Kind() == reflect.Ptr {
// 		return extractType(t.Elem(), pointerCount+1)
// 	}
// 	if t.Kind() == reflect.Struct {
// 		return t, pointerCount
// 	}
// 	return nil, 0
// }

// ptr wraps the given value with pointer: V => *V, *V => **V, etc.
// func ptr(v reflect.Value) reflect.Value {
// 	pt := reflect.PointerTo(v.Type()) // create a *T type.
// 	pv := reflect.New(pt.Elem())      // create a reflect.Value of type *T.
// 	pv.Elem().Set(v)                  // sets pv to point to underlying value of v.
// 	return pv
// }

// return *T as interface{}
func activator(T reflect.Type) any {
	realType := T
	if T.Kind() == reflect.Ptr {
		realType = T.Elem()
	}
	return reflect.New(realType).Interface()
}

type Lifetime int

const (
	Singleton Lifetime = iota
	Transient
	Scoped
	HostedService
)

type item struct {
	Name         string
	Dependencies []string
	Lifetime     Lifetime
	Type         reflect.Type
}

type Container struct {
	depsTree     map[string]item
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

func nameOf(t reflect.Type) string {
	name := t.Name()
	if t.Kind() == reflect.Ptr {
		name = "*"
		pointerTo := t.Elem()
		name = name + nameOf(pointerTo)
	}
	return name
}

func add[T any](c *Container, lifetime Lifetime) {
	t := reflect.TypeFor[T]()
	name := nameOf(t)
	name2 := fmt.Sprintf("%T", t)
	fmt.Println("Name: ", name, "= Name2: ", name2)

	item := item{
		Name:         name,
		Dependencies: []string{},
		Lifetime:     lifetime,
		Type:         t,
	}
	c.add(item)
}
func (c *Container) BuildScope() *Scope {
	c.scopeCounter = c.scopeCounter + 1
	return &Scope{
		Container: c,
		id:        c.scopeCounter,
		deps:      make(map[string]any),
	}
}
func (c *Container) add(dep item) {
	if c.depsTree == nil {
		c.depsTree = make(map[string]item)
	}
	if c.global == nil {
		c.global = c.BuildScope()
	}

	initFunc, ok := dep.Type.MethodByName("Init")
	if !ok {
		panic("Init method not found for " + dep.Name + " dependency. Maybe you need you pointer symbol in type?")
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
				panic("Dependency " + name + " already exists for " + dep.Name)
			}
			dep.Dependencies = append(dep.Dependencies, name)
		}
	}

	_, ok = c.depsTree[dep.Name]
	if ok {
		panic("Dependency " + dep.Name + " already exists in container")
	}
	c.depsTree[dep.Name] = dep
}
func RequireService[T any](c *Container) (T, error) {
	depName := nameOf(reflect.TypeFor[T]())
	dep, err := c.global.requireService(depName)
	return dep.(T), err
}
