package firevault

// A Firevault Query helps to filter and order
// Firestore documents.
//
// Query values are immutable. Each Query method creates
// a new Query - it does not modify the old.
type Query struct {
	filters     []filter
	orders      []order
	startAt     []interface{}
	startAfter  []interface{}
	endBefore   []interface{}
	endAt       []interface{}
	limit       int
	limitToLast int
	offset      int
}

// represents a single filter in a Query
type filter struct {
	path     string
	operator string
	value    interface{}
}

// represents a single order in a Query
type order struct {
	path      string
	direction Direction
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

// Create a new Query instance.
func NewQuery() Query {
	return Query{}
}

// Where returns a new Query that filters the set of results.
// A Query can have multiple filters.
//
// The path argument can be asingle field or a dot-separated
// sequence of fields, and must not contain any of
// the runes "˜*/[]".
//
// The operator argument must be one of "==", "!=", "<", "<=",
// ">", ">=", "array-contains", "array-contains-any", "in" or
// "not-in".
func (q Query) Where(path string, operator string, value interface{}) Query {
	q.filters = append(q.filters, filter{path, operator, value})
	return q
}

// OrderBy returns a new Query that specifies the order in which
// results are returned. A Query can have multiple OrderBy
// specifications. It appends the specification to the list of
// existing ones.
func (q Query) OrderBy(path string, direction Direction) Query {
	q.orders = append(q.orders, order{path, direction})
	return q
}

// StartAt returns a new Query that specifies that results
// should start at the document with the given field values.
//
// StartAt should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q Query) StartAt(value ...interface{}) Query {
	q.startAt = value
	return q
}

// StartAfter returns a new Query that specifies that results
// should start just after the document with the given field values.
//
// StartAfter should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q Query) StartAfter(value ...interface{}) Query {
	q.startAfter = value
	return q
}

// EndBefore returns a new Query that specifies that results
// should end just before the document with the given field values.
//
// EndBefore should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q Query) EndBefore(value ...interface{}) Query {
	q.endBefore = value
	return q
}

// EndAt returns a new Query that specifies that results
// should end at the document with the given field values.
//
// EndAt should be called with one field value for each
// OrderBy clause, in the order that they appear.
func (q Query) EndAt(value ...interface{}) Query {
	q.endAt = value
	return q
}

// Limit returns a new Query that specifies the maximum number of
// first results to return.
func (q Query) Limit(num int) Query {
	q.limit = num
	return q
}

// LimitToLast returns a new Query that specifies the maximum number
// of last results to return.
func (q Query) LimitToLast(num int) Query {
	q.limitToLast = num
	return q
}

// Offset returns a new Query that specifies the number of
// initial results to skip.
func (q Query) Offset(num int) Query {
	q.offset = num
	return q
}
