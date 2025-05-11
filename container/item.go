package container

import (
	"reflect"
)

type descriptorInterface interface {
	Life() Lifetime
	Ints() []string
	Name() string
	TypeOf() reflect.Type
}

type descriptor[T any] struct {
	NameType   string
	Interfaces []string
	Lifetime   Lifetime
	Type       reflect.Type
}

func activatorFor[T any]() *T           { return new(T) }
func (i *descriptor[T]) Life() Lifetime { return i.Lifetime }

func (i *descriptor[T]) Ints() []string       { return i.Interfaces }
func (i *descriptor[T]) Name() string         { return i.NameType }
func (i *descriptor[T]) TypeOf() reflect.Type { return i.Type }
