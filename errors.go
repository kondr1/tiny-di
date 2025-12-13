package container

import "errors"

// Package-level errors used throughout the dependency injection container.
// These errors provide detailed information about various failure modes
// during service registration, container building, and service resolution.
var (
	// ErrUnknownLifetime is returned when an unrecognized service lifetime is encountered.
	// This typically indicates a programming error or corruption in the container's internal state.
	ErrUnknownLifetime = errors.New("unknown lifetime")

	// ErrFailedToBuildDependency is returned when the container fails to construct a service instance.
	// This can occur when a service's Init method returns an error or when there are issues
	// with dependency resolution during the construction process.
	ErrFailedToBuildDependency = errors.New("failed to build dependency")

	// ErrScopeIsNil is returned when a nil scope is provided to functions that require a valid scope.
	// This typically occurs when attempting to resolve scoped services without a proper scope context.
	ErrScopeIsNil = errors.New("scope is nil")

	// ErrScopedDependencyInGlobalScope is returned when attempting to resolve a scoped service
	// from the global scope. Scoped services must be resolved within a specific scope created
	// using Container.CreateScope().
	ErrScopedDependencyInGlobalScope = errors.New("called scoped dependency in global scope")

	// ErrDependencyNotFound is returned when attempting to resolve a service that has not been
	// registered with the container. This typically occurs when there's a mismatch between
	// registered services and their dependencies, or when requesting an unregistered service.
	ErrDependencyNotFound = errors.New("dependency not found")

	// ErrCircleDependency is returned during container building when a circular dependency
	// is detected in the service registration graph. Circular dependencies prevent the
	// container from determining a valid construction order for services.
	ErrCircleDependency = errors.New("circle dependency")

	// ErrExtractDependencyName is returned when the system fails to extract the name of a dependency.
	// User should use functions RequireServiceFor for require struct types or RequireServiceForI for require interface types for resolve this case.
	ErrExtractDependencyName = errors.New("can't extract dependency name")

	// ErrTypeAlreadyRegistered is returned when attempting to register a type that has already been registered.
	ErrTypeAlreadyRegistered = errors.New("type already registered")

	// ErrShouldBeStructType is returned when a type expected to be a struct is not.
	ErrShouldBeStructType = errors.New("should be struct type")

	// ErrShouldBeInterfaceType is returned when a type expected to be an interface is not.
	ErrShouldBeInterfaceType = errors.New("should be interface type")

	// ErrShouldImplementInterface is returned when a type does not implement the required interface.
	ErrShouldImplementInterface = errors.New("should implement interface")

	// ErrShouldImplementInitMethod is returned when a type does not have the required Init method.
	ErrShouldImplementInitMethod = errors.New("should implement Init method")

	// ErrContainerNotBuilt is returned when attempting to resolve services from a container
	// that has not been built yet. The container must be built using the Build() method
	// before resolving any services.
	ErrContainerNotBuilt = errors.New("container not built")

	// ErrContainerAlreadyBuilt is returned when attempting to register services with a container
	// that has already been built. Once a container is built, no new services can be registered.
	ErrContainerAlreadyBuilt = errors.New("container already built")

	// ErrCaptiveDependency occurs when a longer-lived service (e.g., singleton) depends on a shorter-lived service (e.g., scoped or transient).
	ErrCaptiveDependency = errors.New("singleton calls scoped or transient")
)
