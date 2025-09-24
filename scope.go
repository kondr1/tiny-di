package container

import (
	"fmt"
	"reflect"
)

// Scope represents a dependency injection scope for managing scoped service instances.
//
// A scope maintains its own set of service instances separate from the global container.
// Scoped services registered in the container will be instantiated once per scope and
// reused within that scope. This is particularly useful for web applications where
// services should live for the duration of a single HTTP request.
//
// Scopes are created using Container.CreateScope() and should be used with
// RequireServiceFor and RequireServiceForI functions for service resolution.
type Scope struct {
	*Container

	// isGlobal indicates whether this scope is the container's global scope.
	// Global scopes cannot resolve scoped services.
	isGlobal bool

	// instances holds the cached instances of scoped services for this scope.
	// The key is the service type name, and the value is the service instance.
	instances map[string]any
}

func unwrapPtr[T any](v any) (T, error) {
	if v == nil {
		return *new(T), fmt.Errorf("failed unwrap of type %T. Value is nil", *new(T))
	}
	castTPtr, ok := v.(T)
	if !ok {
		if reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).Elem().Kind() == reflect.Ptr {
			return unwrapPtr[T](reflect.ValueOf(v).Elem().Interface())
		}
	}
	return castTPtr, nil
}

func unwrapT[T any](v any) (*T, error) {
	if v == nil {
		return nil, fmt.Errorf("failed unwrap of type %T. Value is nil", *new(T))
	}
	castTPtr, ok := v.(*T)
	if !ok {
		return nil, fmt.Errorf("failed unwrap of type %T", *new(T))
	}
	return castTPtr, nil
}

func unwrapI[I any](v any) (I, error) {
	if v == nil {
		return *new(I), fmt.Errorf("failed unwrap of type %T. Value is nil", *new(I))
	}
	castInterfacePtr, ok := v.(I)
	if !ok {
		return *new(I), fmt.Errorf("failed unwrap of type %T", *new(I))
	}
	return castInterfacePtr, nil
}

// RequireServicePtrForScope resolves a service instance of type T from the specified scope.
//
// This function retrieves a service instance from the dependency injection container
// within the context of the provided scope. The service must have been previously
// registered and the container must be built before calling this function.
//
// Service resolution behavior by lifetime:
//   - Singleton: Returns the same instance regardless of scope
//   - Transient: Creates a new instance for each call
//   - Scoped: Returns the same instance within the same scope, creates new for different scopes
//   - HostedService: Behaves like singleton
//
// The function panics if:
//   - The container is not built
//   - Requested type [T] is not a struct
//
// Returns a pointer to the service instance and an error if resolution fails.
//
// Example:
//
//	scope := container.CreateScope()
//	service, err := RequireServicePtrForScope[UserService](scope)
//	if err != nil {
//		log.Printf("Failed to resolve UserService: %v", err)
//		return
//	}
func RequireServicePtrForScope[T any](s *Scope) (*T, error) {
	if !s.built {
		panic(fmt.Errorf("%w: You should call Build() before RequireServiceFor", ErrContainerNotBuilt))
	}
	nameDep := nameFor[T]()
	if nameDep == "" || nameDep == "<nil>" {
		panic(fmt.Errorf("%w Maybe you should use RequireServiceForI for interfaces?", ErrExtractDependencyName))
	}
	item, ok := s.callSitesRegistry[nameDep]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDependencyNotFound, nameDep)
	}
	dep, err := item.Build(s)
	if err != nil {
		return nil, err
	}
	return unwrapT[T](dep)
}

func RequireServiceForScope[T any](s *Scope) (T, error) {
	if !s.built {
		panic(fmt.Errorf("%w: You should call Build() before RequireServiceFor", ErrContainerNotBuilt))
	}
	nameDep := nameFor[T]()
	if nameDep == "" || nameDep == "<nil>" {
		nameDep = nameForI[T]()
		if nameDep == "" || nameDep == "<nil>" {
			panic(fmt.Errorf("%w", ErrExtractDependencyName))
		}
	}
	item, ok := s.callSitesRegistry[nameDep]
	if !ok {
		return *new(T), fmt.Errorf("%w: %s", ErrDependencyNotFound, nameDep)
	}
	dep, err := item.Build(s)
	if err != nil {
		return *new(T), err
	}
	return unwrapPtr[T](dep)
}
