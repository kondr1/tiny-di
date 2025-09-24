# Tiny DI

Package container provides a lightweight dependency injection container for Go applications.
    
Inspired by [.NET dependency injection](https://learn.microsoft.com/en-us/dotnet/core/extensions/dependency-injection)  
  
This package implements a type-safe dependency injection system with support for different
service lifetimes (Singleton, Transient, Scoped, and HostedService). It uses Go generics
to provide compile-time type safety and automatically resolves dependencies through
reflection-based constructor injection.

### Basic Usage

Create a container, register services, build the container, and resolve dependencies:
```go
	c := &Container{}
	AddSingleton[IUserService, UserService](c)
	AddTransient[IEmailService, EmailService](c)
	err := c.Build()
	if err != nil {
		log.Fatal(err)
	}

	userService, err := RequireService[*UserService](c)
	if err != nil {
		log.Fatal(err)
	}
```
### Service Lifetimes

The container supports four service lifetimes:

  - Singleton: Created once and reused for all requests
  - Transient: Created new for every request
  - Scoped: Created once per scope (useful for request-scoped services)
  - HostedService: Long-running background services with Start/Stop lifecycle

### Constructor Injection

Services must implement an Init method that receives dependencies as parameters.
The container automatically discovers and injects dependencies based on parameter types:
```go
	type UserService struct {
		email IEmailService
		db    IDatabase
	}

	func (u *UserService) Init(email IEmailService, db IDatabase) error {
		u.email = email
		u.db = db
		return nil
	}
```
  
### Complete Example

Here's a complete example demonstrating the container usage:
```go
	package main

	import (
		"context"
		"fmt"
		"log"
	)

	// Define interfaces
	type ILogger interface {
		Log(message string)
	}

	type IUserRepository interface {
		GetUser(id string) (*User, error)
	}

	type IUserService interface {
		GetUserDetails(id string) (*UserDetails, error)
	}

	// Define models
	type User struct {
		ID   string
		Name string
	}

	type UserDetails struct {
		User      *User
		LastLogin string
	}

	// Implement services
	type ConsoleLogger struct{}

	func (c *ConsoleLogger) Init() error { return nil }
	func (c *ConsoleLogger) Log(message string) {
		fmt.Printf("[LOG] %s\n", message)
	}

	type InMemoryUserRepository struct {
		logger ILogger
	}

	func (r *InMemoryUserRepository) Init(logger ILogger) error {
		r.logger = logger
		return nil
	}

	func (r *InMemoryUserRepository) GetUser(id string) (*User, error) {
		r.logger.Log(fmt.Sprintf("Fetching user: %s", id))
		return &User{ID: id, Name: "John Doe"}, nil
	}

	type UserService struct {
		repo   IUserRepository
		logger ILogger
	}

	func (s *UserService) Init(repo IUserRepository, logger ILogger) error {
		s.repo = repo
		s.logger = logger
		return nil
	}

	func (s *UserService) GetUserDetails(id string) (*UserDetails, error) {
		s.logger.Log("Getting user details")
		user, err := s.repo.GetUser(id)
		if err != nil {
			return nil, err
		}
		return &UserDetails{User: user, LastLogin: "2024-01-15"}, nil
	}

	func main() {
		// Create and configure container
		container := &Container{}

		// Register services
		AddSingleton[ILogger, ConsoleLogger](container)
		AddSingleton[IUserRepository, InMemoryUserRepository](container)
		AddTransient[IUserService, UserService](container)

		// Build container
		if err := container.Build(); err != nil {
			log.Fatal("Failed to build container:", err)
		}

		// Resolve and use services
		userService, err := RequireService[IUserService](container)
		if err != nil {
			log.Fatal("Failed to resolve user service:", err)
		}

		details, err := userService.GetUserDetails("123")
		if err != nil {
			log.Fatal("Failed to get user details:", err)
		}

		fmt.Printf("User: %s (Last login: %s)\n", details.User.Name, details.LastLogin)
	}
```
