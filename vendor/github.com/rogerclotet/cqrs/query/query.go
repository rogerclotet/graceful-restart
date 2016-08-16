package query

import (
	"context"
	"fmt"

	"github.com/rogerclotet/cqrs/argument"
)

// Query represents a query to be executed, which contains its name, args and a response channel
type Query struct {
	name     string
	args     argument.Arguments
	response chan Response
}

// New creates a new query with given name and arguments
func New(name string, args argument.Arguments) Query {
	return Query{
		name:     name,
		args:     args,
		response: make(chan Response, 1),
	}
}

// Name returns the name of the query
func (q Query) Name() string {
	return q.name
}

// Response returns a channel which will receive the response when the query is handled
func (q Query) Response() chan Response {
	return q.response
}

// Respond allows to send a query response, allowed once per query
func (q Query) Respond(res Response) {
	q.response <- res
	close(q.response)
}

// Response encapsulate the response and error of a query
type Response struct {
	response interface{}
	err      error
}

// NewResponse creates a new QueryResponse with given response and error
func NewResponse(res interface{}, err error) Response {
	return Response{
		response: res,
		err:      err,
	}
}

// Response getter
func (q Response) Response() interface{} {
	return q.response
}

// Err getter
func (q Response) Err() error {
	return q.err
}

// Handler is a function which handles a query
type Handler func(ctx context.Context, args argument.Arguments) (interface{}, error)

// RegisteredQuery represents the relationship between a query name and its handler
type RegisteredQuery struct {
	name    string
	handler Handler
}

// NewRegisteredQuery returns a new RegisteredQuery with the given name and handler
func NewRegisteredQuery(name string, handler Handler) RegisteredQuery {
	return RegisteredQuery{
		name:    name,
		handler: handler,
	}
}

// Registry is a set of query handlers indexed by name
type Registry map[string]Handler

// NewRegistry creates a new Registry with the given query handlers
func NewRegistry(query ...RegisteredQuery) (Registry, error) {
	registry := make(Registry)
	for _, q := range query {
		if _, ok := registry[q.name]; ok {
			return nil, fmt.Errorf("name already registered: %s", q.name)
		}
		registry[q.name] = q.handler
	}
	return registry, nil
}

// Handle receives a Query and handles it using the registered handlers
func (r Registry) Handle(ctx context.Context, q Query) (interface{}, error) {
	handler, ok := r[q.name]
	if !ok {
		return nil, fmt.Errorf("query not registered: %s", q.name)
	}

	return handler(ctx, q.args)
}
