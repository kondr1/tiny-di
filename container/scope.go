package container

import (
	"fmt"
	"reflect"
)

type Scope struct {
	*Container
	id   int
	deps map[string]any
}

func (c *Scope) DisposeScope() { c.scopeCounter = c.scopeCounter - 1 }

func (c *Scope) requireService(depName string) (interface{}, error) {
	item, ok := c.depsTree[depName]
	if !ok {
		return nil, fmt.Errorf("dependency %s not found", depName)
	}
	if c.id != 1 && item.Lifetime == Scoped {
		scopeDep, scopeOk := c.deps[depName]
		if scopeDep != nil && scopeOk {
			return scopeDep, nil
		}
	} else if c.id == 1 && item.Lifetime == Scoped {
		return nil, fmt.Errorf("dependency %s is Scope dependency. You should BuildScope at first", depName)
	}
	if item.Lifetime == Singleton {
		globalDep, globalOk := c.global.deps[depName]
		if globalDep != nil && globalOk {
			return globalDep, nil
		}
	}

	args := make([]reflect.Value, 0)

	for _, v := range item.Dependencies {
		d, err := c.requireService(v)
		if err != nil {
			return nil, err
		}
		args = append(args, reflect.ValueOf(d))
	}

	dep := activator(item.Type)
	depValue := reflect.ValueOf(dep)
	fmt.Println(depValue.Kind())
	depValue.MethodByName("Init").Call(args)
	if c.id != 1 && item.Lifetime == Scoped {
		c.deps[depName] = dep
	}
	if item.Lifetime == Singleton {
		c.global.deps[depName] = dep
	}
	return dep, nil
}
func RequireScopedService[T any](s *Scope) (T, error) {
	dep, err := s.RequireService(reflect.TypeFor[T]())
	if err != nil {
		return *new(T), err
	}
	return dep.(T), nil
}
func (c *Scope) RequireService(t reflect.Type) (interface{}, error) {
	return c.requireService(nameOf(t))
}
