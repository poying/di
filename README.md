di
[![GoDoc](https://godoc.org/github.com/poying/di?status.svg)](http://godoc.org/github.com/poying/di)
[![Build](https://travis-ci.org/poying/di.svg?branch=master)](https://travis-ci.org/poying/di)
[![Coverage](https://codecov.io/gh/poying/di/branch/master/graph/badge.svg)](https://codecov.io/gh/poying/di)
==

A really simple di library in about 100 lines of code.

```go
import di

type Service1 struct {}

type Service2 struct {
    *Service1
}

di.Register(&Service1{}, func() (*Service1, error) {
    return &Service1{}
})

di.Register(&Service2{}, func(service1 *Service1) (*Service2, error) {
    return &Service2{service1}
})

service2 := &Service2{}
di.Get(service2)
```