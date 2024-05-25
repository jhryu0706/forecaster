package queue

import (
	"proj3/utils"
	"sync/atomic"
	"unsafe"
)

// LockfreeQueue represents a FIFO structure with operations to enqueue
// and dequeue tasks represented as Request
type node struct {
	value *utils.Task
	next  unsafe.Pointer
	prev  unsafe.Pointer
}

type LockFreeQueue struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	Count int32
}

// NewQueue creates and initializes a LockFreeQueue
func NewLockFreeQueue() *LockFreeQueue {
	dummy := unsafe.Pointer(&node{})
	return &LockFreeQueue{head: dummy, tail: dummy}
}

// Enqueue adds a series of Request to the queue
func (q *LockFreeQueue) Enqueue(task *utils.Task) {
	newNode := &node{value: task}
	newPtr := unsafe.Pointer(newNode)

	for {
		tail := (*node)(atomic.LoadPointer(&q.tail))
		next := (*node)(atomic.LoadPointer(&tail.next))

		if tail == (*node)(atomic.LoadPointer(&q.tail)) {
			if next == nil {
				if atomic.CompareAndSwapPointer(&tail.next, nil, newPtr) {
					newNode.prev = unsafe.Pointer(tail)
					atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), newPtr)
					atomic.AddInt32(&q.Count, 1)
					return
				}
			} else {
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			}
		}
	}
}

func (q *LockFreeQueue) PopFront() *utils.Task {
	for {
		head := (*node)(atomic.LoadPointer(&q.head))
		tail := (*node)(atomic.LoadPointer(&q.tail))
		next := (*node)(atomic.LoadPointer(&head.next))

		if head == (*node)(atomic.LoadPointer(&q.head)) {
			if head == tail {
				if next == nil {
					return nil
				}
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			} else {
				value := next.value
				next.value = nil
				if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
					atomic.AddInt32(&q.Count, -1)
					return value
				}
			}
		}
	}
}

func (q *LockFreeQueue) PopBack() *utils.Task {
	for {
		tail := (*node)(atomic.LoadPointer(&q.tail))
		head := (*node)(atomic.LoadPointer(&q.head))
		prev := (*node)(atomic.LoadPointer(&tail.prev))

		if tail == (*node)(atomic.LoadPointer(&q.tail)) {
			if head == tail {
				if prev == nil {
					return nil
				}
				atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(prev))
			} else {
				value := tail.value
				if atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(prev)) {
					atomic.AddInt32(&q.Count, -1)
					prev.next = nil
					return value
				}
			}
		}
	}
}
