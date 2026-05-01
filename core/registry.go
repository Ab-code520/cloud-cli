package core

import "fmt"

var drivers = make(map[string]func() Storage)

// Register adds a driver factory to the registry.
func Register(name string, factory func() Storage) {
	drivers[name] = factory
}

// NewDriver creates a new instance of the named driver.
func NewDriver(name string) (Storage, error) {
	f, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("driver '%s' not registered", name)
	}
	return f(), nil
}

// ListDrivers returns all registered driver names.
func ListDrivers() []string {
	names := make([]string, 0, len(drivers))
	for name := range drivers {
		names = append(names, name)
	}
	return names
}
