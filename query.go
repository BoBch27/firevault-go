package firevault

import (
	"errors"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

// A Firevault Query represents a Firestore Query
type Query[T interface{}] struct {
	connection *Connection
	query      *firestore.Query
}

// A Document holds the Id and data related to fetched document
type Document[T interface{}] struct {
	Id   string
	Data T
}

// Direction is the sort direction for result ordering
type Direction int32

// Asc sorts results from smallest to largest.
const Asc Direction = Direction(1)

// Desc sorts results from largest to smallest.
const Desc Direction = Direction(2)

func newQuery[T interface{}](connection *Connection, q firestore.Query) *Query[T] {
	return &Query[T]{connection, &q}
}

func (q *Query[T]) Where(path string, operator string, value interface{}) *Query[T] {
	return newQuery[T](q.connection, q.query.Where(path, operator, value))
}

func (q *Query[T]) OrderBy(path string, direction Direction) *Query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, firestore.Direction(direction)))
}

func (q *Query[T]) Limit(num int) *Query[T] {
	return newQuery[T](q.connection, q.query.Limit(num))
}

func (q *Query[T]) LimitToLast(num int) *Query[T] {
	return newQuery[T](q.connection, q.query.LimitToLast(num))
}

func (q *Query[T]) StartAt(path string, field interface{}) *Query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, firestore.Asc).StartAt(field))
}

func (q *Query[T]) StartAfter(path string, field interface{}) *Query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, firestore.Asc).StartAfter(field))
}

func (q *Query[T]) EndBefore(path string, field interface{}) *Query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, firestore.Asc).EndBefore(field))
}

func (q *Query[T]) EndAt(path string, field interface{}) *Query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, firestore.Asc).EndAt(field))
}

func (q *Query[T]) Offset(num int) *Query[T] {
	return newQuery[T](q.connection, q.query.Offset(num))
}

// Fetch documents based on query criteria
func (q *Query[T]) Fetch() ([]Document[T], error) {
	var doc T
	var docs []Document[T]

	iter := q.query.Documents(q.connection.ctx)

	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		err = docSnap.DataTo(&doc)
		if err != nil {
			return nil, err
		}

		docs = append(docs, Document[T]{docSnap.Ref.ID, doc})
	}

	return docs, nil
}

// Return document count for specified query criteria
func (q *Query[T]) Count() (int64, error) {
	results, err := q.query.NewAggregationQuery().WithCount("all").Get(q.connection.ctx)
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
