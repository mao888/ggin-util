package gginutil

import (
	"fmt"
	"reflect"
)

func init() {
	register(&required{})
}

type required struct{}

func (r *required) check(name string, value reflect.Value) (bool, string) {
	if value.IsZero() {
		return false, fmt.Sprintf("%s failed is required.", name)
	}
	return true, EmptyString
}

func (r *required) name() string {
	return "required"
}
