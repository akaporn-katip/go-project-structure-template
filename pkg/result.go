package pkg

// Result represents a value that can either be successful or an error.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a new successful Result.
func Ok[T any](val T) Result[T] {
	return Result[T]{value: val}
}

// Err creates a new erroneous Result.
func Err[T any](e error) Result[T] {
	return Result[T]{err: e}
}

// IsOk returns true if the Result holds a successful value.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// Get returns the value and error.
func (r Result[T]) Get() (T, error) {
	return r.value, r.err
}
