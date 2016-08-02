package query_test

import (
	"context"
	"errors"
	"fmt"

	"math/rand"
	"testing"

	"strconv"

	"github.com/rogerclotet/graceful-restart/cqrs/argument"
	"github.com/rogerclotet/graceful-restart/cqrs/query"
	"github.com/stretchr/testify/assert"
)

func ExampleRegisterAndHandleQuery() {
	timesQueryHandler := func() query.Handler {
		return func(ctx context.Context, args argument.Arguments) (interface{}, error) {
			a, err := args.GetInt("a")
			if err != nil {
				return 0, errors.New("could not get argument a")
			}
			b, err := args.GetInt("b")
			if err != nil {
				return 0, errors.New("could not get argument b")
			}

			return a * b, nil
		}
	}
	rq := query.NewRegisteredQuery("multiply", timesQueryHandler())
	r, _ := query.NewRegistry(rq)

	a := argument.New(3)
	b := argument.New(5)
	q := query.New("multiply", argument.Arguments{"a": a, "b": b})
	result, _ := r.Handle(context.TODO(), q)

	fmt.Println(result)

	// Output: 15
}

func TestQueryName(t *testing.T) {
	name := strconv.Itoa(rand.Int())
	q := query.New(name, argument.Arguments{})

	assert.Equal(t, name, q.Name())
}

func TestQueryResponseGetters(t *testing.T) {
	value := rand.Int()
	err := errors.New("test error")

	r := query.NewResponse(value, err)

	assert.Equal(t, value, r.Response())
	assert.Equal(t, err, r.Err())
}

func TestQueryResponse(t *testing.T) {
	q := query.New("test", argument.Arguments{})
	response := query.NewResponse(rand.Int(), nil)

	go q.Respond(response)

	returnedResponse := <-q.Response()

	assert.Equal(t, response, returnedResponse)
}

func TestRegisterAlreadyRegisteredQuery(t *testing.T) {
	h := func() query.Handler {
		return func(ctx context.Context, args argument.Arguments) (interface{}, error) {
			return nil, nil
		}
	}
	qr := query.NewRegisteredQuery("test", h())

	_, err := query.NewRegistry(qr, qr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name already registered")
}

func TestHandleUnregisteredQuery(t *testing.T) {
	h := func() query.Handler {
		return func(ctx context.Context, args argument.Arguments) (interface{}, error) {
			return nil, nil
		}
	}
	qr := query.NewRegisteredQuery("foo", h())

	r, _ := query.NewRegistry(qr)

	q := query.New("bar", argument.Arguments{})
	_, err := r.Handle(context.TODO(), q)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query not registered: bar")
}
