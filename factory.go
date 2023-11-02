package cmotel

import (
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var lock = &sync.Mutex{}
var singleton *cmOtel

// CreateSingleton create a singleton structure
func CreateSingleton(intialTracer trace.Tracer, serviceName string) CMOtel {
	if singleton == nil {
		lock.Lock()
		defer lock.Unlock()

		singleton = &cmOtel{
			tracer:             intialTracer,
			serviceName:        serviceName,
			spans:              map[string]cmSpan{},
			relationships:      map[string]string{},
			spanIDToNameMapper: map[string]string{},
		}
	}

	return singleton
}

// Singleton get the singleton object. If the object was not previously created it will create it with the default otel tracer.
func Singleton() CMOtel {
	if singleton == nil {
		return CreateSingleton(otel.Tracer(""), "service-name")
	}

	return singleton
}

// New creates a new object to handle the traces
func New(initialTracer trace.Tracer, serviceName string) CMOtel {
	return &cmOtel{
		tracer:             initialTracer,
		serviceName:        serviceName,
		spans:              map[string]cmSpan{},
		relationships:      map[string]string{},
		spanIDToNameMapper: map[string]string{},
	}

}
