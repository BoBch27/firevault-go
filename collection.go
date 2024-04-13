package firevault

import (
	"errors"

	"cloud.google.com/go/firestore"
)

// A Firevault Collection holds a reference to a Firestore CollectionRef
type Collection[T interface{}] struct {
	connection *Connection
	ref        *firestore.CollectionRef
}

type ValidationOptions struct {
	SkipRequired bool
}

type CreationOptions struct {
	ValidationOptions
	Id string
}

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
	options := validationOpts{true, true}

	if len(opts) > 0 {
		options.skipRequired = opts[0].SkipRequired
	}

	_, err := c.connection.validator.validate(data, options)
	return err
}

// Create a Firestore document with provided data (after validation)
func (c *Collection[T]) Create(data *T, opts ...CreationOptions) (string, error) {
	var id string
	valOptions := validationOpts{false, false}

	if len(opts) > 0 {
		id = opts[0].Id
		valOptions.skipRequired = opts[0].SkipRequired
	}

	dataMap, err := c.connection.validator.validate(data, valOptions)
	if err != nil {
		return "", err
	}

	if id == "" {
		docRef, _, err := c.ref.Add(c.connection.ctx, dataMap)
		if err != nil {
			return "", err
		}

		id = docRef.ID
	} else {
		_, err = c.ref.Doc(id).Set(c.connection.ctx, dataMap)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

// Find a Firestore document with provided id
func (c *Collection[T]) FindById(id string) (T, error) {
	var doc T

	docSnap, err := c.ref.Doc(id).Get(c.connection.ctx)
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
func (c *Collection[T]) UpdateById(id string, data *T, opts ...ValidationOptions) error {
	options := validationOpts{true, true}

	if len(opts) > 0 {
		options.skipRequired = opts[0].SkipRequired
	}

	dataMap, err := c.connection.validator.validate(data, options)
	if err != nil {
		return err
	}

	_, err = c.ref.Doc(id).Set(c.connection.ctx, dataMap, firestore.MergeAll)
	return err
}

// Delete a Firestore document with provided id
func (c *Collection[T]) DeleteById(id string) error {
	_, err := c.ref.Doc(id).Delete(c.connection.ctx)
	return err
}
