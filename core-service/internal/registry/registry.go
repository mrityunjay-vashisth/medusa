// core-service/internal/registry/registry.go
package registry

import "sync"

// ServiceRegistry provides a generic way for services to register and retrieve dependencies
type ServiceRegistry interface {
	// Register registers a service with a name
	Register(name string, service interface{})

	// Get retrieves a service by name
	Get(name string) interface{}
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() ServiceRegistry {
	return &serviceRegistryImpl{
		services: make(map[string]interface{}),
	}
}

type serviceRegistryImpl struct {
	services map[string]interface{}
	mutex    sync.RWMutex
}

func (r *serviceRegistryImpl) Register(name string, service interface{}) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.services[name] = service
}

func (r *serviceRegistryImpl) Get(name string) interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.services[name]
}
