package structures

import "errors"

type node[T any] struct {
	value T
	next  *node[T]
	prev  *node[T]
}

type Queue[T any] struct {
	Length int
	head   *node[T]
	tail   *node[T]
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		Length: 0,
	}
}

func (q *Queue[T]) Enqueue(value T) {
	q.Length++

	n := createNode(value, nil)
	if q.tail == nil || q.head == nil {
		q.tail = n
		q.head = n

		return
	}

	q.tail.next = n
	q.tail = n
}

func (q *Queue[T]) Deque() (T, error) {
	var out T
	if q.head == nil {
		return out, errors.New("Value not found")
	}
	q.Length--

	head := q.head
	q.head = q.head.next

	head.next = nil

	out = head.value

	return out, nil
}

func (q *Queue[T]) Peek() (T, error) {
	var out T
	if q.head == nil {
		return out, errors.New("There are no values in the queue yet")
	}

	out = q.head.value

	return out, nil
}

func createTail[T any](value T) *node[T] {
	return &node[T]{
		value: value,
		next:  nil,
	}
}

func createNode[T any](value T, next *node[T]) *node[T] {
	return &node[T]{
		value: value,
		next:  next,
	}
}
