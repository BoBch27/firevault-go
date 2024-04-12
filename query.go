package firevault

import (
	"cloud.google.com/go/firestore"
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
