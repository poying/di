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
		make(map[reflect.Type]reflect.Value),
		make(map[reflect.Type][]reflect.Type),
		make(map[reflect.Type]reflect.Value),
	}
}

type Injector interface {
	Register(instance interface{}, factory interface{}) error
	Get(instance interface{}) error
}

type injector struct {
	*sync.RWMutex
	factories map[reflect.Type]reflect.Value
	services  map[reflect.Type][]reflect.Type
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
		depTyps, ok := i.services[typ]
		if !ok {
			return ErrNotRegistered
		}
		deps := make([]reflect.Value, 0)
		for _, depType := range depTyps {
			dep := reflect.New(depType.Elem())
			err := i.get(depType, dep)
			if err != nil {
				return err
			}
			deps = append(deps, dep)
		}
		factory := i.factories[typ]
		vals := factory.Call(deps)
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

	depsCount := factoryType.NumIn()
	deps := make([]reflect.Type, depsCount)
	for i := 0; i < factoryType.NumIn(); i++ {
		deps[i] = factoryType.In(i)
	}

	i.Lock()
	defer i.Unlock()
	_, ok := i.services[typ]
	if ok {
		return ErrDuplicate
	}
	i.services[typ] = deps
	i.factories[typ] = reflect.ValueOf(factory)

	return nil
}
