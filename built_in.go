package firevault

import (
	"fmt"
	"reflect"
	"regexp"
	"time"
)

var builtInValidators = map[string]ValidationFn{
	"required": validateRequired,
	"email":    validateEmail,
	"max":      validateMax,
	"min":      validateMin,
}

// validates if field is zero and returns error if so
func validateRequired(fieldName string, fieldValue reflect.Value, _ string) error {
	if fieldValue.IsZero() {
		return fmt.Errorf("firevault: field %s is required", fieldName)
	}

	return nil
}

// validates if field is a valid email address
func validateEmail(fieldName string, fieldValue reflect.Value, _ string) error {
	emailRegex := regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")
	if !emailRegex.MatchString(fieldValue.String()) {
		return fmt.Errorf("firevault: field %s must be a valid email address", fieldName)
	}

	return nil
}

// validates if field's value is less than or equal to param's value
func validateMax(fieldName string, fieldValue reflect.Value, param string) error {
	if param == "" {
		return fmt.Errorf("firevault: provide a max param for field %s", fieldName)
	}

	switch fieldValue.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		ok := fieldValue.Len() <= int(asInt(param))
		if !ok {
			return fmt.Errorf("firevault: field %s's max length is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ok := fieldValue.Int() <= asInt(param)
		if !ok {
			return fmt.Errorf("firevault: field %s's max value is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ok := fieldValue.Uint() <= asUint(param)
		if !ok {
			return fmt.Errorf("firevault: field %s's max value is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Float32, reflect.Float64:
		ok := fieldValue.Float() <= asFloat(param)
		if !ok {
			return fmt.Errorf("firevault: field %s's max value is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Struct:
		timeType := reflect.TypeOf(time.Time{})

		if fieldValue.Type().ConvertibleTo(timeType) {
			max := asTime(param)
			t := fieldValue.Convert(timeType).Interface().(time.Time)

			if t.After(max) {
				return fmt.Errorf("firevault: field %s's max time is %s", fieldName, param)
			} else {
				return nil
			}
		}
	}

	return fmt.Errorf("firevault: bad field type for field %s", fieldName)
}

// validates if field's value is greater than or equal to param's value
func validateMin(fieldName string, fieldValue reflect.Value, param string) error {
	if param == "" {
		return fmt.Errorf("firevault: provide a min param for field %s", fieldName)
	}

	switch fieldValue.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		ok := fieldValue.Len() >= int(asInt(param))
		if !ok {
			return fmt.Errorf("firevault: field %s's min length is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ok := fieldValue.Int() >= asInt(param)
		if !ok {
			return fmt.Errorf("firevault: field %s's min value is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ok := fieldValue.Uint() >= asUint(param)
		if !ok {
			return fmt.Errorf("firevault: field %s's min value is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Float32, reflect.Float64:
		ok := fieldValue.Float() >= asFloat(param)
		if !ok {
			return fmt.Errorf("firevault: field %s's min value is %s", fieldName, param)
		} else {
			return nil
		}
	case reflect.Struct:
		timeType := reflect.TypeOf(time.Time{})

		if fieldValue.Type().ConvertibleTo(timeType) {
			min := asTime(param)
			t := fieldValue.Convert(timeType).Interface().(time.Time)

			if t.Before(min) {
				return fmt.Errorf("firevault: field %s's min time is %s", fieldName, param)
			} else {
				return nil
			}
		}
	}

	return fmt.Errorf("firevault: bad field type for field %s", fieldName)
}
