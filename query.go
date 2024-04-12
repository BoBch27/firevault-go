package firevault

import (
	"errors"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

type query[T interface{}] struct {
	connection *connection
	query      *firestore.Query
}

type Document[T interface{}] struct {
	Id   string
	Data T
}

func newQuery[T interface{}](connection *connection, q firestore.Query) *query[T] {
	return &query[T]{connection, &q}
}

func (q *query[T]) Where(path string, operator string, value interface{}) *query[T] {
	return newQuery[T](q.connection, q.query.Where(path, operator, value))
}

func (q *query[T]) OrderBy(path string, direction firestore.Direction) *query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, direction))
}

func (q *query[T]) Limit(num int) *query[T] {
	return newQuery[T](q.connection, q.query.Limit(num))
}

func (q *query[T]) LimitToLast(num int) *query[T] {
	return newQuery[T](q.connection, q.query.LimitToLast(num))
}

func (q *query[T]) StartAfter(path string, field interface{}) *query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, firestore.Asc).StartAfter(field))
}

func (q *query[T]) EndBefore(path string, field interface{}) *query[T] {
	return newQuery[T](q.connection, q.query.OrderBy(path, firestore.Asc).EndBefore(field))
}

func (q *query[T]) Offset(num int) *query[T] {
	return newQuery[T](q.connection, q.query.Offset(num))
}

func (q *query[T]) Fetch() ([]Document[T], error) {
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

func (q *query[T]) Count() (int64, error) {
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
