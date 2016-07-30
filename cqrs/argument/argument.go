package argument

import "fmt"

// Argument represents a command or query argument
type Argument struct {
	value interface{}
}

// Int returns an int value for an argument, or an error if it is not an int
func (a Argument) Int() (int, error) {
	value, ok := a.value.(int)
	if !ok {
		return 0, fmt.Errorf("%v is not an int", a)
	}

	return value, nil
}

// String returns a string value for an argument, or an error if it is not a string
func (a Argument) String() (string, error) {
	value, ok := a.value.(string)
	if !ok {
		return "", fmt.Errorf("%v is not a string", a)
	}

	return value, nil
}

// Arguments represent a set of command or query arguments
type Arguments map[string]Argument

// Get returns an argument by name if it exists
func (a Arguments) Get(name string) (interface{}, error) {
	value, ok := a[name]
	if !ok {
		return nil, fmt.Errorf("argument not found: %s", name)
	}

	return value, nil
}
