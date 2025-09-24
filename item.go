package container

import (
	"reflect"
)

type descriptorInterface interface {
	Life() lifetime

	Ints() []string

	Name() string

	TypeOf() reflect.Type
}

type descriptor[T any] struct {
	NameType    string
	PtrNameType string
	Interfaces  []string
	Lifetime    lifetime
	Type        reflect.Type
}

func activatorFor[T any]() *T           { return new(T) }
func (i *descriptor[T]) Life() lifetime { return i.Lifetime }

func (i *descriptor[T]) Ints() []string       { return i.Interfaces }
func (i *descriptor[T]) Name() string         { return i.NameType }
func (i *descriptor[T]) TypeOf() reflect.Type { return i.Type }
