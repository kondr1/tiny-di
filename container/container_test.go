package container

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type Counter struct{ I int }

func (c *Counter) Init() error { return nil }

type A struct{ Str string }

func (a *A) Init(c *Counter) error {
	c.I++
	a.Str = "Only A " + strconv.Itoa(c.I)
	return nil
}

type B struct{ Str string }

func (b *B) Init(a *A, c *Counter) error {
	c.I++
	b.Str = a.Str + " and B " + strconv.Itoa(c.I)
	return nil
}

type C struct{ Str string }

func (c *C) Init(a *A, cc *Counter) error {
	cc.I++
	c.Str = a.Str + " and C " + strconv.Itoa(cc.I)
	return nil
}

func BuildContainer() *Container {
	c := &Container{}
	AddTransient[A](c)
	AddSingleton[B](c)
	AddScoped[C](c)
	AddSingleton[Counter](c)
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
	first, err := RequireService[B](c)
	if first == nil {
		t.Errorf(", got nil. Error %v", err)
		return
	}
	if err != nil {
		t.Errorf(", got %v", err)
		return
	}
	if first.Str != "Only A 1 and B 2" {
		t.Errorf(", got %v", first.Str)
		return
	}
	second, err := RequireService[B](c)
	if second == nil {
		t.Error(", got second nil")
		return
	}
	if err != nil {
		t.Errorf(", got second %e", err)
		return
	}
	if second.Str != "Only A 1 and B 2" {
		t.Errorf(", got second %v", second.Str)
	}
}

func TestTransient(t *testing.T) {
	c := BuildContainer()
	first, err := RequireService[A](c)
	if first.Str != "Only A 1" || err != nil {
		t.Errorf(", got %v", first.Str)
		return
	}
	second, err := RequireService[A](c)
	if second.Str != "Only A 2" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestScoped(t *testing.T) {
	c := BuildContainer()
	scope := c.createScope()
	first, err := RequireServiceFor[C](scope)
	if first.Str != "Only A 1 and C 2" || err != nil {
		t.Errorf(", got %v", first.Str)
		return
	}
	second, err := RequireServiceFor[C](scope)
	if second.Str != "Only A 1 and C 2" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestNameFor(t *testing.T) {
	counterType := reflect.TypeFor[*Counter]()
	counter := &Counter{I: 0}
	var defaultCounter *Counter
	name1 := nameFor[*Counter]()
	name2 := fmt.Sprintf("%T", counter)
	name3 := fmt.Sprintf("%T", defaultCounter)

	if name1 != name2 || name1 != name3 || name1 != counterType.String() {
		t.Errorf("%s and %s is not equal to %s", name3, name2, name1)
	}
}

func ActivatorTest[T any]() *T { return new(T) }

func TestActivator(t *testing.T) {
	a0 := &A{}
	a1, err := activatorFor[*A]()
	a2 := ActivatorTest[*A]()
	a3 := ActivatorTest[A]()
	if err != nil {
		t.Errorf("ActivatorFor returned error %v", err)
		return
	}
	if a1 == nil {
		t.Errorf("ActivatorFor returned nil")
		return
	}
	t.Logf("%T, %+v", a0, a0)
	t.Logf("%T, %+v", a1, *a1)
	t.Logf("%T, %+v", a2, *a2)
	t.Logf("%T, %+v", a3, a3)
}
