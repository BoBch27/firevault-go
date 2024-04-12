package firevault

import (
	"context"

	"cloud.google.com/go/firestore"
)

type connection struct {
	ctx       context.Context
	client    *firestore.Client
	validator *validator
}

func Connect(ctx context.Context, client *firestore.Client) *connection {
	val := newValidator()
	return &connection{ctx, client, val}
}

func (c *connection) RegisterValidation(name string, validation ValidationFn) error {
	return c.validator.registerValidation(name, validation)
}

func (c *connection) RegisterTransformation(name string, transformation TransformationFn) error {
	return c.validator.registerTransformation(name, transformation)
}
