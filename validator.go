package firevault

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type validator struct {
	validations     map[string]ValidationFn
	transformations map[string]TransformationFn
}

type ValidationFn func(reflect.Value, string) bool

type TransformationFn func(reflect.Value) (interface{}, error)

type reflectedStruct struct {
	types  reflect.Type
	values reflect.Value
}

type ValidationOpts struct {
	SkipRequired bool
}

func newValidator() *validator {
	validator := &validator{make(map[string]ValidationFn), make(map[string]TransformationFn)}

	// Register predefined validators
	for k, v := range builtInValidators {
		// no need to error check here, built in validations are always valid
		_ = validator.registerValidation(k, v)
	}

	return validator
}

func (v *validator) registerValidation(name string, validation ValidationFn) error {
	if len(name) == 0 {
		return errors.New("firevault: validation function Name cannot be empty")
	}

	if validation == nil {
		return fmt.Errorf("firevault: validation function %s cannot be empty", name)
	}

	v.validations[name] = validation
	return nil
}

func (v *validator) registerTransformation(name string, transformation TransformationFn) error {
	if len(name) == 0 {
		return errors.New("firevault: transformation function Name cannot be empty")
	}

	if transformation == nil {
		return fmt.Errorf("firevault: transformation function %s cannot be empty", name)
	}

	v.transformations[name] = transformation
	return nil
}

func (v *validator) validate(data interface{}, opts ValidationOpts) (map[string]interface{}, error) {
	rs := reflectedStruct{reflect.TypeOf(data), reflect.ValueOf(data)}

	if rs.values.Kind() != reflect.Pointer && rs.values.Kind() != reflect.Ptr {
		return nil, errors.New("firevault: data must be a pointer to a struct")
	}

	rs.values = rs.values.Elem()
	rs.types = rs.types.Elem()

	if rs.values.Kind() != reflect.Struct {
		return nil, errors.New("firevault: data must be a pointer to a struct")
	}

	dataMap, err := v.validateFields(rs, opts)
	return dataMap, err
}

func (v *validator) validateFields(rs reflectedStruct, opts ValidationOpts) (map[string]interface{}, error) {
	// map which will hold all fields to pass to firestore
	dataMap := make(map[string]interface{})

	// iterate over struct fields
	for i := 0; i < rs.values.NumField(); i++ {
		fieldValue := rs.values.Field(i)
		fieldType := rs.types.Field(i)
		fieldName := fieldType.Name

		// get pointer value
		if fieldValue.Kind() == reflect.Pointer || fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		tag := fieldType.Tag.Get("firevault")

		if tag == "" || tag == "-" {
			continue
		}

		rules := v.parseTag(tag)
		omitEmpty := slices.Contains(rules, "omitempty")

		// skip validation if value is zero and omitempty tag is present
		if omitEmpty {
			if fieldValue.IsZero() {
				continue
			} else {
				// remove omitempty from rules, so no validation is attempted
				index := slices.Index(rules, "omitempty")
				rules = slices.Delete(rules, index, index+1)
			}
		}

		// validate field based on rules
		for _, rule := range rules {
			// skip "required" rule depending on the passed in options
			if opts.SkipRequired && rule == "required" {
				continue
			}

			fe := &fieldError{
				code:        "",
				tag:         rule,
				field:       fieldName,
				structField: fieldType.Name,
				value:       fieldValue.Interface(),
				param:       "",
				kind:        fieldValue.Kind(),
				typ:         fieldValue.Type(),
			}

			if strings.HasPrefix(rule, "name=") {
				fieldName = strings.TrimPrefix(rule, "name=")
			} else if strings.HasPrefix(rule, "transform=") {
				// skip rule if value is zero
				if fieldValue.IsZero() {
					continue
				}

				transName := strings.TrimPrefix(rule, "transform=")

				if transformation, ok := v.transformations[transName]; ok {
					newValue, err := transformation(fieldValue)
					if err != nil {
						fe.code = "failed-transformation"
						return nil, fe
					}

					// check if rule returned a new value and assign it
					if newValue != nil {
						fieldValue = reflect.ValueOf(newValue)
						rs.values.Field(i).Set(fieldValue)
					}
				} else {
					fe.code = "unknown-transformation"
					return nil, fe
				}
			} else {
				// skip rules (apart from "required") if value is zero
				if rule != "required" && fieldValue.IsZero() {
					continue
				}

				// get param value if present
				param := ""
				params := strings.Split(rule, "=")
				if len(params) > 1 {
					param = params[1]
					rule = params[0]
				}

				if validation, ok := v.validations[rule]; ok {
					ok := validation(fieldValue, param)
					if !ok {
						fe.code = "failed-validation"
						fe.param = param
						return nil, fe
					}
				} else {
					fe.code = "unknown-validation"
					fe.param = param
					return nil, fe
				}
			}
		}

		// If the field is a nested struct, recursively validate it and add to map
		if fieldValue.Kind() == reflect.Struct {
			newMap, err := v.validateFields(reflectedStruct{fieldValue.Type(), fieldValue}, opts)
			if err != nil {
				return nil, err
			}

			dataMap[fieldName] = newMap
		} else {
			dataMap[fieldName] = fieldValue.Interface()
		}
	}

	return dataMap, nil
}

func (v *validator) parseTag(tag string) []string {
	rules := strings.Split(tag, ",")

	var validatedRules []string

	for _, rule := range rules {
		trimmedRule := strings.TrimSpace(rule)
		if trimmedRule != "" {
			validatedRules = append(validatedRules, trimmedRule)
		}
	}

	return validatedRules
}
