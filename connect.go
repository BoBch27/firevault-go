package firevault

import (
	"cloud.google.com/go/firestore"
)

// A Firevault Connection contains the Firestore Client
type Connection struct {
	client    *firestore.Client
	validator *validator
}

// Connect to Firevault
func Connect(client *firestore.Client) *Connection {
	val := newValidator()
	return &Connection{client, val}
}

// Register a new validation rule
func (c *Connection) RegisterValidation(name string, validation ValidationFn) error {
	return c.validator.registerValidation(name, validation)
}

// Register a new transformation rule
func (c *Connection) RegisterTransformation(name string, transformation TransformationFn) error {
	return c.validator.registerTransformation(name, transformation)
}
