package firevault

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
)

type validator struct {
	validations     map[string]ValidationFn
	transformations map[string]TransformationFn
}

type ValidationFn func(ctx context.Context, path string, value reflect.Value, param string) (bool, error)

type TransformationFn func(ctx context.Context, path string, value reflect.Value) (interface{}, error)

type reflectedStruct struct {
	types  reflect.Type
	values reflect.Value
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
		return errors.New("firevault: validation function name cannot be empty")
	}

	if validation == nil {
		return fmt.Errorf("firevault: validation function %s cannot be empty", name)
	}

	v.validations[name] = validation
	return nil
}

func (v *validator) registerTransformation(name string, transformation TransformationFn) error {
	if len(name) == 0 {
		return errors.New("firevault: transformation function name cannot be empty")
	}

	if transformation == nil {
		return fmt.Errorf("firevault: transformation function %s cannot be empty", name)
	}

	v.transformations[name] = transformation
	return nil
}

func (v *validator) validate(
	ctx context.Context,
	data interface{},
	opts validationOpts,
) (map[string]interface{}, error) {
	rs := reflectedStruct{reflect.TypeOf(data), reflect.ValueOf(data)}

	if rs.values.Kind() != reflect.Pointer && rs.values.Kind() != reflect.Ptr {
		return nil, errors.New("firevault: data must be a pointer to a struct")
	}

	rs.values = rs.values.Elem()
	rs.types = rs.types.Elem()

	if rs.values.Kind() != reflect.Struct {
		return nil, errors.New("firevault: data must be a pointer to a struct")
	}

	dataMap, err := v.validateFields(ctx, rs, "", opts)
	return dataMap, err
}

func (v *validator) validateFields(
	ctx context.Context,
	rs reflectedStruct,
	path string,
	opts validationOpts,
) (map[string]interface{}, error) {
	// map which will hold all fields to pass to firestore
	dataMap := make(map[string]interface{})

	// iterate over struct fields
	for i := 0; i < rs.values.NumField(); i++ {
		fieldValue := rs.values.Field(i)
		fieldType := rs.types.Field(i)
		fieldName := fieldType.Name

		tag := fieldType.Tag.Get("firevault")

		if tag == "" || tag == "-" {
			continue
		}

		rules := v.parseTag(tag)

		// get field path based on name tag and trim leading dot (if exists)
		fieldPath := path + "." + rules[0]
		fieldPath = strings.TrimPrefix(fieldPath, ".")

		// check if field is of supported type and return error if not
		if !isSupported(fieldValue) {
			return nil, errors.New("firevault: unsupported field type - " + fieldPath)
		}

		// skip validation if value is zero and an omitempty tag is present
		// unless tags are skipped using options
		omitEmptyMethodTag := string("omitempty_" + opts.method)
		shouldOmitEmpty := slices.Contains(rules, "omitempty") || slices.Contains(rules, omitEmptyMethodTag)

		if shouldOmitEmpty {
			if !slices.Contains(opts.emptyFieldsAllowed, fieldPath) {
				if !hasValue(fieldValue) {
					continue
				}
			}
		}

		// remove omitempty tags from rules, so no validation is attempted
		rules = delSliceItem(rules, "omitempty")
		rules = delSliceItem(rules, string("omitempty_"+create))
		rules = delSliceItem(rules, string("omitempty_"+update))
		rules = delSliceItem(rules, string("omitempty_"+validate))

		// get pointer value, only if it's not nil
		if fieldValue.Kind() == reflect.Pointer || fieldValue.Kind() == reflect.Ptr {
			if !fieldValue.IsNil() {
				fieldValue = fieldValue.Elem()
			}
		}

		// validate field based on rules
		for ruleIndex, rule := range rules {

			// use first tag rule as new field name, rather than having a "name=" prefix
			if ruleIndex == 0 && rule != "" {
				fieldName = rule
				continue
			}

			// skip validations depending on the passed in options
			if opts.skipValidation {
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

			if strings.HasPrefix(rule, "transform=") {
				// skip rule if value is zero
				if !hasValue(fieldValue) {
					continue
				}

				transName := strings.TrimPrefix(rule, "transform=")

				if transformation, ok := v.transformations[transName]; ok {
					newValue, err := transformation(ctx, fieldPath, fieldValue)
					if err != nil {
						return nil, err
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
				requiredMethodTag := string("required_" + opts.method)
				if rule != "required" && rule != requiredMethodTag && !hasValue(fieldValue) {
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
					ok, err := validation(ctx, fieldPath, fieldValue, param)
					if err != nil {
						return nil, err
					}
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

		finalValue, err := v.processFinalValue(ctx, fieldValue, fieldPath, opts)
		if err != nil {
			return nil, err
		}

		dataMap[fieldName] = finalValue
	}

	return dataMap, nil
}

func (v *validator) processFinalValue(
	ctx context.Context,
	fieldValue reflect.Value,
	fieldPath string,
	opts validationOpts,
) (interface{}, error) {
	switch fieldValue.Kind() {
	case reflect.Struct:
		return v.processStructValue(ctx, fieldValue, fieldPath, opts)
	case reflect.Map:
		return v.processMapValue(ctx, fieldValue, fieldPath, opts)
	case reflect.Array, reflect.Slice:
		return v.processSliceValue(ctx, fieldValue, fieldPath, opts)
	default:
		return fieldValue.Interface(), nil
	}
}

func (v *validator) processStructValue(
	ctx context.Context,
	fieldValue reflect.Value,
	fieldPath string,
	opts validationOpts,
) (interface{}, error) {
	if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
		return fieldValue.Interface().(time.Time), nil
	}

	return v.validateFields(
		ctx,
		reflectedStruct{fieldValue.Type(), fieldValue},
		fieldPath,
		opts,
	)
}

func (v *validator) processMapValue(
	ctx context.Context,
	fieldValue reflect.Value,
	fieldPath string,
	opts validationOpts,
) (interface{}, error) {
	newMap := make(map[string]interface{})
	iter := fieldValue.MapRange()

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newFieldPath := fmt.Sprintf("%s.%v", fieldPath, key.Interface())

		processedValue, err := v.processFinalValue(ctx, val, newFieldPath, opts)
		if err != nil {
			return nil, err
		}

		newMap[key.String()] = processedValue
	}

	return newMap, nil
}

func (v *validator) processSliceValue(
	ctx context.Context,
	fieldValue reflect.Value,
	fieldPath string,
	opts validationOpts,
) (interface{}, error) {
	newSlice := make([]interface{}, fieldValue.Len())

	for i := 0; i < fieldValue.Len(); i++ {
		val := fieldValue.Index(i)
		newFieldPath := fmt.Sprintf("%s[%d]", fieldPath, i)

		processedValue, err := v.processFinalValue(ctx, val, newFieldPath, opts)
		if err != nil {
			return nil, err
		}

		newSlice[i] = processedValue
	}

	return newSlice, nil
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
