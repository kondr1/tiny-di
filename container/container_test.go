package container

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type CounterInterface interface{}
type AInterface interface{ String() string }
type BInterface interface{ String() string }
type CInterface interface{ String() string }

type Counter struct{ I int }

func (c *Counter) Init() error { return nil }

type A struct{ Str string }

func (a *A) String() string { return a.Str }
func (a *A) Init(c *Counter) error {
	c.I++
	a.Str = "Only A " + strconv.Itoa(c.I)
	return nil
}

type B struct{ Str string }

func (b *B) String() string { return b.Str }
func (b *B) Init(a *A, c *Counter) error {
	c.I++
	b.Str = a.Str + " and B " + strconv.Itoa(c.I)
	return nil
}

type C struct{ Str string }

func (c *C) String() string { return c.Str }
func (c *C) Init(a *A, cc *Counter) error {
	cc.I++
	c.Str = a.Str + " and C " + strconv.Itoa(cc.I)
	return nil
}

func BuildContainer() *Container {
	c := &Container{}
	AddTransient[AInterface, A](c)
	AddSingleton[BInterface, B](c)
	AddScoped[CInterface, C](c)
	AddSingleton[CounterInterface, Counter](c)
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
	first, err := RequireServiceI[BInterface](c)
	if first == nil {
		t.Errorf(", got nil. Error %v", err)
		return
	}
	if err != nil {
		t.Errorf(", got %v", err)
		return
	}
	if first.String() != "Only A 1 and B 2" {
		t.Errorf(", got %v", first.String())
		return
	}
	second, err := RequireService[B](c)
	if second == nil {
		t.Errorf(", got second nil %e", err)
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
	first, err := RequireServiceI[AInterface](c)
	if first.String() != "Only A 1" || err != nil {
		t.Errorf(", got %v", first.String())
		return
	}
	second, err := RequireService[A](c)
	if second.Str != "Only A 2" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestScoped(t *testing.T) {
	c := BuildContainer()
	scope := c.CreateScope()
	first, err := RequireServiceForI[CInterface](scope)
	if first.String() != "Only A 1 and C 2" || err != nil {
		t.Errorf(", got %v", first.String())
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

	nameI := nameForI[CounterInterface]()

	if name1 != name2 || name1 != name3 || name1 != counterType.String() {
		t.Errorf("%s and %s is not equal to %s", name3, name2, name1)
	}
	if nameI != "*container.CounterInterface" {
		t.Errorf("%s is not equal to container.CounterInterface", nameI)
	}
}

func TestActivator(t *testing.T) {
	a0 := &A{}
	a1 := activatorFor[*A]()

	if a1 == nil {
		t.Errorf("ActivatorFor returned nil")
		return
	}
	t.Logf("%T, %+v\n", a0, a0)
	t.Logf("%T, %+v\n", a1, *a1)
}

type NotInContainer struct{}

func TestNoFond(t *testing.T) {
	c := BuildContainer()
	_, err := RequireService[NotInContainer](c)
	if err == nil {
		t.Error("Expect error")
	}
	t.Log(err)
}

func TestAlreadyHas(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
		if r != "First type *container.C argument should be interface type" {
			t.Errorf("Error not same: %s", r)
		}
	}()
	c := BuildContainer()
	AddScoped[C, CInterface](c)
}

type CircleOne struct{}
type CircleTwo struct{}

func (*CircleOne) Init(a *CircleTwo) error {
	return nil
}
func (*CircleTwo) Init(a *CircleOne) error {
	return nil
}

func TestCircleDep(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic!")
		}
		if r != "circle dependecy" {
			t.Errorf("Error not same: %s", r)
		}
	}()

	c := BuildContainer()
	AddTransientWithoutInterface[CircleOne](c)
	AddTransientWithoutInterface[CircleTwo](c)
}
