package di

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

func New() Injector {
	return &injector{
		&sync.RWMutex{},
		make(map[reflect.Type]interface{}),
		make(map[reflect.Type]reflect.Value),
	}
}

type Injector interface {
	Register(instance interface{}, factory interface{}) error
	Get(instance interface{}) error
	InjectF(function interface{}) error
}

type injector struct {
	*sync.RWMutex
	factories map[reflect.Type]interface{}
	instances map[reflect.Type]reflect.Value
}

func (i *injector) Get(instance interface{}) error {
	i.Lock()
	defer i.Unlock()
	return i.get(reflect.TypeOf(instance), reflect.ValueOf(instance))
}

func (i *injector) get(typ reflect.Type, val reflect.Value) error {
	if typ.Kind() != reflect.Ptr {
		return errors.New("The first argument is not a pointer")
	}
	ins, ok := i.instances[typ]
	if !ok {
		factory, ok := i.factories[typ]
		if !ok {
			return ErrNotRegistered
		}
		vals, err := i.injectFunc(factory)
		if err != nil {
			return err
		}
		if vals[1].Interface() != nil {
			err := vals[1].Interface().(error)
			return err
		}
		ins = reflect.Indirect(vals[0])
		i.instances[typ] = ins
	}
	reflect.Indirect(val).Set(ins)
	return nil
}

func (i *injector) Register(instance interface{}, factory interface{}) error {
	typ := reflect.TypeOf(instance)
	factoryType := reflect.TypeOf(factory)

	if factoryType.Kind() != reflect.Func {
		return errors.New("The second argument is not a function")
	}
	if factoryType.NumOut() != 2 {
		return errors.New("The second argument has wrong type. It must return two values")
	}
	if factoryType.Out(1).String() != "error" || factoryType.Out(0) != typ {
		return fmt.Errorf("The second argument has wrong type. It must return (%s, error)", typ.String())
	}

	i.Lock()
	defer i.Unlock()
	_, ok := i.factories[typ]
	if ok {
		return ErrDuplicate
	}
	i.factories[typ] = factory

	return nil
}

func (i *injector) InjectF(function interface{}) error {
	functionType := reflect.TypeOf(function)

	if functionType.Kind() != reflect.Func {
		return errors.New("The first argument is not a function")
	}
	if functionType.NumOut() != 1 {
		return errors.New("The first argument has wrong type. It must return only 1 value")
	}
	if functionType.Out(0).String() != "error" {
		return errors.New("The second argument has wrong type. It must return error")
	}

	vals, err := i.injectFunc(function)
	if err != nil {
		return err
	}
	return vals[0].Interface().(error)
}

func (i *injector) injectFunc(function interface{}) ([]reflect.Value, error) {
	functionType := reflect.TypeOf(function)
	functionValue := reflect.ValueOf(function)
	depsCount := functionType.NumIn()
	deps := make([]reflect.Value, depsCount)

	for j := 0; j < depsCount; j++ {
		paramType := functionType.In(j)
		param := reflect.New(paramType.Elem())
		err := i.get(paramType, param)
		if err != nil {
			return nil, err
		}
		deps[j] = param
	}

	return functionValue.Call(deps), nil
}
