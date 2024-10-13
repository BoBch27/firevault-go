package firevault

import (
	"context"

	"cloud.google.com/go/firestore"
)

// A Firevault Connection provides access to
// Firevault services.
type Connection struct {
	client    *firestore.Client
	validator *validator
}

// Create a new Connection instance.
//
// A Firevault Connection provides access to
// Firevault services.
func Connect(ctx context.Context, projectID string) (*Connection, error) {
	val := newValidator()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &Connection{client, val}, nil
}

// Close closes the connection to Firestore.
//
// Should be invoked when the connection is
// no longer required.
//
// Close need not be called at program exit.
func (c *Connection) Close() error {
	return c.client.Close()
}

// Register a new validation rule.
func (c *Connection) RegisterValidation(name string, validation ValidationFn) error {
	return c.validator.registerValidation(name, validation)
}

// Register a new transformation rule.
func (c *Connection) RegisterTransformation(name string, transformation TransformationFn) error {
	return c.validator.registerTransformation(name, transformation)
}
