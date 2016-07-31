package command_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rogerclotet/graceful-restart/cqrs/argument"
	"github.com/rogerclotet/graceful-restart/cqrs/command"
	"github.com/stretchr/testify/assert"
)

func ExampleRegisterAndHandleCommand() {
	helloHandler := func() command.Handler {
		return func(ctx context.Context, args argument.Arguments) error {
			name, _ := args.GetString("name")
			fmt.Printf("Hello %s!", name)

			return nil
		}
	}

	r, _ := command.NewRegistry(
		command.NewRegisteredCommand("hello", helloHandler()),
	)

	cmd := command.New("hello", argument.Arguments{"name": argument.New("Roger")})
	_ = r.Handle(context.TODO(), cmd)

	// Output: Hello Roger!
}

func TestHandlingUnregisteredCommandReturnsError(t *testing.T) {
	r, _ := command.NewRegistry()
	cmd := command.New("test", argument.Arguments{})

	err := r.Handle(context.TODO(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command not registered")
}

func TestAnErrorInHandlerReturnsError(t *testing.T) {
	testHandler := func() command.Handler {
		return func(ctx context.Context, args argument.Arguments) error {
			n, _ := args.GetInt("n")
			return fmt.Errorf("error: %d", n)
		}
	}
	r, _ := command.NewRegistry(
		command.NewRegisteredCommand("test", testHandler()),
	)
	cmd := command.New("test", argument.Arguments{"n": argument.New(123)})

	err := r.Handle(context.TODO(), cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error: 123")
}
