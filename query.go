package firevault

import (
	"cloud.google.com/go/firestore"
)

// A Firevault Query represents a Firestore Query.
type Query struct {
	query *firestore.Query
}

// Direction is the sort direction for result ordering.
type Direction int32

// Asc sorts results from smallest to largest.
const Asc Direction = Direction(1)

// Desc sorts results from largest to smallest.
const Desc Direction = Direction(2)

// DocumentID is the special field name representing the
// ID of a document in queries.
const DocumentID = "__name__"

// create a new query instance
func newQuery(query firestore.Query) *Query {
	return &Query{&query}
}

// Where returns a new Query that filters the set of results.
//
// A Query can have multiple filters. The path argument can be
// a single field or a dot-separated sequence of fields, and
// must not contain any of the runes "˜*/[]".
//
// The operator argument must be one of "==", "!=", "<", "<=",
// ">", ">=", "array-contains", "array-contains-any", "in" or
// "not-in".
func (q *Query) Where(path string, operator string, value interface{}) *Query {
	return newQuery(q.query.Where(path, operator, value))
}

// OrderBy returns a new Query that specifies the order in which
// results are returned. A Query can have multiple OrderBy
// specifications. It appends the specification to the list of
// existing ones.
func (q *Query) OrderBy(path string, direction Direction) *Query {
	return newQuery(q.query.OrderBy(path, firestore.Direction(direction)))
}

// Limit returns a new Query that specifies the maximum number of
// first results to return.
func (q *Query) Limit(num int) *Query {
	return newQuery(q.query.Limit(num))
}

// LimitToLast returns a new Query that specifies the maximum number
// of last results to return.
func (q *Query) LimitToLast(num int) *Query {
	return newQuery(q.query.LimitToLast(num))
}

// StartAt returns a new Query that specifies that results
// should start at the document with the given field values.
//
// StartAt should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q *Query) StartAt(path string, field interface{}) *Query {
	return newQuery(q.query.StartAt(field))
}

// StartAfter returns a new Query that specifies that results
// should start just after the document with the given field values.
//
// StartAfter should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q *Query) StartAfter(path string, field interface{}) *Query {
	return newQuery(q.query.StartAfter(field))
}

// EndBefore returns a new Query that specifies that results
// should end just before the document with the given field values.
//
// EndBefore should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q *Query) EndBefore(path string, field interface{}) *Query {
	return newQuery(q.query.EndBefore(field))
}

// EndBefore returns a new Query that specifies that results
// should end at the document with the given field values.
//
// EndBefore should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q *Query) EndAt(path string, field interface{}) *Query {
	return newQuery(q.query.EndAt(field))
}

// Offset returns a new Query that specifies the number of
// initial results to skip.
func (q *Query) Offset(num int) *Query {
	return newQuery(q.query.Offset(num))
}
