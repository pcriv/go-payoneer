package payoneer

import (
	"encoding/json"
	"fmt"
)

// Optional is a generic wrapper for nullable or optional API fields.
type Optional[T any] struct {
	value T
	ok    bool
}

// Some returns an Optional with the provided value and ok set to true.
func Some[T any](v T) Optional[T] {
	return Optional[T]{value: v, ok: true}
}

// None returns an Optional with ok set to false.
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// Get returns the value and a boolean indicating if the value is present.
func (o Optional[T]) Get() (T, bool) {
	return o.value, o.ok
}

// OrDefault returns the value if present, or the provided default value otherwise.
func (o Optional[T]) OrDefault(d T) T {
	if o.ok {
		return o.value
	}

	return d
}

// MarshalJSON implements the json.Marshaler interface.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.ok {
		return []byte("null"), nil
	}

	return json.Marshal(o.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.value = *new(T)
		o.ok = false

		return nil
	}

	if err := json.Unmarshal(data, &o.value); err != nil {
		return fmt.Errorf("failed to unmarshal Optional value: %w", err)
	}
	o.ok = true

	return nil
}
