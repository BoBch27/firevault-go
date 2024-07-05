package firevault

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"time"
)

var builtInValidators = map[string]ValidationFn{
	"required":          validateRequired,
	"required_create":   validateRequired,
	"required_update":   validateRequired,
	"required_validate": validateRequired,
	"email":             validateEmail,
	"max":               validateMax,
	"min":               validateMin,
}

// validates if field's value is not the default static value
func hasValue(fieldValue reflect.Value) bool {
	switch fieldValue.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !fieldValue.IsNil()
	default:
		return fieldValue.IsValid() && !fieldValue.IsZero()
	}
}

// validates if field is zero
func validateRequired(_ context.Context, _ string, fieldValue reflect.Value, _ string) bool {
	return hasValue(fieldValue)
}

// validates if field is a valid email address
func validateEmail(_ context.Context, _ string, fieldValue reflect.Value, _ string) bool {
	emailRegex := regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")
	return emailRegex.MatchString(fieldValue.String())
}

// validates if field's value is less than or equal to param's value
func validateMax(
	_ context.Context,
	fieldPath string,
	fieldValue reflect.Value,
	param string,
) bool {
	if param == "" {
		panic(fmt.Sprintf("firevault: provide a max param - %T", fieldPath))
	}

	switch fieldValue.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		return fieldValue.Len() <= int(asInt(param))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldValue.Int() <= asInt(param)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldValue.Uint() <= asUint(param)
	case reflect.Float32, reflect.Float64:
		return fieldValue.Float() <= asFloat(param)
	case reflect.Struct:
		timeType := reflect.TypeOf(time.Time{})

		if fieldValue.Type().ConvertibleTo(timeType) {
			max := asTime(param)
			t := fieldValue.Convert(timeType).Interface().(time.Time)

			return t.After(max)
		}
	}

	panic(fmt.Sprintf("firevault: bad field type - %T", fieldPath))
}

// validates if field's value is greater than or equal to param's value
func validateMin(
	_ context.Context,
	fieldPath string,
	fieldValue reflect.Value,
	param string,
) bool {
	if param == "" {
		panic(fmt.Sprintf("firevault: provide a min param - %T", fieldPath))
	}

	switch fieldValue.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		return fieldValue.Len() >= int(asInt(param))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldValue.Int() >= asInt(param)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldValue.Uint() >= asUint(param)
	case reflect.Float32, reflect.Float64:
		return fieldValue.Float() >= asFloat(param)
	case reflect.Struct:
		timeType := reflect.TypeOf(time.Time{})

		if fieldValue.Type().ConvertibleTo(timeType) {
			min := asTime(param)
			t := fieldValue.Convert(timeType).Interface().(time.Time)

			return t.Before(min)
		}
	}

	panic(fmt.Sprintf("firevault: bad field type - %T", fieldPath))
}
