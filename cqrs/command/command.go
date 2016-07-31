package command

import (
	"context"
	"fmt"

	"github.com/rogerclotet/graceful-restart/cqrs/argument"
)

// Command represents a command to be executed, which contains its name and arguments
type Command struct {
	name string
	args argument.Arguments
}

// New creates a new command with given name and arguments
func New(name string, args argument.Arguments) Command {
	return Command{
		name: name,
		args: args,
	}
}

// Name returns the name of the command
func (c Command) Name() string {
	return c.name
}

// Handler is a function which handles a command
type Handler func(ctx context.Context, args argument.Arguments) error

// RegisteredCommand represents the relationship between a command name and its handler
type RegisteredCommand struct {
	name    string
	handler Handler
}

// NewRegisteredCommand returns a new RegisteredCommand with the given name and handler
func NewRegisteredCommand(name string, handler Handler) RegisteredCommand {
	return RegisteredCommand{
		name:    name,
		handler: handler,
	}
}

// Registry is a set of command handlers indexed by name
type Registry map[string]Handler

// NewRegistry creates a new Registry with the given command handlers
func NewRegistry(commands ...RegisteredCommand) (Registry, error) {
	registry := make(Registry)
	for _, c := range commands {
		if _, ok := registry[c.name]; ok {
			return nil, fmt.Errorf("name already registered: %s", c.name)
		}
		registry[c.name] = c.handler
	}
	return registry, nil
}

// Handle receives a Command and handles it using the registered handlers
func (r Registry) Handle(ctx context.Context, c Command) error {
	handler, ok := r[c.name]
	if !ok {
		return fmt.Errorf("command not registered: %s", c.name)
	}

	return handler(ctx, c.args)
}
