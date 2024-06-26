package firevault

// options used by validator
type validationOpts struct {
	skipValidation       bool
	skipRequired         bool
	allowOmitEmptyUpdate bool
	allowEmptyFields     []string
}

// Firevault Options allows for setting options for
// validation, creation and updating methods.
type Options struct {
	// Skip all validations. Default is "false".
	skipValidation bool
	// Skip the "required" tag during creation method.
	skipRequired bool
	// Honour the "required" tag during validation and
	// updating methods.
	unskipRequired bool
	// Specify which field paths should ignore the "omitempty" and
	// "omitemptyupdate" tags.
	//
	// This can be useful when zero values are needed only during
	// a specific method call.
	//
	// If left empty, those tags will be honoured for all fields.
	allowEmptyFields []FieldPath
	// Specify which field paths to be overwritten. Other fields
	// on the existing document will be untouched.
	//
	// It is an error if a provided field path does not refer to a
	// value in the data passed.
	//
	// If left empty, all the field paths given in the data argument
	// will be overwritten. Only used for updating method.
	mergeFields []FieldPath
	// Specify custom doc ID. If left empty, Firestore will
	// automatically create one. Only used for creation method.
	id string
}

// Create a new Options instance.
func NewOptions() Options {
	return Options{}
}

// Skip all validations - the "name" tag, "omitempty" tags and
// "ignore" tag will still be honoured.
func (o Options) SkipValidation() Options {
	o.skipValidation = true
	return o
}

// Ignore the "required" tag during validation.
//
// Only useful during creation method, as
// the default behaviour is to not skip it.
func (o Options) SkipRequired() Options {
	o.skipRequired = true
	return o
}

// Honour the "required" tag during validation.
//
// Only useful during validation and updating methods,
// as the default behaviour is to skip it.
func (o Options) UnskipRequired() Options {
	o.unskipRequired = true
	return o
}

// Specify which field paths should ignore the "omitempty" and
// "omitemptyupdate" tags.
//
// This can be useful when zero values are needed only during
// a specific method call.
//
// If left empty, those tags will be honoured for all fields.
func (o Options) AllowEmptyFields(fields ...FieldPath) Options {
	o.allowEmptyFields = append(o.allowEmptyFields, fields...)
	return o
}

// Specify which field paths to be overwritten. Other fields
// on the existing document will be untouched.
//
// It is an error if a provided field path does not refer to a
// value in the data passed.
//
// Only used for updating method.
func (o Options) MergeFields(fields ...FieldPath) Options {
	o.mergeFields = append(o.mergeFields, fields...)
	return o
}

// Specify custom doc ID. If left empty, Firestore will
// automatically create one.
//
// Only used for creation method.
func (o Options) ID(id string) Options {
	o.id = id
	return o
}

// FieldPath is a non-empty sequence of non-empty fields that reference
// a value.
//
// For example,
//
//	[]string{"a", "b"}
//
// is equivalent to the string form
//
//	"a.b"
//
// but
//
//	[]string{"*"}
//
// has no equivalent string form.
type FieldPath []string
