package container

import (
	"testing"
)

type TransientDep struct{}

func (d *TransientDep) Init() error { return nil }

// Singleton
type SingletonDep struct{}

func (d *SingletonDep) Init() error { return nil }

// Service with 21 dependencies - will use reflection fallback
type ServiceWith1Deps struct {
	d1 *SingletonDep
}

func (s *ServiceWith1Deps) Init(
	d1 *SingletonDep,
) error {
	s.d1 = d1
	return nil
}

var c *Container
var c2 *Container

func init() {
	c = &Container{}
	AddTransientWithoutInterface[TransientDep](c)
	AddSingletonWithoutInterface[SingletonDep](c)
	AddTransientWithoutInterface[ServiceWith1Deps](c)
	c.Build()

	c2 = &Container{}
	AddSingletonWithoutInterface[SingletonDep](c2)
	AddSingletonWithoutInterface[ServiceWith1Deps](c2)
	c2.Build()
}

func BenchmarkReflectionSingletonDeps(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RequireServicePtr[SingletonDep](c)
	}
}

// Benchmark reflection path for Transient with 25 dependencies
func BenchmarkTransientDeps(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RequireServicePtr[TransientDep](c)
	}
}

// Benchmark reflection path for Transient with 25 dependencies
func BenchmarkServiceWith1Deps1Deps(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RequireServicePtr[ServiceWith1Deps](c)
	}
}

func BenchmarkSingletonServiceWith1Deps1Deps(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RequireServicePtr[ServiceWith1Deps](c2)
	}
}
