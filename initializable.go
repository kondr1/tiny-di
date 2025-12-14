package container

// Initializable interfaces for type switch optimization.
// These interfaces allow avoiding reflection for constructors with 0-5 dependencies.

type Initializable0 interface {
	Init() error
}

type Initializable1[T any] interface {
	Init(dep1 T) error
}
