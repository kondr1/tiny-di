package container

import (
	"fmt"
	"reflect"
)

type callSiteInterface interface {
	Name() string
	Deps() []string
	Build(s *Scope) (any, error)
	BuildSingleton() (any, error)
	BuildTransient() (any, error)
	BuildScoped(s *Scope) (any, error)
	BuildCallSite(c *Container) error
}

type callSite[T any] struct {
	name            string
	lifetime        lifetime
	dependencyNames []string
	dependencies    []callSiteInterface
	built           bool
	instance        *T
}

func (c *callSite[T]) Name() string   { return c.name }
func (c *callSite[T]) Deps() []string { return c.dependencyNames }

func (c *callSite[T]) Build(s *Scope) (any, error) { return c.build(s) }
func (c *callSite[T]) build(s *Scope) (*T, error) {
	switch c.lifetime {
	case HostedService:
		fallthrough
	case Singleton:
		return c.buildSingleton()
	case Transient:
		return c.buildTransient()
	case Scoped:
		return c.buildScoped(s)
	default:
		return nil, fmt.Errorf("%w: for %s", ErrUnknownLifetime, c.Name())
	}
}

func (c *callSite[T]) Constructor(s *Scope) (any, error) { return c.constructor(s) }
func (c *callSite[T]) constructor(s *Scope) (*T, error) {
	resolved := activatorFor[T]()
	args := make([]reflect.Value, 0)
	for _, v := range c.dependencies {
		dep, err := v.Build(s)
		if err != nil {
			return nil, fmt.Errorf("failed to build dependency %s: %w", v.Name(), err)
		}
		args = append(args, reflect.ValueOf(dep))
	}

	errVal := reflect.
		ValueOf(resolved).
		MethodByName("Init").
		Call(args)
	errValf := errVal[0]
	if errValf.Interface() != nil {
		err := errValf.Interface().(error)
		return nil, fmt.Errorf("%w: for %s: %w", ErrFailedToBuildDependency, c.Name(), err)
	}

	return resolved, nil
}

func (c *callSite[T]) BuildSingleton() (any, error) { return c.buildSingleton() }
func (c *callSite[T]) buildSingleton() (*T, error) {
	if c.instance != nil {
		return c.instance, nil
	}
	obj, err := c.constructor(nil)
	if err != nil {
		return nil, err
	}
	c.instance = obj
	return c.instance, nil
}

func (c *callSite[T]) BuildTransient() (any, error) { return c.buildTransient() }
func (c *callSite[T]) buildTransient() (*T, error) {
	return c.constructor(nil)
}

func (c *callSite[T]) BuildScoped(s *Scope) (any, error) { return c.buildScoped(s) }
func (c *callSite[T]) buildScoped(s *Scope) (*T, error) {
	if s == nil {
		return nil, ErrScopeIsNil
	}
	if s.isGlobal {
		return nil, ErrScopedDependencyInGlobalScope
	}
	if s.instances[c.name] != nil {
		return s.instances[c.name].(*T), nil
	}
	obj, err := c.constructor(s)
	if err != nil {
		return nil, err
	}
	s.instances[c.name] = obj
	return obj, nil
}

func (c *callSite[T]) BuildCallSite(container *Container) error {
	if c.built {
		return nil
	}
	dependencies := make([]callSiteInterface, 0, len(c.dependencyNames))
	for _, depName := range c.dependencyNames {
		site, ok := container.callSitesRegistry[depName[1:]]
		if !ok {
			return fmt.Errorf("%w: %s not found for %s", ErrDependencyNotFound, depName, c.Name())
		}
		dependencies = append(dependencies, site)
	}
	c.dependencies = dependencies
	return nil
}
