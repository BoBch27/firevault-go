package firevault

import (
	"fmt"
	"reflect"
)

var builtInValidators = map[string]ValidationFn{
	"required": validateRequired,
}

// validates if field is zero and returns error if so
func validateRequired(fieldName string, fieldValue reflect.Value, _ string) error {
	if fieldValue.IsZero() {
		return fmt.Errorf("firevault: field %s is required", fieldName)
	}

	return nil
}
