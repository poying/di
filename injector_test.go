package di_test

import (
	"errors"
	"testing"

	"github.com/poying/di"
	"github.com/stretchr/testify/assert"
)

type Service1 struct {
	Name string
}

type Service2 struct {
	*Service1
}

type Service3 struct{}

type Service4 struct{}

type Service5 struct{}

func TestRegister(t *testing.T) {
	t.Run("Without deps", func(t *testing.T) {
		injector := di.New()
		err := injector.Register(&Service1{}, func() (*Service1, error) {
			return &Service1{}, nil
		})
		assert.Nil(t, err)
	})

	t.Run("With deps", func(t *testing.T) {
		injector := di.New()
		err := injector.Register(&Service2{}, func(service1 *Service1) (*Service2, error) {
			return &Service2{}, nil
		})
		assert.Nil(t, err)
	})

	t.Run("Duplicate", func(t *testing.T) {
		injector := di.New()
		err := injector.Register(&Service1{}, func() (*Service1, error) {
			return &Service1{}, nil
		})
		assert.Nil(t, err)
		err = injector.Register(&Service1{}, func() (*Service1, error) {
			return &Service1{}, nil
		})
		assert.Equal(t, di.ErrDuplicate, err)
	})

	t.Run("Wrong function type", func(t *testing.T) {
		injector := di.New()
		err := injector.Register(&Service1{}, struct{}{})
		assert.NotNil(t, err)
	})

	t.Run("Wrong function type", func(t *testing.T) {
		injector := di.New()
		err := injector.Register(&Service1{}, func() (*Service2, error) {
			return &Service2{}, nil
		})
		assert.NotNil(t, err)
	})

	t.Run("Wrong function type", func(t *testing.T) {
		injector := di.New()
		err := injector.Register(&Service2{}, func() (*Service2, int) {
			return &Service2{}, 1
		})
		assert.NotNil(t, err)
	})

	t.Run("Wrong function type", func(t *testing.T) {
		injector := di.New()
		err := injector.Register(&Service2{}, func() *Service2 {
			return &Service2{}
		})
		assert.NotNil(t, err)
	})
}

func TestGet(t *testing.T) {
	setup := func(t *testing.T) di.Injector {
		injector := di.New()
		err := injector.Register(&Service1{}, func() (*Service1, error) {
			return &Service1{Name: "poying"}, nil
		})
		assert.Nil(t, err)
		err = injector.Register(&Service2{}, func(service1 *Service1) (*Service2, error) {
			return &Service2{service1}, nil
		})
		assert.Nil(t, err)
		err = injector.Register(&Service4{}, func(service3 *Service3) (*Service4, error) {
			return &Service4{}, nil
		})
		assert.Nil(t, err)
		err = injector.Register(&Service5{}, func() (*Service5, error) {
			return nil, errors.New("~")
		})
		assert.Nil(t, err)
		return injector
	}

	t.Run("Get non-pointer value", func(t *testing.T) {
		injector := setup(t)
		err := injector.Get(Service1{})
		assert.NotNil(t, err)
	})

	t.Run("Get registered service", func(t *testing.T) {
		injector := setup(t)
		service := &Service1{}
		err := injector.Get(service)
		assert.Nil(t, err)
		assert.Equal(t, "poying", service.Name)
	})

	t.Run("Get registered service", func(t *testing.T) {
		injector := setup(t)
		service := &Service2{}
		err := injector.Get(service)
		assert.Nil(t, err)
		assert.Equal(t, "poying", service.Name)
	})

	t.Run("Get registered service which is depends on non-registered service", func(t *testing.T) {
		injector := setup(t)
		service := &Service4{}
		err := injector.Get(service)
		assert.NotNil(t, err)
	})

	t.Run("Get not registered service", func(t *testing.T) {
		injector := setup(t)
		service := &Service3{}
		err := injector.Get(service)
		assert.Equal(t, di.ErrNotRegistered, err)
	})

	t.Run("When factory function return an error", func(t *testing.T) {
		injector := setup(t)
		service := &Service5{}
		err := injector.Get(service)
		assert.NotNil(t, err)
	})
}

func TestInjectF(t *testing.T) {
	t.Run("With non-function argument", func(t *testing.T) {
		injector := di.New()
		err := injector.InjectF(struct{}{})
		assert.NotNil(t, err)
		assert.EqualError(t, err, "The first argument is not a function")
	})

	t.Run("With wrong function type", func(t *testing.T) {
		injector := di.New()
		err := injector.InjectF(func() {})
		assert.NotNil(t, err)
		assert.EqualError(t, err, "The first argument has wrong type. It must return only 1 value")
	})

	t.Run("With wrong function type", func(t *testing.T) {
		injector := di.New()
		err := injector.InjectF(func() int { return 1 })
		assert.EqualError(t, err, "The second argument has wrong type. It must return error")
	})

	t.Run("Depends on non-registered service", func(t *testing.T) {
		injector := di.New()
		err := injector.InjectF(func(service *Service1) error { return nil })
		assert.Equal(t, di.ErrNotRegistered, err)
	})

	t.Run("Depends on non-registered service", func(t *testing.T) {
		injector := di.New()
		err := injector.InjectF(func() error { return errors.New("~") })
		assert.EqualError(t, err, "~")
	})
}
