package firevault

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

// A Firevault Collection holds a reference to a Firestore
// CollectionRef.
type Collection[T interface{}] struct {
	connection *Connection
	ref        *firestore.CollectionRef
}

// A Document holds the ID and data related to fetched document.
type Document[T interface{}] struct {
	ID   string
	Data T
}

type ValidationOptions struct {
	// Skip all validations - the "name" tag, "omitempty" tags and
	// "ignore" tag will still be honoured. Default is "false".
	SkipValidation bool
	// Ignore the "required" tag during validation. Default is "true".
	SkipRequired bool
	// Specify which field paths should ignore the "omitempty" and
	// "omitemptyupdate" tags.
	//
	// This can be useful when zero values are needed only during
	// a specific method call.
	//
	// If left empty, those tags will be honoured for all fields.
	AllowEmptyFields []FieldPath
}

type CreationOptions struct {
	// Skip all validations - the "name", "omitempty" and
	// "ignore" tags will still be honoured. Default is "false".
	SkipValidation bool
	// Ignore the "required" tag during validation. Default is "false".
	SkipRequired bool
	// Specify which field paths should ignore the "omitempty" tag.
	//
	// This can be useful when zero values are needed only during
	// a specific method call.
	//
	// If left empty, those tags will be honoured for all fields.
	AllowEmptyFields []FieldPath
	// Specify custom doc ID. If left empty, Firestore will
	// automatically create one.
	ID string
}

type UpdatingOptions struct {
	// Skip all validations - the "name" tag, "omitempty" tags and
	// "ignore" tag will still be honoured. Default is "false".
	SkipValidation bool
	// Ignore the "required" tag during validation. Default is "true".
	SkipRequired bool
	// Specify which field paths should ignore the "omitempty" and
	// "omitemptyupdate" tags.
	//
	// This can be useful when zero values are needed only during
	// a specific method call.
	//
	// If left empty, those tags will be honoured for all fields.
	AllowEmptyFields []FieldPath
	// Specify which field paths to be overwritten. Other fields
	// on the existing document will be untouched.
	//
	// It is an error if a provided field path does not refer to a
	// value in the data passed.
	//
	// If left empty, all the field paths given in the data argument
	// will be overwritten.
	MergeFields []FieldPath
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

// used to determine how to parse options
type methodType string

const (
	validate methodType = "validate"
	create   methodType = "create"
	update   methodType = "update"
)

// Create a new Collection instance.
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

// Validate provided data.
func (c *Collection[T]) Validate(data *T, opts ...ValidationOptions) error {
	options := validationOpts{false, true, true, make([]string, 0)}

	if len(opts) > 0 {
		options = c.getValidationOpts(validate, options, opts[0])
	}

	_, err := c.connection.validator.validate(data, options)
	return err
}

// Create a Firestore document with provided data (after validation).
func (c *Collection[T]) Create(ctx context.Context, data *T, opts ...CreationOptions) (string, error) {
	id := ""
	options := validationOpts{false, false, false, make([]string, 0)}

	if len(opts) > 0 {
		opts := opts[0]

		if opts.ID != "" {
			id = opts.ID
		}

		options = c.getValidationOpts(create, options, ValidationOptions{
			SkipValidation:   opts.SkipValidation,
			SkipRequired:     opts.SkipRequired,
			AllowEmptyFields: opts.AllowEmptyFields,
		})
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

// Update a Firestore document with provided ID and data
// (after validation).
func (c *Collection[T]) UpdateById(ctx context.Context, id string, data *T, opts ...UpdatingOptions) error {
	options := validationOpts{false, true, true, make([]string, 0)}
	mergeFields := firestore.MergeAll

	if len(opts) > 0 {
		opts := opts[0]

		if len(opts.MergeFields) > 0 {
			fps := make([]firestore.FieldPath, 0)

			for i := 0; i < len(opts.MergeFields); i++ {
				fp := firestore.FieldPath(opts.MergeFields[i])
				fps = append(fps, fp)
			}

			mergeFields = firestore.Merge(fps...)
		}

		options = c.getValidationOpts(update, options, ValidationOptions{
			SkipValidation:   opts.SkipValidation,
			SkipRequired:     opts.SkipRequired,
			AllowEmptyFields: opts.AllowEmptyFields,
		})
	}

	dataMap, err := c.connection.validator.validate(data, options)
	if err != nil {
		return err
	}

	_, err = c.ref.Doc(id).Set(ctx, dataMap, mergeFields)
	return err
}

// Delete a Firestore document with provided ID.
func (c *Collection[T]) DeleteById(ctx context.Context, id string) error {
	_, err := c.ref.Doc(id).Delete(ctx)
	return err
}

// Find a Firestore document with provided ID.
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

// Create a new instance of a Firevault Query.
func (c *Collection[T]) Query() Query {
	return newQuery()
}

// Find all Firestore documents which match provided Query.
func (c *Collection[T]) Find(ctx context.Context, query Query) ([]Document[T], error) {
	builtQuery := c.buildQuery(query)
	iter := builtQuery.Documents(ctx)

	var docs []Document[T]

	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var doc T

		err = docSnap.DataTo(&doc)
		if err != nil {
			return nil, err
		}

		docs = append(docs, Document[T]{docSnap.Ref.ID, doc})
	}

	return docs, nil
}

// Find number of Firestore documents which match provided Query.
func (c *Collection[T]) Count(ctx context.Context, query Query) (int64, error) {
	builtQuery := c.buildQuery(query)
	results, err := builtQuery.NewAggregationQuery().WithCount("all").Get(ctx)
	if err != nil {
		return 0, err
	}

	count, ok := results["all"]
	if !ok {
		return 0, errors.New("firestore: couldn't get alias for COUNT from results")
	}

	countValue := count.(*firestorepb.Value)
	countInt := countValue.GetIntegerValue()

	return countInt, nil
}

// extract passed validation options
func (c *Collection[T]) getValidationOpts(method methodType, curOpts validationOpts, passedOpts ValidationOptions) validationOpts {
	if passedOpts.SkipValidation == !curOpts.skipValidation {
		curOpts.skipValidation = passedOpts.SkipValidation
	}

	if method == validate || method == update {
		if !passedOpts.SkipRequired {
			curOpts.skipRequired = !passedOpts.SkipRequired
		}
	} else {
		if passedOpts.SkipRequired {
			curOpts.skipRequired = passedOpts.SkipRequired
		}
	}

	if len(passedOpts.AllowEmptyFields) > 0 {
		for i := 0; i < len(passedOpts.AllowEmptyFields); i++ {
			fieldPath := ""

			for x := 0; x < len(passedOpts.AllowEmptyFields[i]); x++ {
				fieldPath = fmt.Sprintf("%s.%s", fieldPath, passedOpts.AllowEmptyFields[i][x])
			}

			curOpts.allowEmptyField = append(curOpts.allowEmptyField, fieldPath)
		}
	}

	return curOpts
}

// build a new query
func (c *Collection[T]) buildQuery(query Query) firestore.Query {
	newQuery := c.ref.Query

	for _, filter := range query.filters {
		newQuery = newQuery.Where(filter.path, filter.operator, filter.value)
	}

	for _, order := range query.orders {
		newQuery = newQuery.OrderBy(order.path, firestore.Direction(order.direction))
	}

	if len(query.startAt) > 0 {
		newQuery = newQuery.StartAt(query.startAt...)
	}

	if len(query.startAfter) > 0 {
		newQuery = newQuery.StartAfter(query.startAfter...)
	}

	if len(query.endBefore) > 0 {
		newQuery = newQuery.EndBefore(query.endBefore...)
	}

	if len(query.endAt) > 0 {
		newQuery = newQuery.EndAt(query.endAt...)
	}

	if query.limit > 0 {
		newQuery = newQuery.Limit(query.limit)
	}

	if query.limitToLast > 0 {
		newQuery = newQuery.LimitToLast(query.limitToLast)
	}

	if query.offset > 0 {
		newQuery = newQuery.Offset(query.offset)
	}

	return newQuery
}
