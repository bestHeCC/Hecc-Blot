package ioc

import (
	"fmt"
	"reflect"
)

var instanceValues = make(map[reflect.Type]map[string]reflect.Value)

const (
	instanceIsNotPtr       = "ioc: 注入实例必须是指针"
	invalidTypeFormat      = "ioc: 无效类型(Name = %s, Type = %v)"
	notImplementsFormat    = "ioc: %v没有实现%v"
	notInterfaceTypeFormat = "ioc: 非接口类型(%v)"
)

func Get(interfaceObj interface{}, name string) interface{} {
	return getValueWithName(interfaceObj, name).Interface()
}

func Inject(instance interface{}) {
	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() != reflect.Ptr {
		panic(instanceIsNotPtr)
	}

	inject(instanceValue)
}

func Set(interfaceObj interface{}, instance interface{}) {
	SetWithName(interfaceObj, "", instance)
}

func SetWithName(interfaceObj interface{}, name string, instance interface{}) {
	interfaceType := getInterfaceType(interfaceObj)
	instanceType := reflect.TypeOf(instance)
	if !instanceType.Implements(interfaceType) {
		panic(
			fmt.Errorf(notImplementsFormat, instance, interfaceType),
		)
	}

	if _, ok := instanceValues[interfaceType]; !ok {
		instanceValues[interfaceType] = make(map[string]reflect.Value)
	}

	instanceValues[interfaceType][name] = reflect.ValueOf(instance)
}

func getInterfaceType(interfaceObj interface{}) reflect.Type {
	var interfaceType reflect.Type
	var ok bool
	if interfaceType, ok = interfaceObj.(reflect.Type); !ok {
		interfaceType = reflect.TypeOf(interfaceObj)
	}

	if interfaceType.Kind() == reflect.Ptr {
		interfaceType = interfaceType.Elem()
	}

	if interfaceType.Kind() != reflect.Interface {
		panic(
			fmt.Errorf(notInterfaceTypeFormat, interfaceType),
		)
	}

	return interfaceType
}

func getValueWithName(interfaceObj interface{}, name string) reflect.Value {
	interfaceType := getInterfaceType(interfaceObj)
	if values, ok := instanceValues[interfaceType]; ok {
		if v, ok := values[name]; ok {
			return v
		}
	}

	panic(
		fmt.Errorf(invalidTypeFormat, name, interfaceType),
	)
}

func inject(instanceValue reflect.Value) {
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}

	instanceType := instanceValue.Type()
	for j := 0; j < instanceType.NumField(); j++ {
		field := instanceValue.Type().Field(j)
		fieldValue := instanceValue.FieldByIndex(field.Index)
		if field.Anonymous {
			if field.Type.Kind() == reflect.Struct {
				inject(fieldValue)
			}
			continue
		}

		name, ok := field.Tag.Lookup("inject")
		if !ok {
			return
		}

		if fieldValue.Kind() == reflect.Ptr {
			value := reflect.New(
				field.Type.Elem(),
			)
			fieldValue.Set(value)
			fieldValue = fieldValue.Elem()
		}

		v := getValueWithName(field.Type, name)
		fieldValue.Set(v)
	}

}
