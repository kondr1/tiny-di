package container

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type lifetime int

const (
	Singleton lifetime = iota
	Transient
	Scoped
	HostedService
)

// Container is the main dependency injection container that manages service registration,
// dependency resolution, and lifecycle management.
//
// The container must be built using the Build method before services can be resolved.
// Once built, no new services can be registered.
type Container struct {
	callSitesRegistry  map[string]callSiteInterface
	hostedServiceSites []callSiteInterface // HostedService callSites in registration order
	global             *Scope
	built              bool
	hostedServices     []IHostedService
}

func nameFor[T any]() string  { return fmt.Sprintf("%T", *new(T)) }
func nameForI[T any]() string { return fmt.Sprintf("%T", new(T)) }

func addI[I any, T any](c *Container, lifetime lifetime) {
	if c.built {
		panic(fmt.Errorf("%w: Cannot add dependencies after Build()", ErrContainerAlreadyBuilt))
	}
	nameT := nameFor[T]()
	nameI := nameForI[I]()

	interfaceType := reflect.TypeFor[I]()
	if interfaceType.Kind() != reflect.Interface {
		panic(fmt.Errorf("%w: First type %s argument should be interface type", ErrShouldBeInterfaceType, nameI))
	}

	structType := reflect.TypeFor[T]()
	if structType.Kind() != reflect.Struct {
		panic(fmt.Errorf("%w: Second type %s argument should be struct type", ErrShouldBeStructType, nameT))
	}

	if !reflect.PointerTo(structType).Implements(interfaceType) {
		panic(fmt.Errorf("%w: Second type argument %s should implement interface first type argument %s", ErrShouldImplementInterface, nameT, nameI))
	}

	depByType, okByType := c.callSitesRegistry[nameT]
	_, okByInterface := c.callSitesRegistry[nameI]
	if okByType && !okByInterface {
		c.callSitesRegistry[nameI] = depByType
		return
	}
	if okByType && okByInterface {
		panic(fmt.Errorf("%w: Dependency %s implementation of %s already exists in container", ErrTypeAlreadyRegistered, nameT, nameI))
	}
	if !okByType && okByInterface {
		panic(fmt.Errorf("%w: Dependency %s already exists in container", ErrTypeAlreadyRegistered, nameI))
	}

	callSite := add[T](c, lifetime)
	c.callSitesRegistry[nameI] = callSite
}
func add[T any](c *Container, lifetime lifetime) *callSite[T] {
	if c.built {
		panic(fmt.Errorf("%w: Cannot add dependencies after Build()", ErrContainerAlreadyBuilt))
	}
	depNameType := nameFor[T]()
	depPtrNameType := fmt.Sprintf("%T", new(T))
	depType := reflect.TypeFor[T]()
	if c.callSitesRegistry == nil {
		c.callSitesRegistry = make(map[string]callSiteInterface)
	}
	if c.global == nil {
		c.global = &Scope{
			Container: c,
			isGlobal:  true,
			instances: make(map[string]any),
		}
	}
	kind := depType.Kind()
	if kind != reflect.Struct {
		panic(fmt.Errorf("%w: Type %s should be struct type", ErrShouldBeStructType, depNameType))
	}
	initFunc, ok := reflect.PointerTo(depType).MethodByName("Init")
	if !ok {
		panic(fmt.Errorf("%w: Init method not found for %s dependency", ErrShouldImplementInitMethod, depNameType))
	}
	// Validate Init signature: must return exactly one error value
	if initFunc.Type.NumOut() != 1 {
		panic(fmt.Errorf("%w: Init method for %s must return exactly one value (error), got %d", ErrShouldImplementInitMethod, depNameType, initFunc.Type.NumOut()))
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !initFunc.Type.Out(0).Implements(errorType) {
		panic(fmt.Errorf("%w: Init method for %s must return error, got %s", ErrShouldImplementInitMethod, depNameType, initFunc.Type.Out(0)))
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
				panic("Dependency " + name + " already exists for " + depNameType)
			}
			dependencies = append(dependencies, name)
		}
	}

	_, ok = c.callSitesRegistry[depNameType]
	if ok {
		panic("Dependency " + depNameType + " already exists in container")
	}
	callSite := &callSite[T]{
		name:            depNameType,
		lifetime:        lifetime,
		dependencyNames: dependencies,
		dependencies:    nil,
		initMethod:      initFunc,
		instance:        nil,
	}
	c.callSitesRegistry[depNameType] = callSite
	c.callSitesRegistry[depPtrNameType] = callSite

	if lifetime == HostedService {
		c.hostedServiceSites = append(c.hostedServiceSites, callSite)
	}

	return callSite
}

// CreateScope creates a new dependency injection scope for scoped service resolution.
//
// The returned scope should be used to resolve scoped services using [RequireServiceFor]
// or [RequireServiceForT] functions.
func (c *Container) CreateScope() *Scope {
	if !c.built {
		panic(fmt.Errorf("%w: You should call Build() before CreateScope()", ErrContainerNotBuilt))
	}
	return &Scope{
		Container: c,
		instances: make(map[string]any),
	}
}

func (c *Container) checkCircle(typeName string, visited []string) error {
	source, ok := c.callSitesRegistry[typeName]
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

// Build validates the dependency graph and prepares the container for service resolution.
//
// This method must be called after all services have been registered and before
// attempting to resolve any services. Build performs the following operations:
//
//   - Validates that all dependencies can be resolved
//   - Detects and reports circular dependencies
//   - Builds call sites for efficient service creation
//   - Marks the container as built (no more registrations allowed)
func (c *Container) Build() {
	defer func() { c.built = true }()
	if c.callSitesRegistry == nil {
		return
	}

	for typeName, site := range c.callSitesRegistry {
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

	for _, site := range c.hostedServiceSites {
		instance, err := site.Build(c.global)
		if err != nil {
			panic(fmt.Errorf("failed to build hosted service %s: %w", site.Name(), err))
		}
		if hostedSvc, ok := instance.(IHostedService); ok {
			c.hostedServices = append(c.hostedServices, hostedSvc)
		}
	}
}

// AddHostedService registers a hosted service with the container.
//
// HostedService must implement IHostedService interface with Start/Stop methods.
// Services are started in registration order and stopped in reverse order.
func AddHostedService[T any](c *Container) { add[T](c, HostedService) }

// AddTransientWithoutInterface registers a transient service without an interface mapping.
//
// Transient services are created each time they're requested from the service container.
func AddTransientWithoutInterface[T any](c *Container) { add[T](c, Transient) }

// AddSingletonWithoutInterface registers a singleton service without an interface mapping.
//
// Singleton services are created the first time they're requested and the same instance
// is reused for all subsequent requests.
func AddSingletonWithoutInterface[T any](c *Container) { add[T](c, Singleton) }

// AddScopedWithoutInterface registers a scoped service without an interface mapping.
//
// Scoped services are created once per client request (scope). Within the same scope,
// the same instance is returned for all requests.
func AddScopedWithoutInterface[T any](c *Container) { add[T](c, Scoped) }

// AddTransient registers a transient service with interface mapping.
//
// Transient services are created each time they're requested from the service container.
//
// Type parameters:
//   - I: The interface type that will be used for service resolution
//   - T: The concrete implementation type that implements interface I
func AddTransient[I any, T any](c *Container) { addI[I, T](c, Transient) }

// AddSingleton registers a singleton service with interface mapping.
//
// Singleton services are created the first time they're requested and the same instance
// is reused for all subsequent requests.
//
// Type parameters:
//   - I: The interface type that will be used for service resolution
//   - T: The concrete implementation type that implements interface I
func AddSingleton[I any, T any](c *Container) { addI[I, T](c, Singleton) }

// AddScoped registers a scoped service with interface mapping.
//
// Scoped services are created once per scope and reused within that scope.
//
// Type parameters:
//   - I: The interface type that will be used for service resolution
//   - T: The concrete implementation type that implements interface I
func AddScoped[I any, T any](c *Container) { addI[I, T](c, Scoped) }

// RequireServicePtr resolves a service instance from the container's global scope.
//
// This method retrieves a service instance of type T from the container.
// Returns a pointer to the service instance and an error if resolution fails.
// Shuld be used only for struct types.
func RequireServicePtr[T any](c *Container) (*T, error) {
	return RequireServicePtrForScope[T](c.global)
}

// RequireService resolves a service instance from the container's global scope.
//
// This method retrieves a service instance of type T from the container.
// Returns the service instance and an error if resolution fails.
// Shuld be used only for pointer to struct types or interface types.
func RequireService[T any](c *Container) (T, error) {
	return RequireServiceForScope[T](c.global)
}

// StartAsync starts all registered HostedService instances.
//
// Services are started in registration order. If any service fails to start,
// the error is returned immediately and remaining services are not started.
//
// The provided context can be used for cancellation and timeouts.
func (c *Container) StartAsync(ctx context.Context) error {
	if !c.built {
		return fmt.Errorf("%w: You should call Build() before StartAsync()", ErrContainerNotBuilt)
	}

	for i, svc := range c.hostedServices {
		if err := svc.Start(ctx); err != nil {
			return fmt.Errorf("failed to start hosted service #%d: %w", i, err)
		}
	}
	return nil
}

// StopAsync stops all registered HostedService instances.
//
// Services are stopped in reverse registration order (LIFO).
// All services are stopped regardless of errors, and the first error is returned.
//
// The provided context typically includes a timeout for graceful shutdown.
func (c *Container) StopAsync(ctx context.Context) error {
	if !c.built {
		return fmt.Errorf("%w: You should call Build() before StopAsync()", ErrContainerNotBuilt)
	}

	var firstErr error
	for i := len(c.hostedServices) - 1; i >= 0; i-- {
		if err := c.hostedServices[i].Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
