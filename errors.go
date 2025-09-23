package container

import "errors"

var (
	ErrUnknownLifetime               = errors.New("unknown lifetime")
	ErrFailedToBuildDependency       = errors.New("failed to build dependency")
	ErrScopeIsNil                    = errors.New("scope is nil")
	ErrScopedDependencyInGlobalScope = errors.New("called scoped dependency in global scope")
	ErrDependencyNotFound            = errors.New("dependency not found")
	ErrCircleDependency              = errors.New("circle dependency")
)
