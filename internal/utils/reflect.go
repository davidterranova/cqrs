package utils

import (
	"reflect"
	"strings"
)

func GetTypeName(event interface{}) (reflect.Type, string) {
	rawType := reflect.TypeOf(event)
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}

	name := rawType.String()
	parts := strings.Split(name, ".")
	return rawType, parts[1]
}
