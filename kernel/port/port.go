package port

import (
	"fmt"
	"reflect"

	"github.com/qatoolist/wowapi/v2/kernel/appmodel"
)

// Key represents a typed port key.
type Key[T any] struct {
	id string
}

// NewKey creates a typed Key.
func NewKey[T any](id string) Key[T] {
	return Key[T]{id: id}
}

// ID returns the string identifier of the key.
func (k Key[T]) ID() string {
	return k.id
}

// Define registers a port definition under the given key.
func Define[T any, S any](r appmodel.Registrar[S], key Key[T]) error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return r.DefinePort(key.ID(), t)
}

// Provide registers an implementation for a defined port.
func Provide[T any, S any](r appmodel.Registrar[S], key Key[T], impl T) error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return r.ProvidePort(key.ID(), impl, t)
}

// Require registers a requirement for a defined port.
func Require[T any, S any](r appmodel.Registrar[S], key Key[T]) error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return r.RequirePort(key.ID(), t)
}

// Resolve resolves the implementation of a defined port.
func Resolve[T any, S any](r appmodel.Registrar[S], key Key[T]) (T, error) {
	var zero T
	val, err := r.ResolvePort(key.ID())
	if err != nil {
		return zero, err
	}
	typedVal, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("port %s type conflict: resolved value is not of type %T", key.ID(), zero)
	}
	return typedVal, nil
}
