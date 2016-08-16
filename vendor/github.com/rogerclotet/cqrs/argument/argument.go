package argument

import "fmt"

// Argument represents a command or query argument
type Argument struct {
	value interface{}
}

// New creates a new Argument with a given value
func New(value interface{}) Argument {
	return Argument{value: value}
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
func (a Arguments) Get(name string) (Argument, error) {
	value, ok := a[name]
	if !ok {
		return Argument{}, fmt.Errorf("argument not found: %s", name)
	}

	return value, nil
}

// GetInt returns an argument as int if it exists
func (a Arguments) GetInt(name string) (int, error) {
	arg, err := a.Get(name)
	if err != nil {
		return 0, err
	}

	return arg.Int()
}

// GetString returns an argument as string if it exists
func (a Arguments) GetString(name string) (string, error) {
	arg, err := a.Get(name)
	if err != nil {
		return "", err
	}

	return arg.String()
}
