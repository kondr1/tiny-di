package container

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type Lifetime int

const (
	Singleton Lifetime = iota
	Transient
	Scoped
	HostedService
)

type Container struct {
	dependenciesRegistry map[string]descriptorInterface // map[string]item[T]
	callSites            map[string]callSiteInterface   // callSite registry
	global               *Scope
	builded              bool
}

func nameFor[T any]() string  { return fmt.Sprintf("%T", *new(T)) }
func nameForI[T any]() string { return fmt.Sprintf("%T", new(T)) }

func addI[I any, T any](c *Container, lifetime Lifetime) {
	if c.builded {
		panic("container already builded. You should call Build() before adding dependencies")
	}
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

	if !reflect.PointerTo(structType).Implements(interfaceType) {
		panic("Second type argument " + nameT + " should implement interface first type argument " + nameI)
	}

	depByType, okByType := c.dependenciesRegistry[nameT]
	_, okByInterface := c.dependenciesRegistry[nameI]
	if okByType && !okByInterface {
		item := depByType.(*descriptor[T])
		item.Interfaces = append(item.Interfaces, nameI)
		c.dependenciesRegistry[nameI] = depByType
		return
	}
	if okByType && okByInterface {
		panic("Dependency " + nameT + " implementation of " + nameI + " already exists in container")
	}
	if !okByType && okByInterface {
		panic("Dependency " + nameI + " already exists in container")
	}

	dep, callSite := add[T](c, lifetime)
	dep.Interfaces = append(dep.Interfaces, nameI)
	c.dependenciesRegistry[nameI] = dep
	c.callSites[nameI] = callSite
}
func add[T any](c *Container, lifetime Lifetime) (*descriptor[T], *callSite[T]) {
	if c.builded {
		panic("container already builded. You should call Build() before adding dependencies")
	}
	name := nameFor[T]()
	dep := &descriptor[T]{
		NameType:   name,
		Interfaces: []string{},
		Lifetime:   lifetime,
		Type:       reflect.TypeFor[T](),
	}
	if c.dependenciesRegistry == nil {
		c.dependenciesRegistry = make(map[string]descriptorInterface)
	}
	if c.callSites == nil {
		c.callSites = make(map[string]callSiteInterface)
	}
	if c.global == nil {
		c.global = &Scope{
			Container: c,
			isGlobal:  true,
			instances: make(map[string]any),
		}
	}

	initFunc, ok := reflect.PointerTo(dep.Type).MethodByName("Init")
	if !ok {
		panic("Init method not found for " + dep.NameType + " dependency. Maybe you need you pointer symbol in type?")
	}
	dependencies := []string{}
	for i := range initFunc.Type.NumIn() {
		if i == 0 {
			continue // for any method zero argument would be "this" argument
		}
		arg := initFunc.Type.In(i)
		name := arg.String()
		if arg.Kind() == reflect.Struct || arg.Kind() == reflect.Ptr || arg.Kind() == reflect.Interface {
			if slices.Contains(dependencies, name) {
				panic("Dependency " + name + " already exists for " + dep.NameType)
			}
			dependencies = append(dependencies, name)
		}
	}

	_, ok = c.dependenciesRegistry[dep.NameType]
	if ok {
		panic("Dependency " + dep.NameType + " already exists in container")
	}
	callSite := &callSite[T]{
		name:            dep.NameType,
		lifetime:        lifetime,
		dependencyNames: dependencies,
		dependencies:    nil,
		singleton:       nil,
	}
	c.callSites[dep.NameType] = callSite
	c.dependenciesRegistry[dep.NameType] = dep
	return dep, callSite
}

func (c *Container) CreateScope() *Scope {
	return &Scope{
		Container: c,
		instances: make(map[string]any),
	}
}

func (c *Container) checkCircle(typeName string, visited []string) error {
	source, ok := c.callSites[typeName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDependencyNotFound, typeName)
	}
	if visited == nil {
		visited = make([]string, 0)
	}
	visited = append(visited, typeName)
	for _, dep := range source.Deps() {
		if slices.Contains(visited, dep[1:]) {
			circle := strings.Join(visited, " -> ")
			return fmt.Errorf("%w: %s", ErrCircleDependency, circle)
		}
		err := c.checkCircle(dep[1:], visited)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) Build() error {
	defer func() { c.builded = true }()
	if c.dependenciesRegistry == nil {
		return nil
	}

	for typeName, site := range c.callSites {
		err := c.checkCircle(typeName, nil)
		if err != nil {
			err = fmt.Errorf("in %s found: %w", typeName, err)
			panic(err)
		}
		err = site.BuildCallSite(c)
		if err != nil {
			panic(err)
		}
	}

	return nil
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

func RequireService[T any](c *Container) (*T, error) {
	return RequireServiceFor[T](c.global)
}
func RequireServiceI[I any](c *Container) (I, error) {
	return RequireServiceForI[I](c.global)
}
