package firevault

import (
	"cloud.google.com/go/firestore"
)

type collection[T interface{}] struct {
	connection *connection
	ref        *firestore.CollectionRef
}

type CreationOptions struct {
	ValidationOpts
	Id string
}

func NewCollection[T interface{}](connection *connection, name string) *collection[T] {
	return &collection[T]{
		connection,
		connection.client.Collection(name),
	}
}

func (c *collection[T]) Validate(data T, opts ...ValidationOpts) error {
	options := ValidationOpts{true}

	if len(opts) > 0 {
		options = opts[0]
	}

	_, err := c.connection.validator.validate(data, options)
	return err
}

func (c *collection[T]) Create(data T, opts ...CreationOptions) (string, error) {
	var id string
	valOptions := ValidationOpts{false}

	if len(opts) > 0 {
		id = opts[0].Id
		valOptions = ValidationOpts{opts[0].SkipRequired}
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

func (c *collection[T]) FindById(id string) (T, error) {
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

func (c *collection[T]) UpdateById(id string, data T, opts ...ValidationOpts) error {
	options := ValidationOpts{true}

	if len(opts) > 0 {
		options = opts[0]
	}

	dataMap, err := c.connection.validator.validate(data, options)
	if err != nil {
		return err
	}

	_, err = c.ref.Doc(id).Set(c.connection.ctx, dataMap, firestore.MergeAll)
	return err
}

func (c *collection[T]) DeleteById(id string) error {
	_, err := c.ref.Doc(id).Delete(c.connection.ctx)
	return err
}
