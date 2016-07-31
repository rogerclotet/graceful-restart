package argument_test

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/rogerclotet/graceful-restart/cqrs/argument"
	"github.com/stretchr/testify/assert"
)

func TestIntArgument(t *testing.T) {
	value := rand.Int()
	arg := argument.New(value)

	returned, err := arg.Int()
	assert.NoError(t, err)
	assert.Equal(t, value, returned)
}

func TestInvalidIntArgumentReturnsError(t *testing.T) {
	invalidValues := []interface{}{
		"foo",
		12.32,
		[]int{1, 2, 3},
		func() int { return 3 },
	}

	for _, value := range invalidValues {
		arg := argument.New(value)
		_, err := arg.Int()

		assert.Error(t, err)
	}
}

func TestStringArgument(t *testing.T) {
	value := strconv.Itoa(rand.Int())
	arg := argument.New(value)

	returned, err := arg.String()
	assert.NoError(t, err)
	assert.Equal(t, value, returned)
}

func TestInvalidStringArgumentReturnsError(t *testing.T) {
	invalidValues := []interface{}{
		12,
		12.32,
		[]string{"foo", "bar"},
		func() string { return "foo" },
	}

	for _, value := range invalidValues {
		arg := argument.New(value)
		_, err := arg.String()

		assert.Error(t, err)
	}
}

func TestArguments(t *testing.T) {
	fooArgument := argument.New(3)
	barArgument := argument.New("baz")
	arguments := argument.Arguments{
		"foo": fooArgument,
		"bar": barArgument,
	}

	returnedFooArgument, err := arguments.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, fooArgument, returnedFooArgument)

	returnedBarArgument, err := arguments.Get("bar")
	assert.NoError(t, err)
	assert.Equal(t, barArgument, returnedBarArgument)
}

func TestArgumentTypedReturns(t *testing.T) {
	arguments := argument.Arguments{
		"foo": argument.New(3),
		"bar": argument.New("baz"),
	}

	fooValue, err := arguments.GetInt("foo")
	assert.NoError(t, err)
	assert.Equal(t, 3, fooValue)

	_, err = arguments.GetString("foo")
	assert.Error(t, err)

	barValue, err := arguments.GetString("bar")
	assert.NoError(t, err)
	assert.Equal(t, "baz", barValue)

	_, err = arguments.GetInt("bar")
	assert.Error(t, err)
}
