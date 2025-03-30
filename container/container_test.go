package container

import (
	"strconv"
	"testing"
)

var counterA int

type A struct{ Str string }

func (a *A) Init() error {
	counterA++
	a.Str = "Only A " + strconv.Itoa(counterA)
	return nil
}

var counterB int

type B struct{ Str string }

func (b *B) Init(a *A) error {
	counterB++
	b.Str = a.Str + " and B " + strconv.Itoa(counterB)
	return nil
}

var counterC int

type C struct{ Str string }

func (c *C) Init(a *A) error {
	counterC++
	c.Str = a.Str + " and C " + strconv.Itoa(counterC)
	return nil
}

func BuildContainer() *Container {
	c := &Container{}
	AddTransient[*A](c)
	AddTransient[A](c)   // наверное будет ошибка при запросе этого
	AddTransient[**A](c) // а хуй его знает чо будет
	AddSingleton[*B](c)
	AddScoped[*C](c)
	return c
}

// TODO: Что если зависимостей не хватает?
// TODO: Что если зависимость не инициализируется?
// TODO: Что если запрашиваем Scoped вне скоупа?
// TODO: Добавить интерфейсы
// TODO: Добавить HostedService
// TODO: Добавить метод Build что бы не искать зависимость во время выполнения

func TestSingleton(t *testing.T) {
	c := BuildContainer()
	first, err := RequireService[*B](c)
	if first.Str != "Only A 1 and B 1" || err != nil {
		t.Errorf(", got %v", first.Str)
		return
	}
	second, err := RequireService[*B](c)
	if second.Str != "Only A 1 and B 1" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestTransient(t *testing.T) {
	c := BuildContainer()
	first, err := RequireService[*A](c)
	if first.Str != "Only A 1" || err != nil {
		t.Errorf(", got %v", first.Str)
		return
	}
	second, err := RequireService[*A](c)
	if second.Str != "Only A 2" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestScoped(t *testing.T) {
	c := BuildContainer()
	scope := c.BuildScope()
	defer scope.DisposeScope()
	first, err := RequireScopedService[*C](scope)
	if first.Str != "Only A 1 and C 1" || err != nil {
		t.Errorf(", got %v", first.Str)
		return
	}
	second, err := RequireScopedService[*C](scope)
	if second.Str != "Only A 1 and C 1" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}
