package funnelqueue

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// Intrusive declares an interface for interface for an intrusive node.
type Intrusive interface {
	Next() *unsafe.Pointer
}

// IntrusiveNode implements Intrusive.
type IntrusiveNode struct {
	next unsafe.Pointer
}

// Next returns a pointer to the next object.
func (d *IntrusiveNode) Next() *unsafe.Pointer {
	return &d.next
}

// New creates a new Queue instance
func New() Queue {
	return Queue{}
}

// Queue is a FIFO, lock-free, concurrent, multi-producer/single-consumer,
// linked-list-based queue.
type Queue struct {
	back  unsafe.Pointer
	front unsafe.Pointer
}

// Push adds a value to the end of the queue.
//
// It's allowed to use Push concurrently from different goroutines.
func (q *Queue) Push(v interface{}) bool {
	return q.PushIntrusive(&entry{v: v})
}

// PushIntrusive adds a value to the end of the queue.
// No additional allocations performed in compare of Push.
//
// It's allowed to use Push concurrently from different goroutines.
func (q *Queue) PushIntrusive(v Intrusive) bool {
	new := unsafe.Pointer(&v)
	old := atomic.SwapPointer(&q.back, new)
	if old != nil {
		atomic.StorePointer((*(*Intrusive)(old)).Next(), new)
		return false
	}
	atomic.StorePointer(&q.front, new)
	return true
}

// Pop removes the top value from the queue and returns it as a result
// of the function call.
//
// Pop can not be used concurrently.
func (q *Queue) Pop() interface{} {
	front := atomic.LoadPointer(&q.front)
	if front == nil {
		return nil
	}
	// Is front the last element in the queue?
	if atomic.CompareAndSwapPointer(&q.back, front, nil) {
		// Clear q.front in case it was not changed.
		atomic.CompareAndSwapPointer(&q.front, front, nil)
	} else {
		for {
			next := atomic.LoadPointer((*(*Intrusive)(front)).Next())
			if next != nil {
				atomic.StorePointer(&q.front, next)
				break
			}
			// Wait the other goroutine to help us fixing the tail.
			runtime.Gosched()
		}
	}
	switch v := (*(*Intrusive)(front)).(type) {
	case *entry:
		return v.v
	default:
		return v
	}
}

// IsEmpty checks if the queue is empty.
//
// IsEmpty can not be used concurrently.
func (q *Queue) IsEmpty() bool {
	return atomic.LoadPointer(&q.front) != nil
}

type entry struct {
	IntrusiveNode
	v interface{}
}
