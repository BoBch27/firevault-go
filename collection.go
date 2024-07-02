package firevault

import (
	"context"
	"errors"
	"strings"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

// A Firevault Collection allows for the fetching and
// modifying (with validation) of documents in a
// Firestore Collection.
type Collection[T interface{}] struct {
	connection *Connection
	ref        *firestore.CollectionRef
}

// A Firevault Document holds the ID and data related to
// fetched document.
type Document[T interface{}] struct {
	ID   string
	Data T
}

// Create a new Collection instance.
//
// A Firevault Collection allows for the fetching and
// modifying (with validation) of documents in a
// Firestore Collection.
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
func (c *Collection[T]) Validate(data *T, opts ...Options) error {
	valOptions, _, _ := c.parseOptions(validate, opts...)

	_, err := c.connection.validator.validate(data, valOptions)
	return err
}

// Create a Firestore document with provided data (after validation).
func (c *Collection[T]) Create(ctx context.Context, data *T, opts ...Options) (string, error) {
	valOptions, id, _ := c.parseOptions(create, opts...)

	dataMap, err := c.connection.validator.validate(data, valOptions)
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
func (c *Collection[T]) UpdateById(ctx context.Context, id string, data *T, opts ...Options) error {
	valOptions, _, mergeFields := c.parseOptions(update, opts...)

	dataMap, err := c.connection.validator.validate(data, valOptions)
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

// extract passed options
func (c *Collection[T]) parseOptions(method methodType, opts ...Options) (validationOpts, string, firestore.SetOption) {
	options := validationOpts{
		method:               method,
		skipValidation:       false,
		skipRequired:         false,
		allowOmitEmptyUpdate: false,
		allowEmptyFields:     make([]string, 0),
	}

	if method != create {
		options.skipRequired = true
		options.allowOmitEmptyUpdate = true
	}

	if len(opts) == 0 {
		return options, "", firestore.MergeAll
	}

	// parse options
	passedOpts := opts[0]

	if passedOpts.skipValidation {
		options.skipValidation = true
	}

	if method != create {
		if passedOpts.unskipRequired {
			options.skipRequired = false
		}
	} else {
		if passedOpts.skipRequired {
			options.skipRequired = true
		}
	}

	if len(passedOpts.allowEmptyFields) > 0 {
		options.allowEmptyFields = passedOpts.allowEmptyFields
	}

	if method == update && len(passedOpts.mergeFields) > 0 {
		fps := make([]firestore.FieldPath, 0)

		for i := 0; i < len(passedOpts.mergeFields); i++ {
			fp := firestore.FieldPath(strings.Split(passedOpts.mergeFields[i], "."))
			fps = append(fps, fp)
		}

		return options, passedOpts.id, firestore.Merge(fps...)
	}

	return options, passedOpts.id, firestore.MergeAll
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
