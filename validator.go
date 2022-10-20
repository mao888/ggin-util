package gginutil

import (
	"reflect"
	"sync"
)

const (
	validatorTag = "bt-validator"
)

var (
	checkMap = make(map[string]validator)
	lock     sync.Mutex
)

func register(v validator) {
	lock.Lock()
	checkMap[v.name()] = v
	lock.Unlock()
}

type validator interface {
	check(name string, value reflect.Value) (bool, string)
	name() string
}

func doValidator(voValue reflect.Value) []string {
	var errorMsgs []string
	if voValue.Type().Kind() == reflect.Ptr {
		voValue = voValue.Elem()
	}
	voType := voValue.Type()
	for i := 0; i < voValue.NumField(); i++ {
		fieldType, fieldValue := voType.Field(i), voValue.Field(i)
		switch fieldValue.Kind() {
		case reflect.Struct, reflect.Ptr:
			return append(errorMsgs, doValidator(fieldValue)...)
		case reflect.Slice, reflect.Array:
			count := fieldValue.Len()
			isStruct := false
			for i := 0; i < count; i++ {
				subFieldValue := fieldValue.Index(i)
				//处理结构体和指针结构体
				if subFieldValue.Kind() == reflect.Struct ||
					(subFieldValue.Kind() == reflect.Ptr &&
						subFieldValue.Elem().Kind() == reflect.Struct) {
					//如果不是结构体或者结构体指针，不递归
					isStruct = true
					errorMsgs = append(errorMsgs, doValidator(subFieldValue)...)
				} else {
					//不是结构体类型时，终止循环
					break
				}

			}
			if isStruct {
				return errorMsgs
			}
		}
		//目前只支持必须输入，之后需要支持多种，这里需要写循环处理
		tag := fieldType.Tag.Get(validatorTag)
		if tag == EmptyString {
			continue
		}
		//获取验证类
		if validator, ok := checkMap[tag]; !ok {
			continue
		} else {
			if ok, msg := validator.check(fieldType.Name, fieldValue); !ok {
				errorMsgs = append(errorMsgs, msg)
			}
		}
	}
	return errorMsgs
}
