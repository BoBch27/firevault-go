package firevault

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
)

// A Firevault Collection holds a reference to a Firestore CollectionRef
type Collection[T interface{}] struct {
	connection *Connection
	ref        *firestore.CollectionRef
}

type ValidationOptions struct {
	SkipValidation bool
	SkipRequired   bool
	// Specify which field paths should ignore the "omitempty" and
	// "omitemptyupdate" tags.
	//
	// This can be useful when zero values are needed only during
	// a specific method call.
	//
	// If left empty, both tags will be honoured for all fields.
	AllowEmptyFields []FieldPath
}

type CreationOptions struct {
	ValidationOptions
	Id string
}

type UpdatingOptions struct {
	ValidationOptions
	// Specify which field paths to be overwritten. Other fields
	// on the existing document will be untouched.
	//
	// It is an error if a provided field path does not refer to a value
	// in the data passed.
	//
	// If left empty, all the field paths given in the data argument
	// will be overwritten.
	MergeFields []FieldPath
}

// FieldPath is a non-empty sequence of non-empty fields that reference a value.
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

// Create a new Collection instance
func NewCollection[T interface{}](connection *Connection, name string) (*Collection[T], error) {
	if name == "" {
		return nil, errors.New("firevault: collection name cannot be empty")
	}

	collection := &Collection[T]{
		connection,
		connection.client.Collection(name),
	}

	return collection, nil
}

// Validate provided data
func (c *Collection[T]) Validate(data *T, opts ...ValidationOptions) error {
	options := validationOpts{false, true, true, make([]string, 0)}

	if len(opts) > 0 {
		options.skipRequired = opts[0].SkipRequired

		if len(opts[0].AllowEmptyFields) > 0 {
			for i := 0; i < len(opts[0].AllowEmptyFields); i++ {
				fieldPath := ""

				for x := 0; x < len(opts[0].AllowEmptyFields[i]); x++ {
					fieldPath = fmt.Sprintf("%s.%s", fieldPath, opts[0].AllowEmptyFields[i][x])
				}

				options.allowEmptyField = append(options.allowEmptyField, fieldPath)
			}
		}
	}

	_, err := c.connection.validator.validate(data, options)
	return err
}

// Create a Firestore document with provided data (after validation)
func (c *Collection[T]) Create(ctx context.Context, data *T, opts ...CreationOptions) (string, error) {
	var id string
	options := validationOpts{false, false, false, make([]string, 0)}

	if len(opts) > 0 {
		id = opts[0].Id
		options.skipValidation = opts[0].SkipValidation
		options.skipRequired = opts[0].SkipRequired

		if len(opts[0].AllowEmptyFields) > 0 {
			for i := 0; i < len(opts[0].AllowEmptyFields); i++ {
				fieldPath := ""

				for x := 0; x < len(opts[0].AllowEmptyFields[i]); x++ {
					fieldPath = fmt.Sprintf("%s.%s", fieldPath, opts[0].AllowEmptyFields[i][x])
				}

				options.allowEmptyField = append(options.allowEmptyField, fieldPath)
			}
		}
	}

	dataMap, err := c.connection.validator.validate(data, options)
	if err != nil {
		return "", err
	}

	if id == "" {
		docRef, _, err := c.ref.Add(ctx, dataMap)
		if err != nil {
			return "", err
		}

		id = docRef.ID
	} else {
		_, err = c.ref.Doc(id).Set(ctx, dataMap)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

// Find a Firestore document with provided id
func (c *Collection[T]) FindById(ctx context.Context, id string) (T, error) {
	var doc T

	docSnap, err := c.ref.Doc(id).Get(ctx)
	if err != nil {
		return doc, err
	}

	err = docSnap.DataTo(&doc)
	if err != nil {
		return doc, err
	}

	return doc, err
}

// Create a new instance of a Firevault Query
func (c *Collection[T]) Find() *Query[T] {
	return newQuery[T](c.connection, c.ref.Query)
}

// Update a Firestore document with provided id and data (after validation)
func (c *Collection[T]) UpdateById(ctx context.Context, id string, data *T, opts ...UpdatingOptions) error {
	options := validationOpts{false, true, true, make([]string, 0)}
	mergeFields := firestore.MergeAll

	if len(opts) > 0 {
		options.skipValidation = opts[0].SkipValidation
		options.skipRequired = opts[0].SkipRequired

		if len(opts[0].MergeFields) > 0 {
			fps := make([]firestore.FieldPath, 0)

			for i := 0; i < len(opts[0].MergeFields); i++ {
				fp := firestore.FieldPath(opts[0].MergeFields[i])
				fps = append(fps, fp)
			}

			mergeFields = firestore.Merge(fps...)
		}

		if len(opts[0].AllowEmptyFields) > 0 {
			for i := 0; i < len(opts[0].AllowEmptyFields); i++ {
				fieldPath := ""

				for x := 0; x < len(opts[0].AllowEmptyFields[i]); x++ {
					fieldPath = fmt.Sprintf("%s.%s", fieldPath, opts[0].AllowEmptyFields[i][x])
				}

				options.allowEmptyField = append(options.allowEmptyField, fieldPath)
			}
		}
	}

	dataMap, err := c.connection.validator.validate(data, options)
	if err != nil {
		return err
	}

	_, err = c.ref.Doc(id).Set(ctx, dataMap, mergeFields)
	return err
}

// Delete a Firestore document with provided id
func (c *Collection[T]) DeleteById(ctx context.Context, id string) error {
	_, err := c.ref.Doc(id).Delete(ctx)
	return err
}
