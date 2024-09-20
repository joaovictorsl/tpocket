package torrent

import (
	"fmt"
	"reflect"
)

func getField[T any](field string, source map[string]interface{}) (T, error) {
	var zero T
	iField, ok := source[field]
	if !ok {
		return zero, ErrFieldMissing
	}

	fieldValue, ok := iField.(T)
	if !ok {
		return zero, fmt.Errorf("%s is not a %v, it is a %v", field, reflect.TypeOf(zero), reflect.TypeOf(iField))
	}

	return fieldValue, nil
}
