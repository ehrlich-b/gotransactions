package gotransactions

import (
	"sync"
)

type Node[T any] struct {
	data T
	next *Node[T]
}

type ConcurrentQueue[T any] struct {
	head *Node[T]
	tail *Node[T]
	lock sync.Mutex
	cond *sync.Cond
}

func NewConcurrentQueue[T any]() *ConcurrentQueue[T] {
	q := &ConcurrentQueue[T]{}
	q.cond = sync.NewCond(&q.lock)
	return q
}

func (q *ConcurrentQueue[T]) Enqueue(data T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	node := &Node[T]{data: data}

	if q.tail == nil {
		q.head = node
		q.tail = node
	} else {
		q.tail.next = node
		q.tail = node
	}
	q.cond.Broadcast()
}

func (q *ConcurrentQueue[T]) Dequeue() <-chan T {
    result := make(chan T, 1)

    go func() {
        q.lock.Lock()
        defer q.lock.Unlock()

        for q.head == nil {
            q.cond.Wait()
        }

        data := q.head.data
        q.head = q.head.next
        if q.head == nil {
            q.tail = nil
        }

        result <- data
    }()

    return result
}

func (q *ConcurrentQueue[T]) Peek() (T, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.head == nil {
		return zero[T](), false
	}

	return q.head.data, true
}

func (q *ConcurrentQueue[T]) Push(data T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	node := &Node[T]{data: data}

	if q.head == nil {
		q.head = node
		q.tail = node
	} else {
		node.next = q.head
		q.head = node
	}
	q.cond.Broadcast()
}

func (q *ConcurrentQueue[T]) Pop() (T, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.tail == nil {
		return zero[T](), false
	}

	data := q.tail.data

	if q.head == q.tail {
		q.head = nil
		q.tail = nil
	} else {
		p := q.head
		for p.next != q.tail {
			p = p.next
		}
		q.tail = p
		p.next = nil
	}
	return data, true
}

// helper function to get zero value of any type
func zero[T any]() T {
	var z T
	return z
}
