package container

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type lifetime int

const (
	Singleton lifetime = iota
	Transient
	Scoped
	HostedService
)

// Container is the main dependency injection container that manages service registration,
// dependency resolution, and lifecycle management.
//
// A Container maintains a registry of service descriptors and their corresponding call sites
// for dependency resolution. It supports different service lifetimes and automatic
// dependency injection through reflection-based constructor discovery.
//
// The container must be built using the Build method before services can be resolved.
// Once built, no new services can be registered.
type Container struct {
	dependenciesRegistry map[string]descriptorInterface // map[string]item[T]
	callSitesRegistry    map[string]callSiteInterface   // callSite registry
	global               *Scope
	built                bool
}

func nameFor[T any]() string  { return fmt.Sprintf("%T", *new(T)) }
func nameForI[T any]() string { return fmt.Sprintf("%T", new(T)) }

func addI[I any, T any](c *Container, lifetime lifetime) {
	if c.built {
		panic(fmt.Errorf("%w: You should call Build() before adding dependencies", ErrContainerAlreadyBuilt))
	}
	nameT := nameFor[T]()
	nameI := nameForI[I]()

	interfaceType := reflect.TypeFor[I]()
	if interfaceType.Kind() != reflect.Interface {
		panic(fmt.Errorf("%w: First type %s argument should be interface type", ErrShouldBeInterfaceType, nameI))
	}

	structType := reflect.TypeFor[T]()
	if structType.Kind() != reflect.Struct {
		panic(fmt.Errorf("%w: Second type %s argument should be struct type", ErrShouldBeStructType, nameT))
	}

	if !reflect.PointerTo(structType).Implements(interfaceType) {
		panic(fmt.Errorf("%w: Second type argument %s should implement interface first type argument %s", ErrShouldImplementInterface, nameT, nameI))
	}

	depByType, okByType := c.dependenciesRegistry[nameT]
	_, okByInterface := c.dependenciesRegistry[nameI]
	if okByType && !okByInterface {
		item := depByType.(*descriptor[T])
		item.Interfaces = append(item.Interfaces, nameI)
		c.dependenciesRegistry[nameI] = depByType
		return
	}
	if okByType && okByInterface {
		panic(fmt.Errorf("%w: Dependency %s implementation of %s already exists in container", ErrTypeAlreadyRegistered, nameT, nameI))
	}
	if !okByType && okByInterface {
		panic(fmt.Errorf("%w: Dependency %s already exists in container", ErrTypeAlreadyRegistered, nameI))
	}

	dep, callSite := add[T](c, lifetime)
	dep.Interfaces = append(dep.Interfaces, nameI)
	c.dependenciesRegistry[nameI] = dep
	c.callSitesRegistry[nameI] = callSite
}
func add[T any](c *Container, lifetime lifetime) (*descriptor[T], *callSite[T]) {
	if c.built {
		panic(fmt.Errorf("%w: You should call Build() before adding dependencies", ErrContainerAlreadyBuilt))
	}
	name := nameFor[T]()
	namePtr := fmt.Sprintf("%T", new(T))
	dep := &descriptor[T]{
		NameType:    name,
		PtrNameType: namePtr,
		Interfaces:  []string{},
		Lifetime:    lifetime,
		Type:        reflect.TypeFor[T](),
	}
	if c.dependenciesRegistry == nil {
		c.dependenciesRegistry = make(map[string]descriptorInterface)
	}
	if c.callSitesRegistry == nil {
		c.callSitesRegistry = make(map[string]callSiteInterface)
	}
	if c.global == nil {
		c.global = &Scope{
			Container: c,
			isGlobal:  true,
			instances: make(map[string]any),
		}
	}
	kind := dep.Type.Kind()
	if kind != reflect.Struct {
		panic(fmt.Errorf("%w: Type %s should be struct type", ErrShouldBeStructType, dep.NameType))
	}
	initFunc, ok := reflect.PointerTo(dep.Type).MethodByName("Init")
	if !ok {
		panic(fmt.Errorf("%w: Init method not found for %s dependency", ErrShouldImplementInitMethod, dep.NameType))
	}
	dependencies := []string{}
	for i := range initFunc.Type.NumIn() {
		if i == 0 {
			continue // for any method zero argument would be "this" argument
		}
		arg := initFunc.Type.In(i)
		name := arg.String()
		if arg.Kind() == reflect.Struct || arg.Kind() == reflect.Ptr || arg.Kind() == reflect.Interface {
			if slices.Contains(dependencies, name) {
				panic("Dependency " + name + " already exists for " + dep.NameType)
			}
			dependencies = append(dependencies, name)
		}
	}

	_, ok = c.dependenciesRegistry[dep.NameType]
	if ok {
		panic("Dependency " + dep.NameType + " already exists in container")
	}
	callSite := &callSite[T]{
		name:            dep.NameType,
		lifetime:        lifetime,
		dependencyNames: dependencies,
		dependencies:    nil,
		instance:        nil,
	}
	c.callSitesRegistry[dep.NameType] = callSite
	c.callSitesRegistry[dep.PtrNameType] = callSite
	c.dependenciesRegistry[dep.NameType] = dep
	c.dependenciesRegistry[dep.PtrNameType] = dep
	return dep, callSite
}

// CreateScope creates a new dependency injection scope for scoped service resolution.
//
// Scoped services registered in the container will be instantiated once per scope
// and reused within that scope. This is particularly useful for web applications
// where you want services to live for the duration of a single HTTP request.
//
// The returned scope should be used to resolve scoped services using RequireServiceFor
// or RequireServiceForI methods.
//
// Example:
//
//	scope := container.CreateScope()
//	defer scope.Close() // Clean up scope resources
//	service, err := RequireServiceFor[MyService](scope)
func (c *Container) CreateScope() *Scope {
	return &Scope{
		Container: c,
		instances: make(map[string]any),
	}
}

func (c *Container) checkCircle(typeName string, visited []string) error {
	source, ok := c.callSitesRegistry[typeName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDependencyNotFound, typeName)
	}
	if visited == nil {
		visited = make([]string, 0)
	}
	visited = append(visited, typeName)
	for _, dep := range source.Deps() {
		if slices.Contains(visited, dep[1:]) {
			circle := strings.Join(visited, " -> ")
			return fmt.Errorf("%w: %s", ErrCircleDependency, circle)
		}
		err := c.checkCircle(dep[1:], visited)
		if err != nil {
			return err
		}
	}
	return nil
}

// Build validates the dependency graph and prepares the container for service resolution.
//
// This method must be called after all services have been registered and before
// attempting to resolve any services. Build performs the following operations:
//
//   - Validates that all dependencies can be resolved
//   - Detects and reports circular dependencies
//   - Builds call sites for efficient service creation
//   - Marks the container as built (no more registrations allowed)
//
// If any validation errors occur (such as missing dependencies or circular references),
// Build will panic with a descriptive error message.
//
// Example:
//
//	container := &Container{}
//	AddSingleton[IService, Service](container)
//	err := container.Build()
//	if err != nil {
//		log.Fatal("Failed to build container:", err)
//	}
func (c *Container) Build() error {
	defer func() { c.built = true }()
	if c.dependenciesRegistry == nil {
		return nil
	}

	for typeName, site := range c.callSitesRegistry {
		err := c.checkCircle(typeName, nil)
		if err != nil {
			err = fmt.Errorf("in %s found: %w", typeName, err)
			panic(err)
		}
		err = site.BuildCallSite(c)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

// AddHostedService registers a hosted service with the container.
//
// Hosted services are long-running background services that implement the IHostedService
// interface. They have Start and Stop methods for lifecycle management and are typically
// used for background tasks, schedulers, or other services that need to run continuously.
//
// The service type T must implement IHostedService interface and have an Init method
// for dependency injection.
//
// Example:
//
//	type BackgroundWorker struct {
//		logger ILogger
//	}
//
//	func (b *BackgroundWorker) Init(logger ILogger) error {
//		b.logger = logger
//		return nil
//	}
//
//	func (b *BackgroundWorker) Start(ctx context.Context) error {
//		// Start background work
//		return nil
//	}
//
//	func (b *BackgroundWorker) Stop(ctx context.Context) error {
//		// Stop background work
//		return nil
//	}
//
//	AddHostedService[BackgroundWorker](container)
func AddHostedService[T any](c *Container) { add[T](c, HostedService) }

// AddTransientWithoutInterface registers a transient service without an interface mapping.
//
// Transient services are created each time they're requested from the service container.
// Each resolution call will create a new instance. This registration method is used when
// you want to register the concrete type directly without mapping it to an interface.
//
// The service type T must have an Init method for dependency injection.
//
// Example:
//
//	type EmailService struct {
//		config Config
//	}
//
//	func (e *EmailService) Init(config Config) error {
//		e.config = config
//		return nil
//	}
//
//	AddTransientWithoutInterface[EmailService](container)
//	// Later resolve as:
//	service, err := RequireService[EmailService](container)
func AddTransientWithoutInterface[T any](c *Container) { add[T](c, Transient) }

// AddSingletonWithoutInterface registers a singleton service without an interface mapping.
//
// Singleton services are created the first time they're requested and the same instance
// is reused for all subsequent requests. This registration method is used when you want
// to register the concrete type directly without mapping it to an interface.
//
// The service type T must have an Init method for dependency injection.
//
// Example:
//
//	type DatabaseConnection struct {
//		connectionString string
//	}
//
//	func (d *DatabaseConnection) Init() error {
//		d.connectionString = "connection_string_here"
//		return nil
//	}
//
//	AddSingletonWithoutInterface[DatabaseConnection](container)
func AddSingletonWithoutInterface[T any](c *Container) { add[T](c, Singleton) }

// AddScopedWithoutInterface registers a scoped service without an interface mapping.
//
// Scoped services are created once per client request (scope). Within the same scope,
// the same instance is returned for all requests. This registration method is used when
// you want to register the concrete type directly without mapping it to an interface.
//
// The service type T must have an Init method for dependency injection.
//
// Example:
//
//	type RequestContext struct {
//		requestId string
//		user      User
//	}
//
//	func (r *RequestContext) Init(user User) error {
//		r.user = user
//		r.requestId = generateId()
//		return nil
//	}
//
//	AddScopedWithoutInterface[RequestContext](container)
func AddScopedWithoutInterface[T any](c *Container) { add[T](c, Scoped) }

// AddTransient registers a transient service with interface mapping.
//
// Transient services are created each time they're requested from the service container.
// This method registers a concrete implementation T that will be resolved when interface I
// is requested. The implementation T must implement interface I.
//
// Type parameters:
//   - I: The interface type that will be used for service resolution
//   - T: The concrete implementation type that implements interface I
//
// Example:
//
//	type IEmailService interface {
//		SendEmail(to, subject, body string) error
//	}
//
//	type SMTPEmailService struct {
//		config SMTPConfig
//	}
//
//	func (s *SMTPEmailService) Init(config SMTPConfig) error {
//		s.config = config
//		return nil
//	}
//
//	func (s *SMTPEmailService) SendEmail(to, subject, body string) error {
//		// SMTP implementation
//		return nil
//	}
//
//	AddTransient[IEmailService, SMTPEmailService](container)
func AddTransient[I any, T any](c *Container) { addI[I, T](c, Transient) }

// AddSingleton registers a singleton service with interface mapping.
//
// Singleton services are created the first time they're requested and the same instance
// is reused for all subsequent requests. This method registers a concrete implementation T
// that will be resolved when interface I is requested.
//
// Type parameters:
//   - I: The interface type that will be used for service resolution
//   - T: The concrete implementation type that implements interface I
//
// Example:
//
//	type ILogger interface {
//		Log(message string)
//	}
//
//	type FileLogger struct {
//		filename string
//	}
//
//	func (f *FileLogger) Init() error {
//		f.filename = "app.log"
//		return nil
//	}
//
//	func (f *FileLogger) Log(message string) {
//		// Write to file
//	}
//
//	AddSingleton[ILogger, FileLogger](container)
func AddSingleton[I any, T any](c *Container) { addI[I, T](c, Singleton) }

// AddScoped registers a scoped service with interface mapping.
//
// Scoped services are created once per scope and reused within that scope.
// This method registers a concrete implementation T that will be resolved when
// interface I is requested within a specific scope.
//
// Type parameters:
//   - I: The interface type that will be used for service resolution
//   - T: The concrete implementation type that implements interface I
//
// Example:
//
//	type IUserContext interface {
//		GetUserId() string
//		GetUserRoles() []string
//	}
//
//	type RequestUserContext struct {
//		userId string
//		roles  []string
//	}
//
//	func (r *RequestUserContext) Init() error {
//		// Initialize from request context
//		return nil
//	}
//
//	func (r *RequestUserContext) GetUserId() string {
//		return r.userId
//	}
//
//	func (r *RequestUserContext) GetUserRoles() []string {
//		return r.roles
//	}
//
//	AddScoped[IUserContext, RequestUserContext](container)
func AddScoped[I any, T any](c *Container) { addI[I, T](c, Scoped) }

// RequireServicePtr resolves a service instance from the container's global scope.
//
// This method retrieves a service instance of type T from the container. The service
// must have been previously registered using one of the Add* methods, and the container
// must be built before calling this method.
//
// Returns a pointer to the service instance and an error if resolution fails.
//
// Example:
//
//	AddSingleton[ILogger, FileLogger](container)
//	container.Build()
//
//	logger, err := RequireServicePtr[FileLogger](container)
//	if err != nil {
//		log.Fatal("Failed to resolve logger:", err)
//	}
func RequireServicePtr[T any](c *Container) (*T, error) {
	return RequireServicePtrForScope[T](c.global)
}

func RequireService[T any](c *Container) (T, error) {
	return RequireServiceForScope[T](c.global)
}
