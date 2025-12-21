package container

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type CounterInterface interface{}
type AInterface interface{ String() string }
type BInterface interface{ String() string }
type CInterface interface{ String() string }
type DInterface interface{ String() string }

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

type D struct{ Str string }

func (d *D) String() string { return d.Str }
func (d *D) Init(c *C) error {
	d.Str = c.Str + " and D "
	return nil
}

func GetContainer() *Container {
	c := &Container{}
	AddTransient[AInterface, A](c)
	AddSingleton[BInterface, B](c)
	AddScoped[CInterface, C](c)
	AddSingleton[CounterInterface, Counter](c)
	return c
}
func BuildContainer() *Container {
	c := GetContainer()
	c.Build()
	return c
}

func TestSingleton(t *testing.T) {
	c := BuildContainer()
	first, err := RequireService[BInterface](c)
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
	second, err := RequireServicePtr[B](c)
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
	first, err := RequireService[AInterface](c)
	if first.String() != "Only A 1" || err != nil {
		t.Errorf(", got %v", first.String())
		return
	}
	second, err := RequireServicePtr[A](c)
	if second.Str != "Only A 2" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestScoped(t *testing.T) {
	c := BuildContainer()
	scope := c.CreateScope()
	first, err := RequireServiceForScope[CInterface](scope)
	if first.String() != "Only A 1 and C 2" || err != nil {
		t.Errorf(", got %v", first.String())
		return
	}
	second, err := RequireServicePtrForScope[C](scope)
	if second.Str != "Only A 1 and C 2" || err != nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestCaptiveDependency(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
		if !errors.Is(errors.Unwrap(r.(error)), ErrCaptiveDependency) {
			t.Errorf("Error not same: %s", r)
		}
	}()

	c := GetContainer()

	AddSingleton[DInterface, D](c)

	c.Build()
}

func TestScopedOutOfScope(t *testing.T) {
	c := BuildContainer()
	first, err := RequireService[CInterface](c)
	if err == nil {
		t.Errorf(", got %v", first.String())
		return
	}
	second, err := RequireServicePtr[C](c)
	if err == nil {
		t.Errorf(", got %v", second.Str)
	}
}

func TestNameFor(t *testing.T) {
	counterType := reflect.TypeFor[*Counter]()
	counter := &Counter{I: 0}
	var defaultCounter *Counter
	name1 := nameForT[*Counter]()
	name2 := fmt.Sprintf("%T", counter)
	name3 := fmt.Sprintf("%T", defaultCounter)
	name4 := nameForT[context.Context]()
	nameI5 := nameForPtr[context.Context]()
	nameI := nameForPtr[CounterInterface]()

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

func TestNotFound(t *testing.T) {
	c := BuildContainer()
	_, err := RequireServicePtr[NotInContainer](c)
	if err == nil || !errors.Is(err, ErrDependencyNotFound) {
		t.Errorf("Expected ErrDependencyNotFound, got %v", err)
	}
	t.Log(err)
}

func TestRequireInterfaceService(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
		if errors.Is(errors.Unwrap(r.(error)), ErrExtractDependencyName) {
			return
		}
		t.Errorf("Error not same: %s", r.(error))
	}()
	c := BuildContainer()
	_, err := RequireServicePtr[AInterface](c)
	if err == nil || !errors.Is(err, ErrDependencyNotFound) {
		t.Errorf("Expected ErrDependencyNotFound, got %v", err)
	}
	t.Log(err)
}
func TestAddPrimitive(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
		if !errors.Is(errors.Unwrap(r.(error)), ErrShouldBeStructType) {
			t.Errorf("Error not same: %s", r)
		}
	}()
	c := GetContainer()
	AddSingletonWithoutInterface[int](c)
}

func TestAlreadyHas(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
		if !errors.Is(errors.Unwrap(r.(error)), ErrTypeAlreadyRegistered) {
			t.Errorf("Error not same: %s", r)
		}
	}()
	c := GetContainer()

	AddScoped[CInterface, C](c)
	c.Build()
}

func TestAlreadyHasInterface(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}
		if !errors.Is(errors.Unwrap(r.(error)), ErrTypeAlreadyRegistered) {
			t.Errorf("Error not same: %s", r)
		}
	}()
	c := GetContainer()

	AddScoped[CInterface, D](c)
	c.Build()
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
		err := r.(error)
		if !errors.Is(errors.Unwrap(err), ErrCircleDependency) {
			t.Errorf("Error not same: %s", r)
		}
	}()

	c := GetContainer()
	AddTransientWithoutInterface[CircleOne](c)
	AddTransientWithoutInterface[CircleTwo](c)

	c.Build()

	_, _ = RequireServicePtr[CircleOne](c)
}

type MyHostedService struct {
	started bool
	stopped bool
}

func (h *MyHostedService) Init() error { return nil }

func (h *MyHostedService) Start(ctx context.Context) error {
	if h.started {
		return fmt.Errorf("already started")
	}
	h.started = true
	return nil
}

func (h *MyHostedService) Stop(ctx context.Context) error {
	if h.stopped {
		return fmt.Errorf("already stoped")
	}
	h.stopped = true
	return nil
}

func TestHostedService(t *testing.T) {
	c := &Container{}
	AddHostedService[MyHostedService](c)
	c.Build()

	ctx := context.Background()
	err := c.StartAsync(ctx)
	if err != nil {
		t.Errorf("StartAsync failed: %v", err)
	}

	if len(c.hostedServices) != 1 {
		t.Errorf("Expected 1 hosted service, got %d", len(c.hostedServices))
	}

	svc, err := RequireServicePtr[MyHostedService](c)

	if err != nil {
		t.Errorf("Service not found in container: %v", err)
	}
	if !svc.started {
		t.Errorf("Service was not started")
	}

	err = c.StopAsync(ctx)
	if err != nil {
		t.Errorf("StopAsync failed: %v", err)
	}

	if !svc.stopped {
		t.Errorf("Service was not stopped")
	}
}

type CountingHostedService struct {
	startCount int
	stopCount  int
}

func (h *CountingHostedService) Init() error { return nil }

func (h *CountingHostedService) Start(ctx context.Context) error {
	h.startCount++
	return nil
}

func (h *CountingHostedService) Stop(ctx context.Context) error {
	h.stopCount++
	return nil
}

func TestHostedServiceNoDuplicates(t *testing.T) {
	c := &Container{}
	AddHostedService[CountingHostedService](c)
	c.Build()

	t.Logf("Number of hosted services: %d", len(c.hostedServices))

	ctx := context.Background()
	c.StartAsync(ctx)

	svc, err := RequireServicePtr[CountingHostedService](c)
	if err != nil {
		t.Errorf("Service not found in container: %v", err)
	}
	if svc.startCount != 1 {
		t.Errorf("Start should be called once, got %d times", svc.startCount)
	}

	c.StopAsync(ctx)

	if svc.stopCount != 1 {
		t.Errorf("Stop should be called once, got %d times", svc.stopCount)
	}
}

func TestStartAsyncBeforeBuild(t *testing.T) {
	c := &Container{}
	AddHostedService[MyHostedService](c)

	ctx := context.Background()
	err := c.StartAsync(ctx)
	if err == nil {
		t.Errorf("Expected error when calling StartAsync before Build")
	}
	if !errors.Is(err, ErrContainerNotBuilt) {
		t.Errorf("Expected ErrContainerNotBuilt, got %v", err)
	}
}

func TestStopAsyncBeforeBuild(t *testing.T) {
	c := &Container{}
	AddHostedService[MyHostedService](c)

	ctx := context.Background()
	err := c.StopAsync(ctx)
	if err == nil {
		t.Errorf("Expected error when calling StopAsync before Build")
	}
	if !errors.Is(err, ErrContainerNotBuilt) {
		t.Errorf("Expected ErrContainerNotBuilt, got %v", err)
	}
}

type FailingStartService struct{}

func (f *FailingStartService) Init() error { return nil }
func (f *FailingStartService) Start(ctx context.Context) error {
	return fmt.Errorf("startup failed")
}
func (f *FailingStartService) Stop(ctx context.Context) error { return nil }

func TestHostedServiceStartupError(t *testing.T) {
	c := &Container{}
	AddHostedService[MyHostedService](c)
	AddHostedService[FailingStartService](c)
	AddHostedService[CountingHostedService](c)
	c.Build()

	ctx := context.Background()
	err := c.StartAsync(ctx)
	if err == nil {
		t.Errorf("Expected error when service fails to start")
	}

	svc1, _ := RequireServicePtr[MyHostedService](c)
	if !svc1.started {
		t.Errorf("First service should have started")
	}

	svc3, _ := RequireServicePtr[CountingHostedService](c)
	if svc3.startCount != 0 {
		t.Errorf("Third service should not have started, but startCount=%d", svc3.startCount)
	}
}

type FailingStopService struct {
	started bool
}

func (f *FailingStopService) Init() error { return nil }
func (f *FailingStopService) Start(ctx context.Context) error {
	f.started = true
	return nil
}
func (f *FailingStopService) Stop(ctx context.Context) error {
	return fmt.Errorf("shutdown failed")
}

func TestHostedServiceShutdownError(t *testing.T) {
	c := &Container{}
	AddHostedService[CountingHostedService](c)
	AddHostedService[FailingStopService](c)
	AddHostedService[MyHostedService](c)
	c.Build()

	ctx := context.Background()
	c.StartAsync(ctx)

	err := c.StopAsync(ctx)
	if err == nil {
		t.Errorf("Expected error when service fails to stop")
	}

	svc1, _ := RequireServicePtr[CountingHostedService](c)
	if svc1.stopCount != 1 {
		t.Errorf("First service should have stopped, stopCount=%d", svc1.stopCount)
	}

	svc3, _ := RequireServicePtr[MyHostedService](c)
	if !svc3.stopped {
		t.Errorf("Third service should have stopped")
	}
}

type OrderTracker struct {
	startSeq []string
	stopSeq  []string
}

func (o *OrderTracker) Init() error { return nil }
func (o *OrderTracker) RecordStart(name string) {
	o.startSeq = append(o.startSeq, name)
}
func (o *OrderTracker) RecordStop(name string) {
	o.stopSeq = append(o.stopSeq, name)
}

type FirstOrderedService struct{ tracker *OrderTracker }

func (f *FirstOrderedService) Init(t *OrderTracker) error {
	f.tracker = t
	return nil
}
func (f *FirstOrderedService) Start(ctx context.Context) error {
	f.tracker.RecordStart("first")
	return nil
}
func (f *FirstOrderedService) Stop(ctx context.Context) error {
	f.tracker.RecordStop("first")
	return nil
}

type SecondOrderedService struct{ tracker *OrderTracker }

func (s *SecondOrderedService) Init(t *OrderTracker) error {
	s.tracker = t
	return nil
}
func (s *SecondOrderedService) Start(ctx context.Context) error {
	s.tracker.RecordStart("second")
	return nil
}
func (s *SecondOrderedService) Stop(ctx context.Context) error {
	s.tracker.RecordStop("second")
	return nil
}

type ThirdOrderedService struct{ tracker *OrderTracker }

func (t *ThirdOrderedService) Init(tracker *OrderTracker) error {
	t.tracker = tracker
	return nil
}
func (t *ThirdOrderedService) Start(ctx context.Context) error {
	t.tracker.RecordStart("third")
	return nil
}
func (t *ThirdOrderedService) Stop(ctx context.Context) error {
	t.tracker.RecordStop("third")
	return nil
}

func TestMultipleHostedServicesOrdering(t *testing.T) {
	c := &Container{}
	AddSingletonWithoutInterface[OrderTracker](c)
	AddHostedService[FirstOrderedService](c)
	AddHostedService[SecondOrderedService](c)
	AddHostedService[ThirdOrderedService](c)
	c.Build()

	if len(c.hostedServices) != 3 {
		t.Errorf("Expected 3 hosted services, got %d", len(c.hostedServices))
	}

	ctx := context.Background()
	err := c.StartAsync(ctx)
	if err != nil {
		t.Errorf("StartAsync failed: %v", err)
	}

	tracker, _ := RequireServicePtr[OrderTracker](c)

	// FIFO
	expectedStart := []string{"first", "second", "third"}
	if len(tracker.startSeq) != 3 {
		t.Errorf("Expected 3 starts, got %d", len(tracker.startSeq))
	}
	for i, name := range expectedStart {
		if i >= len(tracker.startSeq) || tracker.startSeq[i] != name {
			t.Errorf("Start order mismatch at %d: expected %s, got %v", i, name, tracker.startSeq)
		}
	}

	err = c.StopAsync(ctx)
	if err != nil {
		t.Errorf("StopAsync failed: %v", err)
	}

	// LIFO
	expectedStop := []string{"third", "second", "first"}
	if len(tracker.stopSeq) != 3 {
		t.Errorf("Expected 3 stops, got %d", len(tracker.stopSeq))
	}
	for i, name := range expectedStop {
		if i >= len(tracker.stopSeq) || tracker.stopSeq[i] != name {
			t.Errorf("Stop order mismatch at %d: expected %s, got %v", i, name, tracker.stopSeq)
		}
	}
}

type ServiceWithFailingInit struct{}

func (s *ServiceWithFailingInit) Init() error {
	return fmt.Errorf("init failed")
}

type ServiceWithBadInitSignature struct{}

func (s *ServiceWithBadInitSignature) Init() {} // No error return!

func TestBadInitSignature(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("Expected panic when Init doesn't return error")
		}
		err := r.(error)
		if !errors.Is(err, ErrShouldImplementInitMethod) {
			t.Errorf("Expected ErrShouldImplementInitMethod, got %v", err)
		}
	}()

	c := &Container{}
	AddSingletonWithoutInterface[ServiceWithBadInitSignature](c)
}

func TestInitMethodError(t *testing.T) {
	c := &Container{}
	AddSingletonWithoutInterface[ServiceWithFailingInit](c)
	c.Build()

	_, err := RequireServicePtr[ServiceWithFailingInit](c)
	if err == nil {
		t.Errorf("Expected error when Init fails")
	}
	if !errors.Is(err, ErrFailedToBuildDependency) {
		t.Errorf("Expected ErrFailedToBuildDependency, got %v", err)
	}
}

type ScopedService struct{}

func (s *ScopedService) Init() error { return nil }

type HostedServiceWithScopedDep struct{}

func (h *HostedServiceWithScopedDep) Init(s *ScopedService) error { return nil }
func (h *HostedServiceWithScopedDep) Start(ctx context.Context) error {
	return nil
}
func (h *HostedServiceWithScopedDep) Stop(ctx context.Context) error { return nil }

func TestHostedServiceCaptiveDependency(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("Expected panic when HostedService depends on Scoped")
		}
		err, ok := r.(error)
		if !ok {
			t.Errorf("Expected error in panic, got %T", r)
		}
		if !errors.Is(err, ErrCaptiveDependency) {
			t.Errorf("Expected ErrCaptiveDependency, got %v", err)
		}
	}()

	c := &Container{}
	AddScopedWithoutInterface[ScopedService](c)
	AddHostedService[HostedServiceWithScopedDep](c)
	c.Build()
}

type TestKeys int

const TestKey TestKeys = iota

func TestAddValueWithContext(t *testing.T) {
	c := &Container{}

	ctx := context.WithValue(context.Background(), TestKey, "test-value")

	AddValue(c, ctx)
	c.Build()

	resolvedCtx, err := RequireService[context.Context](c)
	if err != nil {
		t.Fatalf("Failed to resolve context: %v", err)
	}

	value := resolvedCtx.Value(TestKey)
	if value != "test-value" {
		t.Errorf("Expected context value 'test-value', got %v", value)
	}

	resolvedCtx2, err := RequireService[context.Context](c)
	if err != nil {
		t.Fatalf("Failed to resolve context second time: %v", err)
	}

	value2 := resolvedCtx2.Value(TestKey)
	if value2 != "test-value" {
		t.Errorf("Expected context value 'test-value' on second resolve, got %v", value2)
	}
}

type Config struct {
	Host string
	Port int
}

func TestAddValueWithPointer(t *testing.T) {
	c := &Container{}

	cfg := &Config{
		Host: "localhost",
		Port: 8080,
	}

	AddValue(c, cfg)
	c.Build()

	resolvedCfg, err := RequireService[*Config](c)
	if err != nil {
		t.Fatalf("Failed to resolve config: %v", err)
	}

	if resolvedCfg != cfg {
		t.Errorf("Expected same pointer instance")
	}

	if resolvedCfg.Host != "localhost" || resolvedCfg.Port != 8080 {
		t.Errorf("Config values don't match: got %+v", resolvedCfg)
	}

	resolvedCfg.Port = 9000

	resolvedCfg2, err := RequireService[*Config](c)
	if err != nil {
		t.Fatalf("Failed to resolve config second time: %v", err)
	}

	if resolvedCfg2.Port != 9000 {
		t.Errorf("Expected Port to be 9000 (modified), got %d", resolvedCfg2.Port)
	}

	resolvedCfg3, err := RequireServicePtr[Config](c)
	if err != nil {
		t.Fatalf("Failed to resolve config: %v", err)
	}
	if resolvedCfg3 != cfg {
		t.Errorf("Expected same pointer instance")
	}
}

func TestAddValueWithValueType(t *testing.T) {
	c := &Container{}

	cfg := Config{
		Host: "localhost",
		Port: 3000,
	}

	AddValue(c, cfg)
	c.Build()

	resolvedCfg, err := RequireService[Config](c)
	if err != nil {
		t.Fatalf("Failed to resolve config: %v", err)
	}

	if resolvedCfg.Host != "localhost" || resolvedCfg.Port != 3000 {
		t.Errorf("Config values don't match: got %+v", resolvedCfg)
	}

	resolvedCfgPtr, err := RequireServicePtr[Config](c)
	if err != nil {
		t.Fatalf("Failed to resolve config ptr: %v", err)
	}

	if resolvedCfgPtr.Host != "localhost" || resolvedCfgPtr.Port != 3000 {
		t.Errorf("Config ptr values don't match: got %+v", resolvedCfgPtr)
	}
}

type WithContextDependency struct {
	Ctx context.Context
}

func (w *WithContextDependency) Init(ctx context.Context) error {
	w.Ctx = ctx
	return nil
}
func TestWithContextDependency(t *testing.T) {
	c := &Container{}

	AddTransientWithoutInterface[WithContextDependency](c)

	ctx := context.WithValue(context.Background(), TestKey, "context-in-dependency")

	AddValue(c, ctx)

	c.Build()

	dep, err := RequireServicePtr[WithContextDependency](c)

	if err != nil {
		t.Fatalf("Failed to resolve WithContextDependency: %v", err)
	}

	value := dep.Ctx.Value(TestKey)

	if value != "context-in-dependency" {
		t.Errorf("Expected context value 'context-in-dependency', got %v", value)
	}
}
