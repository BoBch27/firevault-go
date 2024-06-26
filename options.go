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
	// Specify which fields (using "dot notation") should ignore
	// the "omitempty" and "omitemptyupdate" tags.
	//
	// This can be useful when zero values are needed only during
	// a specific method call.
	//
	// If left empty, those tags will be honoured for all fields.
	allowEmptyFields []string
	// Specify which fields (using "dot notation") should ignore
	// the "omitempty" and "omitemptyupdate" tags.
	//
	// It is an error if a provided field path does not refer to a
	// value in the data passed.
	//
	// If left empty, all the field paths given in the data argument
	// will be overwritten. Only used for updating method.
	mergeFields []string
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

// Specify which field paths (using dot-separated strings)
// should ignore the "omitempty" and "omitemptyupdate" tags.
//
// This can be useful when zero values are needed only during
// a specific method call.
//
// If left empty, those tags will be honoured for all fields.
func (o Options) AllowEmptyFields(fields ...string) Options {
	o.allowEmptyFields = append(o.allowEmptyFields, fields...)
	return o
}

// Specify which field paths (using dot-separated strings)
// to be overwritten. Other fields on the existing document
// will be untouched.
//
// It is an error if a provided field path does not refer to a
// value in the data passed.
//
// Only used for updating method.
func (o Options) MergeFields(fields ...string) Options {
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
