package di_test

import (
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

type Service3 struct {
}

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
		return injector
	}

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

	t.Run("Get not registered service", func(t *testing.T) {
		injector := setup(t)
		service := &Service3{}
		err := injector.Get(service)
		assert.Equal(t, di.ErrNotRegistered, err)
	})
}